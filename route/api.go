package route

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/controller"
)

// SetRoute ...
func SetRoute(router *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	apiV1 := router.Group("/api/v1")
	{
		privkeys := apiV1.Group("/privkeys")
		{
			privkeys.GET("/today", authMiddleware.MiddlewareFunc(), controller.GetPrivkeyHandler)
		}

		users := apiV1.Group("/users")
		{
			users.PATCH("/tccg/login", authMiddleware.LoginHandler)
		}

		guests := apiV1.Group("/guests")
		{
			guests.GET("/:user_id/qr", authMiddleware.MiddlewareFunc(), controller.ShowGuestUserQRCodeHandler)
			guests.PATCH("/verify/:user_id/idcard", authMiddleware.MiddlewareFunc(), controller.VerifyGuestUserIDCardHandler)

			guests.PATCH("/login", authMiddleware.LoginHandler)
			guests.POST("/register", controller.RegisterGuestUserHandler)
			guests.PATCH("/register/phone/otp", controller.SendGuestPhoneOTPHandler)
			guests.PATCH("/register/phone/otp/verify", controller.VerifyGuestPhoneOTPHandler)
		}

		visits := apiV1.Group("/visits")
		{
			// visits.GET("/", authMiddleware.MiddlewareFunc(), controller.ListAllGuestVisitRecordHandler)
			visits.PUT("/users/:username", authMiddleware.MiddlewareFunc(), controller.UserVisitHandler)
			visits.GET("/buildings", authMiddleware.MiddlewareFunc(), controller.GetBuildingsHandler)
		}

		auth := apiV1.Group("/auth")
		{
			auth.PATCH("/refresh_token", authMiddleware.RefreshHandler)
		}
	}
}

// SetAdminRoute ...
func SetAdminRoute(router *gin.Engine) {
	apiV1 := router.Group("/api/v1")
	{
		guests := apiV1.Group("/guests")
		{
			guests.DELETE("/today/visits", controller.DeleteVisitRecordsAndGuestsHandler)
		}
	}
}
