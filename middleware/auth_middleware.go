package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/controller"
	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/model"

	jwt "github.com/appleboy/gin-jwt/v2"
	log "github.com/sirupsen/logrus"
)

var (
	baseURL     = environment.BaseURL
	tokenSecret = environment.TokenSecret
)

// AuthMiddleware ...
func AuthMiddleware() *jwt.GinJWTMiddleware {
	logger := log.WithFields(log.Fields{"middleware": "Auth"})
	identityKey := "username"
	enforcer := model.Enforcer

	return &jwt.GinJWTMiddleware{
		Realm:       baseURL,
		Key:         []byte(tokenSecret),
		Timeout:     30 * time.Minute,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(ctx *gin.Context) interface{} {
			claims := jwt.ExtractClaims(ctx)
			username := claims[identityKey].(string)

			return username
		},
		Authenticator: func(ctx *gin.Context) (interface{}, error) {
			var loginVals model.User
			if err := ctx.ShouldBind(&loginVals); err != nil {
				fmt.Println(err.Error())
				return nil, jwt.ErrMissingLoginValues
			}

			if loginVals.Type == model.EmployeeUser {
				tccgUser := model.TccgUser{
					Account:  loginVals.Username,
					Password: loginVals.Password,
				}

				user, err := controller.LoginTccgUserHandler(&tccgUser)
				ctx.Set("user", user)
				ctx.Set("username", loginVals.Username)

				return user, err
			}

			if loginVals.Type == model.GuestUser {
				guest := model.Guest{
					Phone:      loginVals.Username,
					PhoneToken: loginVals.Password,
				}

				user, err := controller.LoginGuestUserHandler(&guest)
				ctx.Set("user", user)
				ctx.Set("username", loginVals.Username)

				return user, err
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, ctx *gin.Context) bool {
			logger = logger.WithFields(log.Fields{"event": "Authorizator"})

			if val, ok := data.(string); ok {
				logger = logger.WithFields(log.Fields{"username": val})

				result, err := enforcer.Enforce(val, ctx.Request.URL.Path, ctx.Request.Method)
				if err != nil {
					logger.Errorln(err.Error())
				}

				return result
			}

			return false
		},
		Unauthorized: func(ctx *gin.Context, code int, message string) {
			logger = logger.WithFields(log.Fields{"event": "Unauthorized"})

			value, ok := ctx.Get("username")
			if ok {
				username := value.(string)
				logger = logger.WithFields(log.Fields{"username": username})
			}

			result := model.NewFailureResult().SetLogger(logger)
			result.AddInfo(message)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, result)
		},
		LoginResponse: func(ctx *gin.Context, code int, token string, expire time.Time) {
			logger = logger.WithFields(log.Fields{"event": "LoginResponse"})

			value, ok := ctx.Get("user")
			user := value.(*model.User)
			if ok {
				user.Token = model.Token{
					Token:  token,
					Expire: expire,
				}
			}
			logger = logger.WithFields(log.Fields{"username": user.Username})

			result := model.NewSuccessResult().SetLogger(logger)
			result.AddInfo("您已登入成功")
			result.SetData(&user)

			ctx.JSON(http.StatusOK, result)
		},
		RefreshResponse: func(ctx *gin.Context, code int, token string, expire time.Time) {
			logger = logger.WithFields(log.Fields{"event": "RefreshResponse"})

			newToken := model.Token{
				Token:  token,
				Expire: expire,
			}

			result := model.NewSuccessResult().SetLogger(logger)
			result.AddInfo("您已成功更新 TOKEN")
			result.SetData(&newToken)

			ctx.JSON(http.StatusOK, result)
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		SendCookie:    false,
	}
}
