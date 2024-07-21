package redis

import (
	redis_models "ImageUploadMiniIo/pkg/redis/models"
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// Declaring a RedisClient type declared inside the models.go file in model package.
var redisClient redis_models.RedisClient

// Init() function to initiate the redis client.
func init() {
	// Loading the redis address and password environment variables.
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error: Problem while loading environment variables.")
		os.Exit(1)
	}

	address := os.Getenv("REDIS_ADDRESS")
	password := os.Getenv("REDIS_PASSOWRD")

	// Once Do is used to make sure only one time this code within is executed in case of multi-threaded excution.
	redisClient.Once.Do(func() {
		// Set redis context.
		redisClient.Ctx, redisClient.Cancel = context.WithTimeout(context.Background(), 20*time.Second)

		// Set new redis client.
		redisClient.Client = redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       0,
		})

		// Subscribe to keyspace notifications for expired keys.
		redisClient.ExpireChannel = redisClient.Client.PSubscribe(redisClient.Ctx, "__keyevent@0__:expired")
		log.Println("Message: Subscribed to keyspace notifications.")

		// Starting a go routine to handle the incoming messages, which willl be used to delete the expired session folders.
		redisClient.Wg.Add(1)
		go handleMessages()

		// Ping the redis client.
		_, err := redisClient.Client.Ping(redisClient.Ctx).Result()
		if err != nil {
			log.Fatalf("Error: Problem while pinging the redis client.")
			os.Exit(1)
		}

		log.Println("Message: Pinging redis client successful.")
	})
}
