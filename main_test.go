package main

import (
	"testing"

	"github.com/mirror520/tiwengo/model"
	"github.com/stretchr/testify/assert"
)

func TestConnRedis(t *testing.T) {
	assert := assert.New(t)

	ctx := model.RedisContext
	client := connRedis(ctx)

	pong := client.Ping(ctx).Val()
	assert.Equal("PONG", pong, "ping success")
}
