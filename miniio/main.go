package main

import (
	chunk_redis "ImageUploadMiniIo/pkg/redis"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Loading environment variable.
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error: Environment file could not be loaded.")
		os.Exit(1)
	}

	// Getting the chunk api port.
	chunkPort := os.Getenv("CHUNK_PORT")

	// Declaring a gin server.
	chunkRouter := gin.New()

	// Staring the gin server.
	chunkRouter.Run(":", chunkPort)

	// Getting the redis client.
	redisClient := chunk_redis.GetRedisClient()

	// Creating a channel to receive OS signals.
	sigs := make(chan os.Signal, 1)
	log.Println("Message: Application Started.")

	// Block the main routine until a signal is received.
	<-sigs

	// Once a signal is received, call shutdown to clean up resources.
	redisClient.ShutDown()
	log.Println("Message: Application Stopped.")
}
