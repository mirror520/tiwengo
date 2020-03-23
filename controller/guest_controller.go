package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"
	"github.com/mirror520/tiwengo/util"

	log "github.com/sirupsen/logrus"
)

// RegisterGuestUserHandler ...
func RegisterGuestUserHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "GuestController",
		"event":      "RegisterGuestUserEvent",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	var input model.Guest
	err := ctx.ShouldBind(&input)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("訪客輸入資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"phone": input.Phone})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", input.Phone).First(&user)
	if user.Username == input.Phone {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您已經註冊過了")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	user = input.User()
	db.Create(&user)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("訪客註冊成功")
	result.SetData(&user)

	ctx.JSON(http.StatusOK, result)
}

// SendGuestPhoneOTPHandler ...
func SendGuestPhoneOTPHandler(ctx *gin.Context) {
	redisClient := model.RedisClient
	logger := log.WithFields(log.Fields{
		"controller": "GuestController",
		"event":      "SendGuestPhoneOTPHandler",
	})

	// var db *gorm.DB = model.DB
	var result *model.Result

	var input model.Guest
	err := ctx.ShouldBind(&input)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("訪客輸入資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"phone": input.Phone})

	sms, err := util.NewSMS()
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("SMS 系統初始化發生錯誤")
		result.AddInfo(err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	otp := sms.SetOTP(&input)
	logger.WithFields(log.Fields{"otp": otp}).Infoln("取得 OTP 驗證碼")

	key := fmt.Sprintf("otp-%s", input.Phone)
	redisResult := redisClient.SetNX(
		key,
		otp,
		5*time.Minute,
	)
	if !redisResult.Val() {
		ttl := redisClient.TTL(key)
		msg := fmt.Sprintf("SMS 請求太頻繁，請於 %s 後再嘗試", ttl.Val().String())

		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo(msg)
		result.SetData(ttl.Val().String())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger.Infoln("SMS OTP 已加入 Redis")

	smsResult, err := sms.SendSMS()
	logger.WithFields(log.Fields{"credit": smsResult.Credit})

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("SMS OTP 發送成功")

	ctx.JSON(http.StatusOK, result)
}

// VerifyGuestPhoneOTPHandler ...
func VerifyGuestPhoneOTPHandler(ctx *gin.Context) {

}
