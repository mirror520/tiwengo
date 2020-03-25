package controller

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mirror520/tiwengo/model"
	"github.com/skip2/go-qrcode"

	log "github.com/sirupsen/logrus"
)

// GetPrivkeyHandler ...
func GetPrivkeyHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Rsa",
		"event":      "GetPrivkeyHandler",
	})

	var result *model.Result

	dateStr := time.Now().Format("20060102")
	dateKey := fmt.Sprintf("date-%s", dateStr)

	privkeyPem, err := getPrivkeyPem(dateKey)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("無法取得私鑰")
		result.AddInfo(err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("成功取得今天的私鑰")
	result.SetData(privkeyPem)

	ctx.JSON(http.StatusOK, result)
}

func generateKeyPair(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
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

func parsePemToPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	privPem, _ := pem.Decode([]byte(pemStr))
	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("RSA 私鑰是錯誤的型態")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func parsePemToPublicKey(pemStr string) (*rsa.PublicKey, error) {
	pubPem, _ := pem.Decode([]byte(pemStr))
	if pubPem.Type != "PUBLIC KEY" {
		return nil, errors.New("RSA 公鑰是錯誤的型態")
	}

	pubkey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		return nil, err
	}

	return pubkey.(*rsa.PublicKey), nil
}

func createPrivkey(dateKey string) (string, error) {
	redisClient := model.RedisClient

	privkey, _ := generateKeyPair(2048)
	privkeyPem := new(bytes.Buffer)

	encodePrivateKeyPem(privkeyPem, privkey)

	happyColor := colorful.HappyColor()
	message := []byte(happyColor.Hex())
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privkey.PublicKey, message)
	if err != nil {
		return "", err
	}

	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)
	err = redisClient.HSet(dateKey, map[string]interface{}{
		"privkey":    privkeyPem.String(),
		"ciphertext": encodedCiphertext,
	}).Err()
	if err != nil {
		return "", err
	}
	redisClient.Expire(dateKey, 24*time.Hour)

	return privkeyPem.String(), nil
}

func getPrivkeyPem(dateKey string) (string, error) {
	redisClient := model.RedisClient
	privkeyPem, err := redisClient.HGet(dateKey, "privkey").Result()
	if err != nil {
		privkeyPem, err = createPrivkey(dateKey)
		if err != nil {
			return "", nil
		}
	}

	return privkeyPem, nil
}

func getPrivkey(dateKey string) (*rsa.PrivateKey, error) {
	privkeyPem, err := getPrivkeyPem(dateKey)
	if err != nil {
		return nil, err
	}

	return parsePemToPrivateKey(privkeyPem)
}

func getTodayGuestUserQRCode(user model.User) (image.Image, error) {
	dateStr := time.Now().Format("20060102")
	dateKey := fmt.Sprintf("date-%s", dateStr)

	privkey, err := getPrivkey(dateKey)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("%d,%s", user.ID, user.Username)
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privkey.PublicKey, []byte(message))
	if err != nil {
		return nil, err
	}

	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	qrCode, err := qrcode.Encode(encodedCiphertext, qrcode.Medium, 600)
	if err != nil {
		return nil, err
	}
	img, _, _ := image.Decode(bytes.NewBuffer(qrCode))

	return img, err
}
