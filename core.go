package tau

import (
	"github.com/pkg/errors"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"io"

	ics "github.com/iotaledger/iota.go/checksum"
	con "github.com/iotaledger/iota.go/converter"
	curl "github.com/iotaledger/iota.go/curl"
)

// GenerateKey -
func GenerateKey(seed string) (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), newWrappingStringReader(seed))
	return key, err
}

// GeneratePrivatePem -
func GeneratePrivatePem(key *ecdsa.PrivateKey) ([]byte, error) {
	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal key failed")
	}

	privatePem := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	}

	buffer := new(bytes.Buffer)
	err = pem.Encode(buffer, &privatePem)
	if err != nil {
		return nil, errors.Wrap(err, "could not encode the key")
	}

	return buffer.Bytes(), nil
}

// GeneratePublicPem -
func GeneratePublicPem(key *ecdsa.PrivateKey) ([]byte, error) {
	publicKey, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "marshal key failed")
	}

	publicPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	}
	buffer := new(bytes.Buffer)
	err = pem.Encode(buffer, &publicPem)
	if err != nil {
		return nil, errors.Wrap(err, "could not encode the key")
	}

	return buffer.Bytes(), nil
}

// GenerateReferenceFromSignature -
func GenerateReferenceFromSignature(signature string) (string, error) {
	trytes, err := con.ASCIIToTrytes(signature)
	if err != nil {
		return "", errors.Wrap(err, "convert the signature to trytes failed")
	}
	reference, err := curl.HashTrytes(trytes, 81)
	if err != nil {
		return "", errors.Wrap(err, "hashing trytes failed")
	}

	referenceWithChecksum, err := ics.AddChecksum(reference[:81], true, 9)

	return referenceWithChecksum, err
}

// Sign -
func Sign(rand io.Reader, key *ecdsa.PrivateKey, value []byte) ([]byte, error) {
	checksum := sha512.New()
	publicHash := checksum.Sum(value)
	r, s, err := ecdsa.Sign(rand, key, publicHash)
	if err != nil {
		return nil, errors.Wrap(err, "signing hash failed")
	}
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	return signature, nil
}
