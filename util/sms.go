package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/model"
)

var (
	shortURL   = environment.ShortURL
	smsBaseURL = environment.SMSBaseURL
)

const shortSMSLen = 70

// SMS ...
type SMS struct {
	Phone   string `json:"phone" binding:"required"`
	Message string `json:"message" binding:"required"`
	Comment string `json:"comment"`
}

// SMSResult ...
type SMSResult struct {
	ID     string
	Credit int
}

// NewSMS ...
func NewSMS() *SMS {
	return &SMS{}
}

// SetOTP ...
func (s *SMS) SetOTP(guest *model.Guest) (string, string) {
	subject := fmt.Sprintf("行動電話: %s", guest.Phone)
	otp, _ := getRandNum()
	token := generateRandomString(30)
	originMsg := fmt.Sprintf("驗證碼: %s ,再次登入: %s/t/%s", otp, shortURL, token)
	limitMsg := string([]rune(originMsg)[0:shortSMSLen])

	re := regexp.MustCompile(`^.*/t/(?P<token>.*)`)
	token = re.ReplaceAllString(limitMsg, `${token}`)

	s.Phone = guest.Phone
	s.Message = limitMsg
	s.Comment = subject

	return otp, token
}

// SendSMS ...
func (s *SMS) SendSMS() (*SMSResult, error) {
	client := resty.New().
		SetHostURL(smsBaseURL)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(s).
		SetResult(&SMSResult{}).
		Post("/send")

	result := resp.Result().(*SMSResult)
	return result, err
}

func getRandNum() (string, error) {
	nBig, e := rand.Int(rand.Reader, big.NewInt(8999))
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(nBig.Int64()+1000, 10), nil
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)

	return base64.URLEncoding.EncodeToString(b)
}
