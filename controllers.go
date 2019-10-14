package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func ShortUrlResolveHandler(c *gin.Context) {
	url := c.Param("shortUrl")
	if url == "" {
		c.AbortWithStatus(400)
	}

	cachedLongUrl := RedisClient.Get(url).Val()

	response := make(map[string]string)

	if cachedLongUrl != "" {
		log.Printf("Entry %s found in cache", url)
		response["url"] = cachedLongUrl
		c.JSON(200, response)
		return
	}

	log.Printf("Entry %s not found in cache", url)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	urlFromDB := Url{}
	result := MongoCollection.FindOne(ctx, bson.M{"short_url": url})

	if result.Err() != nil {
		c.AbortWithStatus(404)
		return
	}

	result.Decode(&urlFromDB)

	if time.Now().Unix() > urlFromDB.Expiry.Unix() {
		c.AbortWithError(http.StatusNotFound, errors.New("Expired"))
		return
	}

	cacheSetResult := RedisClient.Set(urlFromDB.ShortURL, urlFromDB.LongURL, time.Minute*5)

	if cacheSetResult.Err() != nil {
		c.AbortWithStatus(500)
		return
	}
	response["url"] = urlFromDB.LongURL
	c.JSON(200, response)
}

func ShortenUrlHandler(c *gin.Context) {
	urlDataBytes, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	var urlData Url
	err = json.Unmarshal(urlDataBytes, &urlData)

	if err != nil || urlData.LongURL == "" || urlData.Validity == 0 {
		c.AbortWithStatus(400)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	result := MongoCollection.FindOne(ctx, bson.M{"url": urlData.LongURL})

	if result.Err() == nil {
		_ = result.Decode(&urlData)
		c.JSON(200, map[string]string{"short_url": "http://localhost:8080/" + urlData.ShortURL})
		return
	}

	newUuid, _ := uuid.NewV4()

	h := sha1.New()
	h.Write(newUuid.Bytes())
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	urlData.ShortURL = sha1_hash[:8]

	urlData.Expiry = time.Now().Add(time.Second * time.Duration(urlData.Validity))

	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	_, err = MongoCollection.InsertOne(ctx, urlData)

	c.JSON(200, map[string]string{"short_url": "http://localhost:8080/" + urlData.ShortURL})
}
