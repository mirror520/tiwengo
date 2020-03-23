package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/mirror520/tiwengo/database"
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
		log.Fatal("Failed to connect database")
	}
	defer model.DB.Close()

	model.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	defer model.RedisClient.Close()

	database.Migrate(model.DB)
	database.Seed(model.DB)

	router := gin.Default()
	router.Use(cors.AllowAll())
	route.SetRoute(router)

	router.Run(":6080")
}
