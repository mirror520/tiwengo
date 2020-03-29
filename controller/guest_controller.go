package controller

import (
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/model"
	"github.com/mirror520/tiwengo/util"

	log "github.com/sirupsen/logrus"
)

// LoginGuestUserHandler ...
func LoginGuestUserHandler(input *model.Guest) (*model.User, error) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "LoginGuestUser",
		"username":   input.Phone,
	})

	var db *gorm.DB = model.DB

	var user model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", input.Phone).First(&user)
	if (user.Username != input.Phone) || (user.Guest.PhoneToken != input.PhoneToken) {
		return nil, errors.New("訪客驗證資料錯誤")
	}

	if !user.Guest.PhoneVerify {
		return nil, errors.New("您的手機尚未驗證通過，請重新驗證")
	}
	logger.Infoln("訪客登入成功")

	return &user, nil
}

// ShowGuestUserQRCodeHandler ...
func ShowGuestUserQRCodeHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "ShowGuestUserQRCode",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	userID := ctx.Param("user_id")
	logger = logger.WithFields(log.Fields{"user_id": userID})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", userID).First(&user)

	img, err := getTodayGuestUserQRCode(user)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("無法取得今天的 QR Code")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	w := ctx.Writer
	png.Encode(w, img)
}

// RegisterGuestUserHandler ...
func RegisterGuestUserHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "RegisterGuestUser",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	var input model.Guest
	err := ctx.ShouldBind(&input)
	if (err != nil) || (input.Phone == "") {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"phone": input.Phone})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("username = ?", input.Phone).First(&user)
	if user.Username == input.Phone {
		result = model.NewFailureResult().SetLogger(logger)
		if user.Guest.PhoneVerify {
			result.AddInfo("您的電話已經驗證過了")
			result.SetData(user)
			ctx.AbortWithStatusJSON(http.StatusConflict, result)
		} else {
			result.AddInfo("您需要驗證電話")
			result.SetData(user)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, result)
		}
		return
	}
	user = input.User()
	db.Create(&user)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("您註冊成功了")
	result.SetData(&user)

	ctx.JSON(http.StatusOK, result)
}

// SendGuestPhoneOTPHandler ...
func SendGuestPhoneOTPHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "SendGuestPhoneOTP",
	})

	var db *gorm.DB = model.DB
	var redisClient *redis.Client = model.RedisClient
	var result *model.Result

	var input model.Guest
	err := ctx.ShouldBind(&input)
	if (err != nil) || (input.Phone == "") {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"user_id": input.UserID})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", input.UserID).First(&user)
	if user.Username != input.Phone {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料不正確")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	var guest model.Guest = user.Guest
	sms, err := util.NewSMS()
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("SMS 系統初始化發生錯誤")
		result.AddInfo(err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	otp, token := sms.SetOTP(&guest)
	logger.WithFields(log.Fields{
		"otp":   otp,
		"token": token,
	}).Infoln("取得 OTP 驗證碼")

	key := fmt.Sprintf("otp-%s", guest.Phone)
	redisResult := redisClient.SetNX(
		key,
		otp,
		1*time.Minute,
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

	guest.PhoneToken = token
	db.Save(&guest)

	if os.Getenv("GIN_MODE") == "release" {
		smsResult, err := sms.SendSMS()
		if err != nil {
			result = model.NewFailureResult().SetLogger(logger)
			result.AddInfo("寄送 SMS OTP 失敗")
			result.AddInfo(err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
			return
		}
		logger.WithFields(log.Fields{"credit": smsResult.Credit})
	}

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("SMS OTP 發送成功")
	result.SetData(guest)

	ctx.JSON(http.StatusOK, result)
}

// VerifyGuestPhoneOTPHandler ...
func VerifyGuestPhoneOTPHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "VerifyGuestPhoneOTP",
	})

	var db *gorm.DB = model.DB
	var redisClient *redis.Client = model.RedisClient
	var result *model.Result

	var input model.Guest
	err := ctx.ShouldBind(&input)
	if (err != nil) || (input.Phone == "") {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"user_id": input.UserID})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", input.UserID).First(&user)
	if user.Username != input.Phone {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料不正確")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	var guest model.Guest = user.Guest
	key := fmt.Sprintf("otp-%s", guest.Phone)
	otp, err := redisClient.Get(key).Result()
	if (err != nil) || (otp != input.PhoneOTP) || (guest.PhoneToken != input.PhoneToken) {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("驗證碼不正確或已經失效")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	guest.PhoneVerify = true
	db.Save(&guest)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("您的手機已通過驗證")
	result.SetData(guest)

	ctx.JSON(http.StatusOK, result)
}
