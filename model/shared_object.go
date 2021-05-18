package model

import (
	"context"

	"github.com/casbin/casbin/v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
)

// DB ...
var DB *gorm.DB

// RedisClient ...
var RedisClient *redis.Client

// RedisContext ...
var RedisContext = context.Background()

// MQTTClient ...
var MQTTClient mqtt.Client

// Enforcer ...
var Enforcer *casbin.Enforcer
