package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
	Completed bool               `json:"completed" bson:"completed"`
}

var router *gin.Engine

var Client *mongo.Client

func SetUpDatabase() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	Client = client
	return client
}

func throwError(err error, c *gin.Context) {
	c.JSON(400, gin.H{
		"error": err.Error(),
	})
}

func CreateTask(c *gin.Context) {
	var task Task
	err := c.BindJSON(&task)

	_, err = Client.Database(databaseName).Collection(collectionName).InsertOne(context.Background(), task)

	if err != nil {
		throwError(err, c)
		return
	}

	c.JSON(200, gin.H{
		"message": "Task created",
	})

}

func GetTasks(c *gin.Context) {
	var tasks []bson.M
	cursor, err := Client.Database(databaseName).Collection(collectionName).Find(context.Background(), bson.D{})
	if err != nil {
		throwError(err, c)
		return
	}
	err = cursor.All(context.Background(), &tasks)
	if err != nil {
		throwError(err, c)
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
		throwError(err, c)
		return
	}
	var task Task
	err = Client.Database(databaseName).Collection(collectionName).FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&task)
	if err != nil {
		throwError(err, c)
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
		throwError(err, c)
		return
	}

	var task Task
	err = c.BindJSON(&task)
	if err != nil {
		throwError(err, c)
		return
	}
	_, err = Client.Database(databaseName).Collection(collectionName).UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": task})
	if err != nil {
		throwError(err, c)
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
		throwError(err, c)
		return
	}
	_, err = Client.Database(databaseName).Collection(collectionName).DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		throwError(err, c)
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
	SetUpDatabase()
	SetupRoutes()

	router.Run(port)

}
