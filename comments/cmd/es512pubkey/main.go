package main

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("reading from stdin: %v", err)
	}
	for {
		block, rest := pem.Decode(data)
		if block.Type == "PRIVATE KEY" {
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err != nil {
				log.Fatalf("parsing private key: %v", err)
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
			return
		}
		data = rest
	}
}
