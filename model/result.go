package model

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

// ResultStatus ...
type ResultStatus string

const (
	// Success ...
	Success ResultStatus = "success"

	// Failure ...
	Failure ResultStatus = "failure"
)

// Result ...
type Result struct {
	Status ResultStatus `json:"status"`
	Info   []string     `json:"info"`
	Data   interface{}  `json:"data"`
	Time   time.Time    `json:"time"`
	Logger *log.Entry   `json:"-"`
}

// New ...
func New(status ResultStatus) *Result {
	return &Result{
		Status: status,
		Info:   nil,
		Data:   nil,
		Time:   time.Now(),
	}
}

// NewSuccessResult ...
func NewSuccessResult() *Result {
	return New(Success)
}

// NewFailureResult ...
func NewFailureResult() *Result {
	return New(Failure)
}

// SetData ...
func (r *Result) SetData(data interface{}) {
	r.Data = data
}

// AddInfo ...
func (r *Result) AddInfo(info string) {
	r.Info = append(r.Info, info)

	if r.Logger != nil {
		if r.Status == Success {
			r.Logger.Infoln(info)
		} else {
			r.Logger.Errorln(info)
		}
	}
}

// SetLogger ...
func (r *Result) SetLogger(logger *log.Entry) *Result {
	r.Logger = logger
	return r
}

// JSON ...
func (r *Result) JSON() []byte {
	b, _ := json.Marshal(r)
	return b
}
