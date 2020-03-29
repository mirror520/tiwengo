package route

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/controller"
)

// SetRoute ...
func SetRoute(router *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	router.Group("/api/v1")
	apiV1 := router.Group("/api/v1")
	{
		privkeys := apiV1.Group("/privkeys")
		{
			privkeys.GET("/today", controller.GetPrivkeyHandler)
		}

		users := apiV1.Group("/users")
		{
			users.PATCH("/tccg/login", authMiddleware.LoginHandler)
		}

		guests := apiV1.Group("/guests")
		{
			guests.PATCH("/login", authMiddleware.LoginHandler)
			guests.GET("/:user_id/qr", controller.ShowGuestUserQRCodeHandler)
			guests.POST("/register", controller.RegisterGuestUserHandler)
			guests.PATCH("/register/phone/otp", controller.SendGuestPhoneOTPHandler)
			guests.PATCH("/register/phone/otp/verify", controller.VerifyGuestPhoneOTPHandler)
		}

		auth := apiV1.Group("/auth")
		{
			auth.PATCH("/refresh_token", authMiddleware.RefreshHandler)
		}
	}
}
