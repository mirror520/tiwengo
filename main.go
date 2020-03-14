package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Fatal(err)
	}

	return privkey, &privkey.PublicKey
}

func encodePrivateKeyPem(out io.Writer, key *rsa.PrivateKey) {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	pem.Encode(out, block)
}

func encodePublicKeyPem(out io.Writer, key *rsa.PublicKey) {
	pubKeyByte, _ := x509.MarshalPKIXPublicKey(key)

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyByte,
	}

	pem.Encode(out, block)
}

func parsePemToPrivateKey(filename string) *rsa.PrivateKey {
	priv, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("Unable to open private key pem")
	}

	privPem, _ := pem.Decode(priv)
	if privPem.Type != "RSA PRIVATE KEY" {
		log.Fatalln("RSA private key is of the wrong type")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		log.Fatalln("Unable to parse RSA private key")
	}

	return privKey
}

func parsePemToPublicKey(filename string) *rsa.PublicKey {
	pub, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err.Error())
	}

	pubPem, _ := pem.Decode(pub)
	if pubPem.Type != "PUBLIC KEY" {
		log.Fatalln("RSA public key is of the wrong type")
	}

	pubkey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		log.Fatalln("Unable to parse RSA public key")
	}

	return pubkey.(*rsa.PublicKey)
}

func main() {
	privkey, _ := generateKeyPair(1024)

	// encodePublicKeyPem(os.Stdout, &privkey.PublicKey)
	// encodePrivateKeyPem(os.Stdout, privkey)

	pubkeyPEM, _ := os.Create("pubkey.pem")
	encodePublicKeyPem(pubkeyPEM, &privkey.PublicKey)

	privkeyPEM, _ := os.Create("privkey.pem")
	encodePrivateKeyPem(privkeyPEM, privkey)

	// pubkey := parsePemToPublicKey("pubkey.pem")
	// privkey := parsePemToPrivateKey("privkey.pem")

	message := []byte("#333333")

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privkey.PublicKey, message)
	if err != nil {
		log.Fatalf("Error from encryption: %s\n", err)
		return
	}

	fmt.Printf("Ciphertext: %x\n", ciphertext)

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privkey, ciphertext)
	if err != nil {
		log.Fatalf("Error from decryption: %s\n", err)
		return
	}

	fmt.Printf("Plaintext: %s\n", string(plaintext))
}
