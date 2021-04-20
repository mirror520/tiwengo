package environment

import "os"

var (
	// BaseURL ...
	BaseURL = os.Getenv("TIWENGO_BASE_URL")

	// ShortURL ...
	ShortURL = os.Getenv("TIWENGO_SHORT_URL")

	// TCCGBaseURL ...
	TCCGBaseURL = os.Getenv("TCCG_BASE_URL")

	// SMSBaseURL ...
	SMSBaseURL = os.Getenv("SMS_BASE_URL")

	// TokenSecret ...
	TokenSecret = os.Getenv("TOKEN_SECRET")

	// DBHost ...
	DBHost = os.Getenv("DB_HOST")

	// DBUsername ...
	DBUsername = os.Getenv("DB_USERNAME")

	// DBPassword ...
	DBPassword = os.Getenv("DB_PASSWORD")

	// DBName ...
	DBName = os.Getenv("DB_NAME")

	// RedisHost ...
	RedisHost = os.Getenv("REDIS_HOST")

	// APILimitRate ...
	APILimitRate = os.Getenv("API_LIMIT_RATE")
)
