package model

import (
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
)

// DB ...
var DB *gorm.DB

// RedisClient ...
var RedisClient *redis.Client
