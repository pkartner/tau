package tau

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/dgrijalva/jwt-go"
	adr "github.com/iotaledger/iota.go/address"
	ioa "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	con "github.com/iotaledger/iota.go/converter"
	tran "github.com/iotaledger/iota.go/transaction"
	tri "github.com/iotaledger/iota.go/trinary"
)

type byTimeStampDesc tran.Transactions

func (a byTimeStampDesc) Len() int { return len(a) }
func (a byTimeStampDesc) Less(i, j int) bool {
	return a[i].AttachmentTimestamp > a[j].AttachmentTimestamp
}
func (a byTimeStampDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// CreateOrUpdateTangleID -
func CreateOrUpdateTangleID(api *ioa.API, seed string, optionalIDFields OptionalIDFields) (string, error) {
	if seed == "" {
		return "", ErrEmptySeed
	}

	if api == nil {
		return "", ErrInvalidIOTAAPI
	}

	key, err := GenerateKey(seed)
	if err != nil {
		return "", errors.Wrap(err, "generate key failed")
	}

	publicPem, err := GeneratePublicPem(key)
	if err != nil {
		return "", errors.Wrap(err, "generate public pem failed")
	}

	// Hash the public key
	signature, err := Sign(newWrappingStringReader(seed), key, publicPem)
	if err != nil {
		return "", errors.Wrap(err, "signing the public pem failed")
	}
	signatureString := base64.StdEncoding.EncodeToString(signature[:])

	reference, err := GenerateReferenceFromSignature(signatureString)
	if err != nil {
		return "", errors.Wrap(err, "could not generate a reference from the signature")
	}

	encodedPublicPem := base64.StdEncoding.EncodeToString(publicPem)
	msg := TangleID{
		OptionalIDFields: optionalIDFields,
		PublicKey:        encodedPublicPem,
		Signature:        signatureString,
	}

	jsonMsg, err := json.Marshal(msg)
	if nil != err {
		return "", errors.Wrap(err, "marshall message to json failed")
	}

	jsonMsgString := base64.StdEncoding.EncodeToString(jsonMsg[:])

	tryteMsg, err := con.ASCIIToTrytes(jsonMsgString)
	if err != nil {
		return "", errors.Wrap(err, "converting json string to trytes failed")
	}
	transfers := bundle.Transfers{
		{
			Address: reference,
			Value:   0,
			Tag:     "TANGLEAUTH",
			Message: tryteMsg,
		},
	}

	timestamp := uint64(time.Now().UnixNano() / int64(time.Second))
	bundleEntries, err := bundle.TransfersToBundleEntries(timestamp, transfers...)
	if err != nil {
		return "", errors.Wrap(err, "converting transactions to bundle entries failed")
	}

	txs := tran.Transactions{}
	for i := range bundleEntries {
		txs = bundle.AddEntry(txs, bundleEntries[i])
	}

	finalizedBundle, err := bundle.Finalize(txs)
	if err != nil {
		return "", errors.Wrap(err, "could not finalize bundle")
	}

	_, err = api.SendTrytes(tran.MustFinalTransactionTrytes(finalizedBundle), 3, 14)
	if err != nil {
		return "", ErrCallingTangle
	}

	return reference, nil
}

// GenerateReference generates a new reference from a seed
func GenerateReference(seed string) (string, error) {
	if seed == "" {
		return "", ErrEmptySeed
	}

	key, err := GenerateKey(seed)
	if err != nil {
		return "", errors.Wrap(err, "generate key failed")
	}

	publicPem, err := GeneratePublicPem(key)
	if err != nil {
		return "", errors.Wrap(err, "generate public pem failed")
	}

	// Hash the public key
	signature, err := Sign(newWrappingStringReader(seed), key, publicPem)
	if err != nil {
		return "", errors.Wrap(err, "signing the public pem failed")
	}
	signatureString := base64.StdEncoding.EncodeToString(signature[:])

	reference, err := GenerateReferenceFromSignature(signatureString)
	if err != nil {
		return "", errors.Wrap(err, "could not generate a reference from the signature")
	}

	return reference, nil
}

// SignRequest populates the request Authorization header with a JWT
// This JWT has the sub set to your reference on the tangle and the aud set to the request url
func SignRequest(seed string, r *http.Request) error {
	if r == nil {
		return ErrNilRequest
	}
	if seed == "" {
		return ErrEmptySeed
	}
	key, err := GenerateKey(seed)
	if err != nil {
		return errors.Wrap(err, "could not generate key from seed")
	}
	privatePem, err := GeneratePrivatePem(key)
	if err != nil {
		return errors.Wrap(err, "could not generate private pem")
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privatePem)
	if err != nil {
		return errors.Wrap(err, "could not parse private pem")
	}

	reference, err := GenerateReference(seed)
	if err != nil {
		return errors.Wrap(err, "could not generate a reference from the seed")
	}

	url := r.Host + r.URL.Path
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := jwt.StandardClaims{
		// Subject is the same as the reference address on the tangle
		Subject: reference,
		// Audience is the url for which we make the request, this makes sure that our signed jwt cannot be used for any other purpose
		Audience:  url,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return errors.Wrap(err, "could not sign the jwt")
	}

	r.Header.Set("Authorization", "Bearer "+tokenString)

	return nil
}

// VerifyRequest verifies the request based on the authorization header with the tangle.
// If the person is authorized this function returns a TangleID else it return nil
func VerifyRequest(api *ioa.API, r *http.Request) (*TangleID, error) {
	if api == nil {
		return nil, ErrInvalidIOTAAPI
	}
	url := r.Host + r.URL.Path
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrEmptyAuthorization
	}
	tokenRaw := strings.Replace(authHeader, "Bearer ", "", 1)

	claims := &jwt.StandardClaims{}
	// First we parse the token without verifying it so we can get the claims out.
	_, err := jwt.ParseWithClaims(tokenRaw, claims, nil)
	// If the token is malformed we cannot use it if it is a verification error we can just move on
	if err.(*jwt.ValidationError).Errors&jwt.ValidationErrorMalformed != 0 {
		return nil, nil
	}

	if claims.Subject == "" {
		return nil, nil
	}

	if claims.Audience == "" {
		return nil, nil
	}

	if claims.Audience != url {
		return nil, nil
	}

	// If the subject is not a valid iota address we know for sure that the token is invalid
	if adr.ValidAddress(claims.Subject) != nil {
		return nil, nil
	}

	hashes, err := api.FindTransactions(ioa.FindTransactionsQuery{
		Addresses: []string{claims.Subject},
	})
	if err != nil {
		// We checked the address so the only way this fails should be if we can't reach the node
		return nil, ErrCallingTangle
	}
	// If we didn't get any hashes it means there are not transactions so we cannot authorize
	if len(hashes) == 0 {
		return nil, nil
	}

	transactions, err := api.GetTransactionObjects(hashes...)
	if err != nil {
		// We got the hashes directly from the tangle so this only fails if we can't reach the node
		return nil, ErrCallingTangle
	}
	sort.Sort(byTimeStampDesc(transactions))
	for _, v := range transactions {
		signatureMessageFragment := strings.TrimRight(v.SignatureMessageFragment, "9")

		// If any of the following operations fail that means the data in the transaction wasn't valid.
		// The transaction cannot be considered for authentication so we continue to the next one.
		if err := tri.ValidTrytes(signatureMessageFragment); err != nil {
			continue
		}
		msg, err := con.TrytesToASCII(signatureMessageFragment)
		if err != nil {
			continue
		}
		msgJSON, err := base64.StdEncoding.DecodeString(msg)
		if err != nil {
			continue
		}
		tangleID := TangleID{}
		if err := json.Unmarshal(msgJSON, &tangleID); err != nil {
			continue
		}
		decodedPublicKeyData, err := base64.StdEncoding.DecodeString(tangleID.PublicKey)
		if err != nil {
			continue
		}
		publicKey, err := jwt.ParseECPublicKeyFromPEM(decodedPublicKeyData)
		if err != nil {
			continue
		}

		// Verify the jwt for validty
		token, err := jwt.Parse(tokenRaw, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil || !token.Valid {
			continue
		}

		signature, err := base64.StdEncoding.DecodeString(tangleID.Signature)
		if err != nil {
			continue
		}

		if !Verify(decodedPublicKeyData, signature, publicKey) {
			continue
		}

		reference, err := GenerateReferenceFromSignature(tangleID.Signature)
		if err != nil || reference != claims.Subject {
			continue
		}

		return &tangleID, nil
	}

	// We couldn't find a valid transaction on the tangle
	return nil, nil
}
