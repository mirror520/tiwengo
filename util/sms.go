package util

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/model"
)

var (
	shortURL    = environment.ShortURL
	smsBaseURL  = environment.SMSBaseURL
	smsUsername = environment.SMSUsername
	smsPassword = environment.SMSPassword
)

const shortSMSLen = 70

// SMS ...
type SMS struct {
	UID  string
	PWD  string
	SB   string
	MSG  string
	DEST string
}

// SMSResult ...
type SMSResult struct {
	Credit  float64
	Sended  uint64
	Cost    uint64
	Unsend  uint64
	BatchID string
}

// NewSMS ...
func NewSMS() (*SMS, error) {
	if smsUsername == "" || smsPassword == "" {
		return nil, errors.New("未設定簡訊寄送帳號密碼")
	}

	return &SMS{
		UID:  smsUsername,
		PWD:  smsPassword,
		SB:   "",
		MSG:  "",
		DEST: "",
	}, nil
}

// SetOTP ...
func (s *SMS) SetOTP(guest *model.Guest) (string, string) {
	subject := fmt.Sprintf("驗證訪客: %s, 行動電話: %s", guest.Name, guest.Phone)
	otp, _ := getRandNum()
	token := generateRandomString(30)
	originMsg := fmt.Sprintf("驗證碼: %s ,再次登入: %s?t=%s", otp, shortURL, token)
	limitMsg := string([]rune(originMsg)[0:shortSMSLen])

	re := regexp.MustCompile(`^.*\?t=(?P<token>.*)`)
	token = re.ReplaceAllString(limitMsg, `${token}`)

	s.SB = subject
	s.MSG = limitMsg
	s.DEST = guest.Phone

	return otp, token
}

// SendSMS ...
func (s *SMS) SendSMS() (*SMSResult, error) {
	data := s.URLValues()

	resp, err := http.PostForm(smsBaseURL+"/sendSMS.ashx", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	contents := strings.Split(string(body), ",")

	credit, _ := strconv.ParseFloat(contents[0], 64)
	if credit < 0 {
		return nil, errors.New(strings.Trim(contents[1], " "))
	}

	sended, _ := strconv.ParseUint(contents[1], 10, 64)
	cost, _ := strconv.ParseUint(contents[2], 10, 64)
	unsend, _ := strconv.ParseUint(contents[3], 10, 64)
	batchID := contents[4]

	result := &SMSResult{
		Credit:  credit,
		Sended:  sended,
		Cost:    cost,
		Unsend:  unsend,
		BatchID: batchID,
	}

	return result, nil
}

// URLValues ...
func (s *SMS) URLValues() url.Values {
	return url.Values{
		"UID":  {s.UID},
		"PWD":  {s.PWD},
		"SB":   {s.SB},
		"MSG":  {s.MSG},
		"DEST": {s.DEST},
	}
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
