package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"rmrf-slash.com/go-srbackend/configurations/logger"
	"rmrf-slash.com/go-srbackend/routers"
)

func Secure(client *mongo.Client) gin.HandlerFunc {
	col := client.Database("ddsrdb").Collection("apikeys")
	log := logger.GetInstance()
	return func(c *gin.Context) {
		key := c.Request.Header.Get("x-api-key")
		if len(key) == 0 {
			log.Println("key:", key, len(key))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		res := col.FindOne(context.TODO(), bson.M{"value": key})
		if res.Err() != nil {
			log.Println(res.Err().Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		c.Next()
	}
}

func CheckStatus(client *mongo.Client) gin.HandlerFunc {
	col := client.Database("ddsrdb").Collection("configs")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		config := routers.Configuration{}
		res := col.FindOne(context.TODO(), bson.M{"name": "accepting"})
		err := res.Decode(&config)
		if err != nil {
			log.Println(err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, routers.ApiError{Message: err.Error()})
			return
		}

		if config.Value == false {
			log.Println("block due to accepting status")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, routers.ApiError{Message: "currently not accepting"})
			return
		}

		ctx.Next()
	}
}
