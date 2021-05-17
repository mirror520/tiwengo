package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	assert := assert.New(t)

	key := hashPassword("hello world")

	assert.Contains(key, "PBKDF2")
	assert.Len(key, 67)
}
