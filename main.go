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
	"github.com/mirror520/tiwengo/model"
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

func createAndUpdatePrivkey(w io.Writer, dateKey string) error {
	privkey, _ := generateKeyPair(512)
	privkeyPem := new(bytes.Buffer)

	encodePrivateKeyPem(privkeyPem, privkey)

	happyColor := colorful.HappyColor()
	message := []byte(happyColor.Hex())
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privkey.PublicKey, message)
	if err != nil {
		fmt.Fprintf(w, "加密時發生錯誤: %s\n", err.Error())
		return err
	}

	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Fprintf(w, "Base64封裝後密文:\n %s\n", encodedCiphertext)

	err = redisClient.HSet(dateKey, map[string]interface{}{
		"privkey":    privkeyPem.String(),
		"ciphertext": encodedCiphertext,
	}).Err()
	if err != nil {
		fmt.Fprintf(w, "將新的私鑰加入Redis資料庫時發生錯誤: %s\n", err.Error())
		return err
	}

	return nil
}

func createPrivkeyHandler(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	dateKey := fmt.Sprintf("date-%s", date)

	result, _ := redisClient.Exists(dateKey).Result()
	if result == 1 {
		fmt.Fprintf(w, "Date: %s 的私鑰已經產生過了", dateKey)
		return
	}

	if err := createAndUpdatePrivkey(w, dateKey); err != nil {
		fmt.Fprintf(w, "新增金鑰失敗: %s\n", err.Error())
		return
	}
	fmt.Fprintf(w, "成功產生%s的私鑰，並成功加入至Redis資料庫", date)
}

func updatePrivkeyHandler(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	dateKey := fmt.Sprintf("date-%s", date)

	result, _ := redisClient.Exists(dateKey).Result()
	if result == 0 {
		fmt.Fprintf(w, "Date: %s 的私鑰不存在", dateKey)
		return
	}
	if err := createAndUpdatePrivkey(w, dateKey); err != nil {
		fmt.Fprintf(w, "更新金鑰失敗: %s\n", err.Error())
		return
	}
	fmt.Fprintf(w, "成功更新%s的私鑰，及更新至Redis資料庫", date)
}

func showPrivkeyQrCodeHandler(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	dateKey := fmt.Sprintf("date-%s", date)

	privkey, err := redisClient.HGet(dateKey, "privkey").Result()
	if err != nil {
		fmt.Fprintf(w, "無法取得私鑰: %s\n", err.Error())
		log.Fatalln(err.Error())
	}

	qrCode, err := qrcode.Encode(privkey, qrcode.Medium, 600)
	if err != nil {
		fmt.Fprintf(w, "無法產生QR Code: %s", err)
	}
	img, _, _ := image.Decode(bytes.NewBuffer(qrCode))

	png.Encode(w, img)
}

func showPrivkeyCiphertextHandler(w http.ResponseWriter, r *http.Request) {
	var result *model.Result
	w.Header().Set("Content-Type", "application/json")

	date := mux.Vars(r)["date"]
	dateKey := fmt.Sprintf("date-%s", date)

	ciphertext, err := redisClient.HGet(dateKey, "ciphertext").Result()
	if err != nil {
		result = model.NewFailureResult()
		result.AddInfo(fmt.Sprintf("無法取得 %s 密文", date))

		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		result = model.NewSuccessResult()
		result.AddInfo(fmt.Sprintf("成功取得 %s 密文", date))
		result.SetData(ciphertext)

		w.WriteHeader(http.StatusOK)
	}

	w.Write(result.JSON())
}

func indexPrivkeysHandler(w http.ResponseWriter, r *http.Request) {
	keys, _ := redisClient.Keys("date-*").Result()
	for _, key := range keys {
		val, _ := redisClient.HGet(key, "privkey").Result()
		fmt.Fprintf(w, "Key: %s, Value: \n%s\n", key, val)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	router := mux.NewRouter()
	router.HandleFunc("/privkeys", indexPrivkeysHandler).Methods("GET")
	router.HandleFunc("/privkeys/{date}", createPrivkeyHandler).Methods("POST")
	router.HandleFunc("/privkeys/{date}", updatePrivkeyHandler).Methods("PUT", "PATCH")
	router.HandleFunc("/privkeys/{date}/qr", showPrivkeyQrCodeHandler).Methods("GET")
	router.HandleFunc("/privkeys/{date}/ciphertext", showPrivkeyCiphertextHandler).Methods("GET")
	router.Use(loggingMiddleware)
	log.Fatal(http.ListenAndServe(":6080", cors.Default().Handler(router)))
}
