package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
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
	"github.com/lucasb-eyer/go-colorful"
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
		log.Fatalln("無法開啟私鑰PEM檔")
	}

	privPem, _ := pem.Decode(priv)
	if privPem.Type != "RSA PRIVATE KEY" {
		log.Fatalln("RSA私鑰是錯誤的型態")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		log.Fatalln("無法剖析RSA私鑰")
	}

	return privKey
}

func parsePemToPublicKey(filename string) *rsa.PublicKey {
	pub, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("無法開啟公鑰PEM檔")
	}

	pubPem, _ := pem.Decode(pub)
	if pubPem.Type != "PUBLIC KEY" {
		log.Fatalln("RSA公鑰是錯誤的型態")
	}

	pubkey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		log.Fatalln("無法剖析RSA公鑰")
	}

	return pubkey.(*rsa.PublicKey)
}

func createPrivkeyHandler(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]

	privkey, _ := generateKeyPair(512)
	privkeyPem := new(bytes.Buffer)

	encodePrivateKeyPem(privkeyPem, privkey)

	happyColor := colorful.HappyColor()
	message := []byte(happyColor.Hex())
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privkey.PublicKey, message)
	if err != nil {
		log.Fatalf("加密時發生錯誤: %s\n", err)
		return
	}

	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Fprintf(w, "Base64封裝後密文: %s\n", encodedCiphertext)

	err = redisClient.HSet("date-"+date, map[string]interface{}{
		"privkey":    privkeyPem.String(),
		"ciphertext": encodedCiphertext,
	}).Err()
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Fprintf(w, "成功加入%s的私鑰至Redis資料庫", date)
}

func showPrivkeyQrCodeHandler(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	dateKey := fmt.Sprintf("date-%s", date)

	content, err := redisClient.HGet(dateKey, "privkey").Result()
	if err != nil {
		log.Fatalln(err.Error())
	}

	qrCode, err := qrcode.Encode(content, qrcode.Medium, 600)
	if err != nil {
		fmt.Fprintf(w, "無法產生QR Code: %s", err)
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

	router := mux.NewRouter()
	router.HandleFunc("/privkey/{date}", createPrivkeyHandler).Methods("POST")
	router.HandleFunc("/privkey/{date}/qr", showPrivkeyQrCodeHandler).Methods("GET")
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(":6080", cors.Default().Handler(router)))
}
