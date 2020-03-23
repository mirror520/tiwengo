package model

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// DB ...
var DB *gorm.DB

// RedisClient ...
var RedisClient *redis.Client
