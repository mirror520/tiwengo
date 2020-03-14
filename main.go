package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/skip2/go-qrcode"
)

var redisClient *redis.Client

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

func viewQrCodeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := redisClient.Get("date-20200314").Result()
	if err != nil {
		log.Fatalln(err.Error())
	}

	qrCode, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		fmt.Fprintf(w, "Unable to generate qr code: %s", err)
	}
	img, _, _ := image.Decode(bytes.NewBuffer(qrCode))

	png.Encode(w, img)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	err := redisClient.Set("date-20200314", "#333333", 0).Err()
	if err != nil {
		log.Fatalln(err.Error())
	}

	router := mux.NewRouter()
	router.HandleFunc("/qr", viewQrCodeHandler)
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(":6080", cors.Default().Handler(router)))
}
