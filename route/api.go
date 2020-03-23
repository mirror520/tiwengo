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
		// privkeys := apiV1.Group("/privkeys")
		// {
		// 	privkeys.GET("/", indexPrivkeysHandler)
		// 	privkeys.POST("/:date", createPrivkeyHandler)
		// 	privkeys.PUT("/:date", updatePrivkeyHandler)
		// 	privkeys.PATCH("/:date", updatePrivkeyHandler)
		// 	privkeys.GET("/:date/qr", showPrivkeyQrCodeHandler)
		// 	privkeys.GET("/:date/ciphertext", showPrivkeyCiphertextHandler)
		// }

		users := apiV1.Group("/users")
		{
			// 管理者權限
			// users.GET("/") // ListAllUsersHandler

			users.PATCH("/tccg/login", controller.LoginTccgUserHandler)
			// users.GET("/tccg/:account/qr") // ShowTccgUserQRCodeHandler
		}

		guests := apiV1.Group("/guests")
		{
			// guests.GET("/")                        // ListAllGuestsHandler
			guests.POST("/register", controller.RegisterGuestUserHandler)
			guests.POST("/register/phone/otp", controller.SendGuestPhoneOTPHandler)
			guests.PATCH("/register/phone/otp/verify", controller.VerifyGuestPhoneOTPHandler)
			// guests.POST("/register/idcard/verify") // VerifyGuestIDCardHandler
			// guests.PATCH("/login")                 // LoginGuestUserHandler
			// guests.GET("/:guest_id/qr")            // ShowGuestQRCodeHandler
		}

		// visits := apiV1.Group("/visits")
		// {
		// 	visits.POST("/:guest_id/:department_id") // AddGuestVisitedDepartmentHandler

		// 	visits.GET("/")                // ListAllVisitedInfoHandler
		// 	visits.GET("/:date")           // ListSpecificDateGuestVisistedInfoHandler
		// 	visits.GET("/:institution_ou") // ListSpecificInstitutionGuestVisitedInfoHandler
		// 	visits.GET("/:department_ou")  // ListSpecificDepartmentGuestVisitedInfoHandler
		// 	visits.GET("/:guest_id")       // ListSpecificGuestVisitedInfoHandler
		// }
	}
}
