package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/model"
	"golang.org/x/crypto/pbkdf2"

	log "github.com/sirupsen/logrus"
)

func RefreshMQTTUserTokenHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "MQTT",
		"event":      "RefreshMQTTUserToken",
	})

	redisClient := model.RedisClient
	redisCtx := model.RedisContext

	u, _ := ctx.Get("username")
	username := fmt.Sprintf("mqtt:%v", u)

	t, _ := ctx.Get("JWT_TOKEN")
	token := fmt.Sprintf("%v", t)

	redisClient.SetEX(redisCtx, username, hashPassword(token), 24*time.Hour)

	result := model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("成功刷新 MQTT 使用者")

	ctx.JSON(http.StatusOK, result)
}

func hashPassword(password string) string {
	iter := 901
	saltLen := 12
	keyLen := 24

	salt := make([]byte, saltLen)
	rand.Read(salt)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	dk := pbkdf2.Key([]byte(password), []byte(encodedSalt), iter, keyLen, sha256.New)
	encodedKey := base64.StdEncoding.EncodeToString(dk)

	return fmt.Sprintf("PBKDF2$%s$%d$%s$%s", "sha256", iter, encodedSalt, encodedKey)
}
