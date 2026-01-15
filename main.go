package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName   = "taskdb"
	collectionName = "tasks"
	port           = ":8000"
)

type Task struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed,omitempty"`
}

var router *gin.Engine

var Client *mongo.Client

func SetUpDatabase() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return nil, fmt.Errorf("MONGO_URI not set")
	}
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	Client = client
	return client, nil
}

func throwError(status int, err error, c *gin.Context) {
	code := status
	if code == 0 {
		code = http.StatusInternalServerError
	}
	c.JSON(code, gin.H{
		"error": err.Error(),
	})
}

func CreateTask(c *gin.Context) {
	var task Task
	err := c.BindJSON(&task)
	if err != nil {
		throwError(http.StatusBadRequest, fmt.Errorf("invalid json"), c)
		return
	}

	title := task.Title
	if title == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Title cannot be empty"})
		return
	}
	if len(title) < 5 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Title length must be at least 5"})
		return
	}
	task.Completed = false

	_, err = Client.Database(databaseName).Collection(collectionName).InsertOne(context.Background(), task)

	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	c.JSON(201, gin.H{
		"message": "Task created",
		"task":    task,
	})

}

func GetTasks(c *gin.Context) {
	var tasks []bson.M
	cursor, err := Client.Database(databaseName).Collection(collectionName).Find(context.Background(), bson.D{})
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}
	err = cursor.All(context.Background(), &tasks)
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	c.JSON(200, gin.H{
		"tasks": tasks,
	})

}

func GetSpecificTask(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		throwError(http.StatusBadRequest, err, c)
		return
	}
	var task Task
	err = Client.Database(databaseName).Collection(collectionName).FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	c.JSON(200, gin.H{
		"task": task,
	})
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		throwError(http.StatusBadRequest, err, c)
		return
	}

	var task Task
	err = c.BindJSON(&task)
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	if task.Title == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Title cannot be empty"})
		return
	}
	if len(task.Title) < 5 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "Title length must be at least 5"})
		return
	}

	res, err := Client.Database(databaseName).Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": task})
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Task not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Task updated",
	})
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		throwError(http.StatusBadRequest, err, c)
		return
	}

	res, err := Client.Database(databaseName).Collection(collectionName).DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		throwError(http.StatusInternalServerError, err, c)
		return
	}

	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Task not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Task deleted",
	})
}

func SetupRoutes() {
	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "hello")
	})
	router.POST("/tasks", CreateTask)
	router.GET("/tasks", GetTasks)
	router.GET("/tasks/:id", GetSpecificTask)
	router.PUT("/tasks/:id", UpdateTask)
	router.DELETE("/tasks/:id", DeleteTask)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Error loading env")
	}
}
func main() {
	router = gin.Default()
	_, err := SetUpDatabase()
	if err != nil {
		fmt.Print(err)
		return
	}
	SetupRoutes()

	go func() {
		if err := router.Run(port); err != nil {
			log.Printf("server stopped: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if Client != nil {
		if err := Client.Disconnect(ctx); err != nil {
			log.Println("mongo disconnect error:", err)
		}
	}
	log.Println("shutting down")
}
