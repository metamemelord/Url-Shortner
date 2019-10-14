package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
)

var MongoCollection *mongo.Collection
var RedisClient *redis.Client

func init() {
	log.Println("Connecting to cache...")
	clientOptions := redis.Options{Addr: os.Getenv("REDIS_ADDRESS")}
	RedisClient = redis.NewClient(&clientOptions)
	log.Println("Connected to Redis!")
	log.Println("Connecting to Mongo...")
	var err error
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_ADDRESS")))

	if err != nil {
		fmt.Println(err)
		log.Fatal("Invalid Mongo connection data")
	}

	err = client.Connect(context.Background())

	if err != nil {
		log.Fatal("Could not connect to DB")
	}
	log.Println("Connected to Mongo!")

	MongoCollection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("urls")
}

func main() {
	g := gin.New()
	g.GET("/:shortUrl", ShortUrlResolveHandler)
	g.POST("/shorten", ShortenUrlHandler)

	PORT := os.Getenv("TINYURL_PORT")
	if PORT == "" {
		PORT = "8080"
	}
	err := g.Run(":" + PORT)

	if err != nil {
		log.Fatal("Could not start the server!")
	}
}
