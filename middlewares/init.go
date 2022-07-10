package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Secure(client *mongo.Client) gin.HandlerFunc {
	col := client.Database("ddsrdb").Collection("apikeys")
	return func(c *gin.Context) {
		key := c.Request.Header.Get("x-api-key")
		if len(key) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		res := col.FindOne(context.TODO(), bson.M{"value": key})
		if res.Err() != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		c.Next()
	}
}
