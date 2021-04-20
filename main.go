package main

import (
	"context"
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	casbin "github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/database"
	"github.com/mirror520/tiwengo/environment"
	"github.com/mirror520/tiwengo/middleware"
	"github.com/mirror520/tiwengo/model"
	"github.com/mirror520/tiwengo/route"
	cors "github.com/rs/cors/wrapper/gin"
	log "github.com/sirupsen/logrus"
	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func connDB() *gorm.DB {
	dbArgs := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=True&loc=%s",
		environment.DBUsername,
		environment.DBPassword,
		environment.DBHost,
		environment.DBName,
		"Asia%2FTaipei",
	)

	db, err := gorm.Open("mysql", dbArgs)
	if err != nil {
		log.WithFields(log.Fields{"db": "mysql"}).
			Fatalln(err.Error())
	}

	database.Migrate(db)
	// database.Seed(db)

	return db
}

func connRedis(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", environment.RedisHost),
		Password: "",
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.WithFields(log.Fields{"db": "redis"}).
			Fatalln(err.Error())
	}

	return client
}

func loadCasbinEnforcer(db *gorm.DB) *casbin.Enforcer {
	adapter, _ := gormadapter.NewAdapterByDB(db)
	enforcer, _ := casbin.NewEnforcer("keymatch_rbac_model.conf", adapter)
	enforcer.LoadPolicy()

	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/privkeys/today", "GET")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/visits/users/:username", "PUT")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/visits/buildings", "GET")
	enforcer.AddNamedPolicy("p", "tccg_user", "/api/v1/guests/verify/:user_id/idcard", "PATCH")

	return enforcer
}

func createLimitMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	rate, _ := limiter.NewRateFromFormatted(environment.APILimitRate)
	store, _ := sredis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   "limiter_gin",
		MaxRetry: 3,
	})

	limitMiddleware := mgin.NewMiddleware(
		limiter.New(store, rate),
		mgin.WithLimitReachedHandler(func(ctx *gin.Context) {
			logger := log.WithFields(log.Fields{
				"client": ctx.ClientIP(),
			})

			result := model.NewFailureResult().SetLogger(logger)
			result.AddInfo("您嘗試使用資源的次數太多囉，請休息一下！")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, result)
		}),
	)

	return limitMiddleware
}

func externalRouter() *gin.Engine {
	router := gin.Default()
	authMiddleware, _ := jwt.New(middleware.AuthMiddleware())
	limitMiddleware := createLimitMiddleware(model.RedisClient)
	router.Use(cors.AllowAll())
	route.SetRoute(router, authMiddleware, limitMiddleware)
	return router
}

func internalRouter() *gin.Engine {
	router := gin.Default()
	route.SetAdminRoute(router)
	return router
}

func main() {
	model.DB = connDB()
	model.RedisClient = connRedis(model.RedisContext)
	model.Enforcer = loadCasbinEnforcer(model.DB)

	defer model.DB.Close()
	defer model.RedisClient.Close()

	go internalRouter().Run(":9000")
	externalRouter().Run(":6080")
}
