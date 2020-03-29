package model

import "time"

// Token ...
type Token struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}
