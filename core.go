package tau

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"

	ics "github.com/iotaledger/iota.go/checksum"
	con "github.com/iotaledger/iota.go/converter"
	curl "github.com/iotaledger/iota.go/curl"
)

// GenerateKey-
func GenerateKey(seed string) (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), newWrappingStringReader(seed))
	if err != nil {
		return nil, err
	}
	return key, err
}

// GeneratePrivatePem -
func GeneratePrivatePem(key *ecdsa.PrivateKey) ([]byte, error) {
	privateKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	privatePem := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	}

	buffer := new(bytes.Buffer)
	err = pem.Encode(buffer, &privatePem)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// GeneratePublicPem -
func GeneratePublicPem(key *ecdsa.PrivateKey) ([]byte, error) {
	publicKey, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	publicPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	}
	buffer := new(bytes.Buffer)
	err = pem.Encode(buffer, &publicPem)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// GenerateReferenceFromSignature -
func GenerateReferenceFromSignature(signature string) (string, error) {
	trytes, err := con.ASCIIToTrytes(signature)
	if err != nil {
		return "", err
	}
	reference, err := curl.HashTrytes(trytes, 81)
	if err != nil {
		return "", err
	}

	referenceWithChecksum, err := ics.AddChecksum(reference[:81], true, 9)

	return referenceWithChecksum, err
}
