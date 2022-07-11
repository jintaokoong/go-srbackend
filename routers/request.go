package routers

import (
	"context"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rmrf-slash.com/go-srbackend/configurations/logger"
)

type SongRequest struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Key       string             `bson:"key,omitempty" json:"key"`
	Name      string             `bson:"name,omitempty" json:"name"`
	Audience  string             `bson:"audience,omitempty" json:"audience"`
	Done      bool               `bson:"done" json:"done"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type CreateSongRequest struct {
	Name     string `json:"name" binding:"required"`
	Audience string `json:"audience"`
}

type SingleRequestURI struct {
	ID string `uri:"id"`
}

type RequestListParams struct {
	Page     int64 `form:"page" binding:"gt=0"`
	PageSize int64 `form:"pageSize" binding:"gt=0"`
}

func ListRequests(client *mongo.Client) func(ctx *gin.Context) {
	col := client.Database("ddsrdb").Collection("requests")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		t := time.Now()
		params := RequestListParams{
			Page:     1,
			PageSize: 10,
		}
		if ctx.BindQuery(&params) != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"message": "bad query params"})
			return
		}
		log.Println("params bound", "time elapsed:", time.Since(t))

		var requests []SongRequest = make([]SongRequest, 0)

		skip := (params.Page - 1) * params.PageSize
		opts := options.FindOptions{
			Skip:  &skip,
			Limit: &params.PageSize,
			Sort:  bson.M{"updatedAt": -1},
		}

		filter := bson.M{}
		cursor, err := col.Find(ctx, filter, &opts)
		if err != nil {
			log.Println("error:", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error querying data",
			})
			return
		}
		if err = cursor.All(ctx, &requests); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error querying data",
			})
			return
		}
		log.Println("documents found", "time elapsed:", time.Since(t))

		count, err := col.CountDocuments(ctx, filter)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "error querying data",
			})
			return
		}
		log.Println("documents counted", "time elapsed:", time.Since(t))

		ctx.JSON(http.StatusOK, gin.H{
			"data":       requests,
			"records":    count,
			"page":       params.Page,
			"pageSize":   params.PageSize,
			"totalPages": math.Ceil(float64(count) / float64(params.PageSize)),
		})
	}
}

func CreateRequest(client *mongo.Client) func(ctx *gin.Context) {
	col := client.Database("ddsrdb").Collection("requests")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		body := CreateSongRequest{
			Audience: "系統",
		}
		if err := ctx.BindJSON(&body); err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		t := time.Now()
		time.LoadLocation("Asia/Taipei")

		r := SongRequest{
			Key:       t.Format("2006-01-02"),
			Name:      body.Name,
			Audience:  body.Audience,
			Done:      false,
			CreatedAt: t,
			UpdatedAt: t,
		}
		res, err := col.InsertOne(context.TODO(), r)
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		newID, ok := res.InsertedID.(primitive.ObjectID)
		if ok {
			r.ID = newID
		}

		ctx.JSON(200, r)
	}
}

func ToggleRequest(client *mongo.Client) func(ctx *gin.Context) {
	col := client.Database("ddsrdb").Collection("requests")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		uri := SingleRequestURI{}
		if err := ctx.BindUri(&uri); err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		id, err := primitive.ObjectIDFromHex(uri.ID)
		if err != nil {
			log.Println(err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		time.LoadLocation("Asia/Taipei")
		update := []bson.M{{"$set": bson.M{"done": bson.M{"$not": "$done"}, "updatedAt": time.Now()}}}
		after := options.After
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}
		res := col.FindOneAndUpdate(context.TODO(), bson.M{"_id": id}, update, &opt)
		if res.Err() != nil {
			log.Println(res.Err().Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}

		updated := SongRequest{}
		err = res.Decode(&updated)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		ctx.JSON(200, updated)
	}
}

func DeleteRequest(client *mongo.Client) func(ctx *gin.Context) {
	col := client.Database("ddsrdb").Collection("requests")
	log := logger.GetInstance()
	return func(ctx *gin.Context) {
		uri := SingleRequestURI{}
		if err := ctx.BindUri(&uri); err != nil {
			log.Println(err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		id, err := primitive.ObjectIDFromHex(uri.ID)
		if err != nil {
			log.Println(err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		r := col.FindOneAndDelete(context.TODO(), bson.M{"_id": id})
		if r.Err() != nil {
			log.Println(r.Err().Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": r.Err().Error()})
			return
		}

		ctx.Status(200)
	}
}
