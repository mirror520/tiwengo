package environment

import "os"

var (
	// BaseURL ...
	BaseURL = os.Getenv("TIWENGO_BASE_URL")

	// ShortURL ...
	ShortURL = os.Getenv("TIWENGO_SHORT_URL")

	// TCCGBaseURL ...
	TCCGBaseURL = os.Getenv("TCCG_BASE_URL")

	// TokenSecret ...
	TokenSecret = os.Getenv("TOKEN_SECRET")

	// SMSBaseURL ...
	SMSBaseURL = "https://oms.every8d.com/API21/HTTP"

	// SMSUsername ...
	SMSUsername = os.Getenv("SMS_UID")

	// SMSPassword ...
	SMSPassword = os.Getenv("SMS_PWD")
)
