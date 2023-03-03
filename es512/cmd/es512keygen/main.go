package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

func main() {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatalf("generating ecdsa key: %v", err)
	}

	data, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatalf("x509 marshaling ecdsa private key: %v", err)
	}

	if err := pem.Encode(
		os.Stdout,
		&pem.Block{Type: "PRIVATE KEY", Bytes: data},
	); err != nil {
		log.Fatalf("encoding private key as pem: %v", err)
	}

	data, err = x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Fatalf("x509 marshaling ecdsa public key: %v", err)
	}

	if err := pem.Encode(
		os.Stdout,
		&pem.Block{Type: "PUBLIC KEY", Bytes: data},
	); err != nil {
		log.Fatalf("encoding public key as pem: %v", err)
	}
}
