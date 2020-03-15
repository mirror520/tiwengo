package model

import (
	"encoding/json"
	"time"
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
	Data   string       `json:"data"`
	Time   time.Time    `json:"time"`
}

// New ...
func New(status ResultStatus) *Result {
	return &Result{
		Status: status,
		Info:   nil,
		Data:   "",
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
func (r *Result) SetData(data string) {
	r.Data = data
}

// AddInfo ...
func (r *Result) AddInfo(info string) {
	r.Info = append(r.Info, info)
}

// JSON ...
func (r *Result) JSON() []byte {
	b, _ := json.Marshal(r)
	return b
}
