package main

import (
	"testing"

	"github.com/mirror520/tiwengo/model"
	"github.com/stretchr/testify/assert"
)

func TestConnDB(t *testing.T) {
	assert := assert.New(t)

	db := connDB()

	var tables []struct{}
	db.Raw("SHOW TABLES").Scan(&tables)

	assert.Equal(11, len(tables), "11 tables initialize")
}

func TestConnRedis(t *testing.T) {
	assert := assert.New(t)

	ctx := model.RedisContext
	client := connRedis(ctx)

	pong := client.Ping(ctx).Val()
	assert.Equal("PONG", pong, "ping success")
}
