package main

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	casbin "github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/database"
	"github.com/mirror520/tiwengo/middleware"
	"github.com/mirror520/tiwengo/model"
	"github.com/mirror520/tiwengo/route"
	cors "github.com/rs/cors/wrapper/gin"
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	var err error
	model.DB, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.WithFields(log.Fields{"db": "sqlite3"}).
			Fatalln("資料庫連結失敗")
	}
	defer model.DB.Close()

	model.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	defer model.RedisClient.Close()

	if err := model.RedisClient.Ping().Err(); err != nil {
		log.WithFields(log.Fields{"db": "redis"}).
			Fatalln("Redis 資料庫連結失敗")
	}

	database.Migrate(model.DB)
	database.Seed(model.DB)

	adapter, _ := gormadapter.NewAdapterByDB(model.DB)
	enforcer, _ := casbin.NewEnforcer("keymatch_model.conf", adapter)
	enforcer.LoadPolicy()

	model.Enforcer = enforcer

	router := gin.Default()
	authMiddleware, err := jwt.New(middleware.AuthMiddleware())
	router.Use(cors.AllowAll())
	route.SetRoute(router, authMiddleware)

	router.Run(":6080")
}
