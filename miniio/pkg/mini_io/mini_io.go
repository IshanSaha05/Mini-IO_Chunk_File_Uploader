package miniio

import (
	"log"
	"os"
	"strconv"

	miniio_models "ImageUploadMiniIo/pkg/mini_io/models"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go"
)

// Declaring a mini-io client variable.
var miniIoClient miniio_models.MiniIoClient

func init() {
	// Loading the redis address and password environment variables.
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error: Problem while loading environment variables.")
		os.Exit(1)
	}

	// Getting all the parameters from the environment file.
	endPoint := os.Getenv("MINIIO_ENDPOINT")
	accessKeyId := os.Getenv("MINIIO_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("MINIIO_SECRET_ACCESS_KEY")
	useSSLString := os.Getenv("MINIIO_USESSL")
	bucketName := os.Getenv("MINIIO_BUCKET_NAME")
	location := os.Getenv("MINIIO_LOCATION")

	// Convert useSSL value to bool.
	useSSL, err := strconv.ParseBool(useSSLString)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
		os.Exit(1)
	}

	miniIoClient.Once.Do(func() {
		// Set new mini-io client.
		miniIoClient.Client, err = minio.New(endPoint, accessKeyId, secretAccessKey, useSSL)
		if err != nil {
			log.Fatalf("Error: Problem while connecting to Mini-Io client.       %s", err.Error())
			os.Exit(1)
		}

		// Set the bucket name and location to the structure.
		miniIoClient.BucketName = bucketName
		miniIoClient.Location = location

		log.Println("Message: Connected to Minio-Io client successfully.")
	})

}
