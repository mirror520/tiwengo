package route

import (
	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/controller"
)

// SetRoute ...
func SetRoute(router *gin.Engine) {
	router.Group("/api/v1")
	apiV1 := router.Group("/api/v1")
	{
		privkeys := apiV1.Group("/privkeys")
		{
			privkeys.GET("/today", controller.GetPrivkeyHandler)
		}

		users := apiV1.Group("/users")
		{
			users.PATCH("/tccg/login", controller.LoginTccgUserHandler)
		}

		guests := apiV1.Group("/guests")
		{
			guests.PATCH("/login", controller.LoginGuestUserHandler)
			guests.GET("/:user_id/qr", controller.ShowGuestUserQRCodeHandler)
			guests.POST("/register", controller.RegisterGuestUserHandler)
			guests.PATCH("/register/phone/otp", controller.SendGuestPhoneOTPHandler)
			guests.PATCH("/register/phone/otp/verify", controller.VerifyGuestPhoneOTPHandler)
		}
	}
}
