package routers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rmrf-slash.com/go-srbackend/configurations/logger"
)

type ApiError struct {
	Message string `json:"message"`
}

type Configuration struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	Name      string             `bson:"name,omitempty" json:"name"`
	Value     any                `bson:"value,omitempty" json:"value"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"-"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"-"`
}

func FindStatus(client *mongo.Client) gin.HandlerFunc {
	col := client.Database("ddsrdb").Collection("configs")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		config := Configuration{}
		res := col.FindOne(context.TODO(), bson.M{"name": "accepting"})
		if res.Err() != nil {
			log.Println(res.Err().Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, ApiError{Message: res.Err().Error()})
			return
		}
		res.Decode(&config)
		ctx.JSON(http.StatusOK, config)
	}
}

func ToggleStatus(client *mongo.Client) gin.HandlerFunc {
	col := client.Database("ddsrdb").Collection("configs")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		time.LoadLocation("Asia/Taipei")
		update := []bson.M{{"$set": bson.M{"value": bson.M{"$not": "$value"}, "updatedAt": time.Now()}}}
		after := options.After
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}
		res := col.FindOneAndUpdate(context.TODO(), bson.M{"name": "accepting"}, update, &opt)
		if res.Err() != nil {
			log.Println(res.Err().Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, ApiError{Message: res.Err().Error()})
			return
		}

		updated := Configuration{}
		err := res.Decode(&updated)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, ApiError{Message: err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, updated)
	}
}
