package tau

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	adr "github.com/iotaledger/iota.go/address"
	ioa "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/bundle"
	con "github.com/iotaledger/iota.go/converter"
	tran "github.com/iotaledger/iota.go/transaction"
)

type byTimeStampDesc tran.Transactions

func (a byTimeStampDesc) Len() int { return len(a) }
func (a byTimeStampDesc) Less(i, j int) bool {
	return a[i].AttachmentTimestamp > a[j].AttachmentTimestamp
}
func (a byTimeStampDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// CreateOrUpdateTangleID -
func CreateOrUpdateTangleID(api *ioa.API, seed string, optionalIDFields OptionalIDFields) (string, error) {
	pKey, err := ecdsa.GenerateKey(elliptic.P521(), newWrappingStringReader(seed))
	if err != nil {
		panic(err)
	}

	// Marshall the private key
	privateKeyString, err := x509.MarshalECPrivateKey(pKey)
	if err != nil {
		panic(err)
	}

	// Marshall the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&pKey.PublicKey)
	if err != nil {
		panic(err)
	}

	// Create and encode the pem block.
	// Pem blocks are the format used when using the keys
	privatePem := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyString,
	}

	privateBuf := new(bytes.Buffer)
	err = pem.Encode(privateBuf, &privatePem)
	if err != nil {
		panic(err)
	}
	publicPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicBuf := new(bytes.Buffer)
	err = pem.Encode(publicBuf, &publicPem)
	if err != nil {
		panic(err)
	}
	// Print out the private and public key
	fmt.Println(privateBuf.String())
	fmt.Println(publicBuf.String())

	// Hash the public key
	checksum := sha512.New()
	publicHash := checksum.Sum(publicKeyBytes)

	fmt.Println("---Public Hash---")
	fmt.Println(base64.StdEncoding.EncodeToString(publicHash[:]))
	r, s, err := ecdsa.Sign(newWrappingStringReader(seed), pKey, publicHash)
	if err != nil {
		panic(err)
	}
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	signatureString := base64.StdEncoding.EncodeToString(signature[:])
	fmt.Println("---Signature Length---")
	fmt.Println(strconv.Itoa(len(signature)))
	fmt.Println(base64.StdEncoding.EncodeToString(signature[:]))

	reference, err := GenerateReferenceFromSignature(signatureString)
	if err != nil {
		panic(err)
	}
	fmt.Println("---Reference address---")
	fmt.Println(reference)

	msg := TangleID{
		OptionalIDFields: optionalIDFields,
		PublicKey:        publicBuf.String(),
		Signature:        signatureString,
	}
	_ = msg

	jsonMsg, err := json.Marshal(msg)
	if nil != err {
		panic(err)
	}

	jsonMsgString := base64.StdEncoding.EncodeToString(jsonMsg[:])

	tryteMsg, err := con.ASCIIToTrytes(jsonMsgString)
	if err != nil {
		panic(err)
	}

	fmt.Println("---Message---")
	fmt.Println(tryteMsg)

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
		panic(err)
	}

	txs := tran.Transactions{}
	for i := range bundleEntries {
		txs = bundle.AddEntry(txs, bundleEntries[i])
	}

	finalizedBundle, err := bundle.Finalize(txs)
	if err != nil {
		panic(err)
	}

	finishedBundle, err := api.SendTrytes(tran.MustFinalTransactionTrytes(finalizedBundle), 3, 14)
	if err != nil {
		panic(err)
	}
	fmt.Println("---Tail transaction hash---")
	fmt.Println(bundle.TailTransactionHash(finishedBundle))

	return reference, nil
}

// GenerateReference generates a new reference from a seed
func GenerateReference(seed string) (string, error) {
	pKey, err := ecdsa.GenerateKey(elliptic.P521(), newWrappingStringReader(seed))
	if err != nil {
		panic(err)
	}

	// Marshall the private key
	privateKeyString, err := x509.MarshalECPrivateKey(pKey)
	if err != nil {
		panic(err)
	}

	// Marshall the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&pKey.PublicKey)
	if err != nil {
		panic(err)
	}

	// Create and encode the pem block.
	// Pem blocks are the format used when using the keys
	privatePem := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyString,
	}

	privateBuf := new(bytes.Buffer)
	err = pem.Encode(privateBuf, &privatePem)
	if err != nil {
		panic(err)
	}
	publicPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicBuf := new(bytes.Buffer)
	err = pem.Encode(publicBuf, &publicPem)
	if err != nil {
		panic(err)
	}
	// Print out the private and public key
	fmt.Println(privateBuf.String())
	fmt.Println(publicBuf.String())

	// Hash the public key
	checksum := sha512.New()
	publicHash := checksum.Sum(publicKeyBytes)

	fmt.Println("---Public Hash---")
	fmt.Println(base64.StdEncoding.EncodeToString(publicHash[:]))
	r, s, err := ecdsa.Sign(newWrappingStringReader(seed), pKey, publicHash)
	if err != nil {
		panic(err)
	}
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	signatureString := base64.StdEncoding.EncodeToString(signature[:])
	fmt.Println("---Signature Length---")
	fmt.Println(strconv.Itoa(len(signature)))
	fmt.Println(base64.StdEncoding.EncodeToString(signature[:]))

	reference, err := GenerateReferenceFromSignature(signatureString)
	if err != nil {
		panic(err)
	}
	fmt.Println("---Reference address---")
	fmt.Println(reference)

	return reference, nil
}

// SignRequest populates the request Authorization header with a JWT
// This JWT has the sub set to your reference on the tangle and the aud set to the request url
func SignRequest(seed string, r *http.Request) error {
	if seed == "" {
		return ErrEmptySeed
	}
	key, err := GenerateKey(seed)
	if err != nil {
		return err
	}
	privatePem, err := GeneratePrivatePem(key)
	if err != nil {
		return err
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privatePem)
	if err != nil {
		return err
	}

	reference, err := GenerateReference(seed)
	if err != nil {
		return err
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

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return err
	}

	r.Header.Set("Authorization", "Bearer "+tokenString)

	return nil
}

// VerifyRequest -
func VerifyRequest(api *ioa.API, r *http.Request) (*TangleID, error) {
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
		return nil, err
	}

	if claims.Subject == "" {
		return nil, ErrNotVerified
	}

	if claims.Audience == "" {
		return nil, ErrNotVerified
	}

	if claims.Audience != url {
		return nil, ErrNotVerified
	}

	// If the subject is not a valid iota address we know for sure that the token is invalid
	if adr.ValidAddress(claims.Subject) != nil {
		return nil, ErrNotVerified
	}

	hashes, err := api.FindTransactions(ioa.FindTransactionsQuery{
		Addresses: []string{claims.Subject},
	})
	if err != nil {
		return nil, err
	}

	transactions, err := api.GetTransactionObjects(hashes...)
	if err != nil {
		return nil, err
	}
	sort.Sort(byTimeStampDesc(transactions))
	for _, v := range transactions {
		fmt.Println("transaction hash")
		fmt.Println(v.Hash)
		signatureMessageFragment := strings.TrimRight(v.SignatureMessageFragment, "9")
		msg, err := con.TrytesToASCII(signatureMessageFragment)
		if err != nil {
			panic(err)
		}
		msgJSON, err := base64.StdEncoding.DecodeString(msg)
		if err != nil {
			panic(err)
		}
		tangleID := TangleID{}
		if err := json.Unmarshal(msgJSON, &tangleID); err != nil {
			panic(err)
		}

		publicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(tangleID.PublicKey))
		if err != nil {
			panic(err)
		}

		token, err := jwt.Parse(tokenRaw, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})
		if err != nil {
			panic(err)
		}

		if !token.Valid {
			panic(err)
		}

		signature, err := base64.StdEncoding.DecodeString(tangleID.Signature)
		if err != nil {
			panic(err)
		}
		signatureLength := len(signature) / 2
		r := new(big.Int).SetBytes(signature[:signatureLength])
		s := new(big.Int).SetBytes(signature[signatureLength:])

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			panic(err)
		}
		checksum := sha512.New()
		publicHash := checksum.Sum(publicKeyBytes)
		if !ecdsa.Verify(publicKey, publicHash, r, s) {
			fmt.Println("signature couldn not be verified")
			continue
		}

		reference, err := GenerateReferenceFromSignature(tangleID.Signature)
		if err != nil {
			panic(err)
		}

		if reference != claims.Subject {
			fmt.Println("reference did not match subject")
			continue
		}

		return &tangleID, nil
	}

	// We couldn't find a valid transaction on the tangle
	return nil, ErrNotVerified
}
