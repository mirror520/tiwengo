package controller

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
	})

	var db *gorm.DB = model.DB

	var guest model.Guest
	db.Where("phone_token = ?", input.PhoneToken).First(&guest)

	if guest.UserID == 0 {
		return nil, errors.New("訪客驗證資料錯誤")
	}

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", guest.UserID).First(&user)

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
	followers := ctx.Query("followers")
	logger = logger.WithFields(log.Fields{"user_id": userID})

	if followers != "" {
		logger = logger.WithFields(log.Fields{"followers": followers})
	}

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", userID).First(&user)

	img, err := getTodayGuestUserQRCode(user, followers)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("無法取得今天的 QR Code")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	w := ctx.Writer
	png.Encode(w, img)

	logger.Info("成功產製 QR Code")
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
	if (err != nil) || ((input.Phone == "") && (input.IDCard == "")) {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}

	var user model.User
	if input.IDCard != "" {
		logger = logger.WithFields(log.Fields{"username": input.IDCard})

		db.Set("gorm:auto_preload", true).Where("username = ?", input.IDCard).First(&user)
		if user.Username == input.IDCard {
			result = model.NewFailureResult().SetLogger(logger)
			result.AddInfo("您的身分證已經登錄過")
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
			return
		}
	} else {
		logger = logger.WithFields(log.Fields{"username": input.Phone})

		db.Set("gorm:auto_preload", true).Where("username = ?", input.Phone).First(&user)
		if user.Username == input.Phone {
			result = model.NewFailureResult().SetLogger(logger)
			user.Mask(model.RegisterMask)

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
	var redisCtx context.Context = model.RedisContext
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
		redisCtx,
		key,
		otp,
		1*time.Minute,
	)
	if !redisResult.Val() {
		ttl := redisClient.TTL(redisCtx, key)
		msg := fmt.Sprintf("SMS 請求太頻繁，請於 %s 後再嘗試", ttl.Val().String())

		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo(msg)
		result.SetData(ttl.Val().String())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger.Infoln("SMS OTP 已加入 Redis")

	guest.PhoneToken = token
	guest.PhoneVerify = false
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

	guest.Mask(model.RegisterMask)

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
	var enforcer *casbin.Enforcer = model.Enforcer
	var redisClient *redis.Client = model.RedisClient
	var redisCtx context.Context = model.RedisContext
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
	otp, err := redisClient.Get(redisCtx, key).Result()
	if (err != nil) || (otp != input.PhoneOTP) {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("驗證碼不正確或已經失效")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	guest.PhoneVerify = true
	db.Save(&guest)

	authPath := fmt.Sprintf("/api/v1/guests/%d/qr", user.ID)
	enforcer.AddNamedPolicy("p", user.Username, authPath, "GET")
	logger.Infoln("新增訪客使用者權限")

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("您的手機已通過驗證")
	result.SetData(guest)

	ctx.JSON(http.StatusOK, result)
}

// VerifyGuestUserIDCardHandler ...
func VerifyGuestUserIDCardHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "VerifyGuestUserIDCard",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	userID, err := strconv.ParseUint(ctx.Param("user_id"), 10, 32)
	if err != nil {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料格式錯誤")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	logger = logger.WithFields(log.Fields{"user_id": userID})

	var user model.User
	db.Set("gorm:auto_preload", true).Where("id = ?", userID).First(&user)
	if user.ID != uint(userID) {
		result = model.NewFailureResult().SetLogger(logger)
		result.AddInfo("您輸入的資料不正確")
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, result)
		return
	}
	user.Guest.IDCardVerify = true
	db.Save(&user)

	user.Mask(model.VisitMask)

	result = model.NewSuccessResult().SetLogger(logger)
	result.AddInfo("您已完成身分證驗證")
	result.SetData(user)

	ctx.JSON(http.StatusOK, result)
}

// DeleteVisitRecordsAndGuestsHandler ...
func DeleteVisitRecordsAndGuestsHandler(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{
		"controller": "Guest",
		"event":      "DeleteVisitRecordsAndGuestsHandler",
	})

	var db *gorm.DB = model.DB
	var result *model.Result

	now := time.Now()
	today := now.Format("2006-01-02")
	todayStart := fmt.Sprintf("%s 00:00:00", now.Format("2006-01-02"))
	todayEnd := fmt.Sprintf("%s 23:59:59", now.Format("2006-01-02"))
	targetTime := now.AddDate(0, 0, -28)
	targetDate := fmt.Sprintf("%s 00:00:00", targetTime.Format("2006-01-02"))

	db.Exec(`
UPDATE visits
SET deleted_at = ?, 
    guest_user_id = 0
WHERE deleted_at IS NULL 
  AND created_at < ?`, now, targetDate)

	db.Exec(`
DELETE FROM followers 
WHERE visit_id IN (
  SELECT id FROM visits 
  WHERE guest_user_id = 0
)`)

	db.Exec(`
UPDATE users 
INNER JOIN guests ON guests.user_id=users.id
SET users.deleted_at = ?,
    users.username = LOWER(HEX(RANDOM_BYTES(8))),
    users.name = '○○○',
    guests.name = '○○○',
    guests.phone = '0000000000',
    guests.phone_token = '',
    guests.id_card = ''
WHERE users.deleted_at IS NULL
 AND users.type = 1 
 AND users.created_at < ?
 AND users.id NOT IN (
  SELECT visits.guest_user_id
  FROM visits 
  WHERE visits.deleted_at IS NULL
  GROUP BY visits.guest_user_id
)`, now, targetDate)

	db.Exec(`
DELETE FROM casbin_rule 
WHERE p_type LIKE 'p'
 AND v1 LIKE '%qr'
 AND v2 LIKE 'GET'
 AND v0 NOT IN (
  SELECT username FROM users
)`)

	model.Enforcer.LoadPolicy()

	result = model.NewSuccessResult().SetLogger(logger)

	var count int
	db.Unscoped().Model(&model.Visit{}).Where("deleted_at BETWEEN ? AND ?", todayStart, todayEnd).Count(&count)
	result.AddInfo(fmt.Sprintf("您已於 %s 清除 %d 筆洽公紀錄", today, count))

	db.Unscoped().Model(&model.User{}).Where("deleted_at BETWEEN ? AND ?", todayStart, todayEnd).Count(&count)
	result.AddInfo(fmt.Sprintf("您已於 %s 清除 %d 筆驗證紀錄", today, count))

	ctx.JSON(http.StatusOK, result)
}
