package main

import (
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	casbin "github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/database"
	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/middleware"
	"github.com/mirror520/tiwengo/model"
	"github.com/mirror520/tiwengo/route"
	cors "github.com/rs/cors/wrapper/gin"
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	dbArgs := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=True&loc=%s",
		environment.DBUsername,
		environment.DBPassword,
		environment.DBHost,
		environment.DBName,
		"Asia%2FTaipei",
	)

	var err error
	model.DB, err = gorm.Open("mysql", dbArgs)
	if err != nil {
		log.WithFields(log.Fields{"db": "mysql"}).
			Fatalln(err.Error())
	}
	defer model.DB.Close()

	model.RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", environment.RedisHost),
		Password: "",
		DB:       0,
	})
	defer model.RedisClient.Close()

	if err := model.RedisClient.Ping().Err(); err != nil {
		log.WithFields(log.Fields{"db": "redis"}).
			Fatalln(err.Error())
	}

	database.Migrate(model.DB)
	// database.Seed(model.DB)

	adapter, _ := gormadapter.NewAdapterByDB(model.DB)
	enforcer, _ := casbin.NewEnforcer("keymatch_rbac_model.conf", adapter)
	enforcer.LoadPolicy()

	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/privkeys/today", "GET")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/visits/", "GET")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/visits/users/:username", "PUT")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/visits/locations", "GET")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/guests/verify/:user_id/idcard", "PATCH")

	model.Enforcer = enforcer

	router := gin.Default()
	authMiddleware, err := jwt.New(middleware.AuthMiddleware())
	router.Use(cors.AllowAll())
	route.SetRoute(router, authMiddleware)

	router.Run(":6080")
}
