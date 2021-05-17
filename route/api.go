package route

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/mirror520/tiwengo/controller"
)

// SetRoute ...
func SetRoute(router *gin.Engine, authMiddleware *jwt.GinJWTMiddleware, limitMiddleware gin.HandlerFunc) {
	apiV1 := router.Group("/api/v1")
	{
		privkeys := apiV1.Group("/privkeys")
		{
			privkeys.GET("/today", authMiddleware.MiddlewareFunc(), controller.GetPrivkeyHandler)
		}

		users := apiV1.Group("/users")
		users.Use(limitMiddleware)
		{
			users.PATCH("/tccg/login", authMiddleware.LoginHandler)
			users.PATCH("/mqtt/token", authMiddleware.MiddlewareFunc(), controller.RefreshMQTTUserTokenHandler)
		}

		guests := apiV1.Group("/guests")
		{
			guests.GET("/:user_id/qr", authMiddleware.MiddlewareFunc(), controller.ShowGuestUserQRCodeHandler)
			guests.PATCH("/verify/:user_id/idcard", authMiddleware.MiddlewareFunc(), controller.VerifyGuestUserIDCardHandler)

			guests.PATCH("/login", limitMiddleware, authMiddleware.LoginHandler)
			guests.POST("/register", limitMiddleware, controller.RegisterGuestUserHandler)
			guests.PATCH("/register/phone/otp", limitMiddleware, controller.SendGuestPhoneOTPHandler)
			guests.PATCH("/register/phone/otp/verify", limitMiddleware, controller.VerifyGuestPhoneOTPHandler)
		}

		visits := apiV1.Group("/visits")
		{
			// visits.GET("/", authMiddleware.MiddlewareFunc(), controller.ListAllGuestVisitRecordHandler)
			visits.PUT("/users/:username", authMiddleware.MiddlewareFunc(), controller.UserVisitHandler)
			visits.GET("/buildings", authMiddleware.MiddlewareFunc(), controller.GetBuildingsHandler)

			visits.PUT("/tcpass/users/:uuid", authMiddleware.MiddlewareFunc(), controller.TcpassUserVisitHandler)
		}

		auth := apiV1.Group("/auth")
		{
			auth.PATCH("/refresh_token", limitMiddleware, authMiddleware.RefreshHandler)
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
