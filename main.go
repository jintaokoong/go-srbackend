package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rmrf-slash.com/go-srbackend/configurations/environments"
	"rmrf-slash.com/go-srbackend/middlewares"
	"rmrf-slash.com/go-srbackend/routers"
)

func main() {
	cfg := environments.GetVariables()

	ctx := context.TODO()
	opts := options.Client().SetAuth(options.Credential{Username: cfg.MongoUsername, Password: cfg.MongoPassword}).ApplyURI(cfg.MongoConnection)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(ctx)

	origins := strings.Split(cfg.AllowOrigins, ",")
	corsConfig := cors.DefaultConfig()
	corsConfig.AddAllowHeaders("x-api-key")
	corsConfig.AllowOrigins = origins

	r := gin.Default()
	r.Use(cors.New(corsConfig))
	r.SetTrustedProxies(nil)
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "service is up!",
		})
	})
	r.GET("/health", func(ctx *gin.Context) {
		ctx.Status(200)
	})

	apis := r.Group("/api", middlewares.Secure(client))
	{
		apis.GET("/requests", routers.ListRequests(client))
		apis.POST("/requests", routers.CreateRequest(client))
		apis.PATCH("/requests/:id", routers.ToggleRequest(client))
		apis.DELETE("/requests/:id", routers.DeleteRequest(client))
	}
	r.Run()
}
