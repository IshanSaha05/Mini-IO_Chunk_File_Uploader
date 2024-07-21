package miniio

import (
	chunk_models "ImageUploadMiniIo/pkg/image_chunks/models"
	miniio_models "ImageUploadMiniIo/pkg/mini_io/models"
	redis_database "ImageUploadMiniIo/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go"
)

// Function to get the minio-io client.
func GetMiniIoClient() *miniio_models.MiniIoClient {
	return &miniIoClient
}

// Function to upload files to mini-io bucket.
func UploadSessionFilesToMiniIoBucket(sessionId string) error {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Get the chunk details for the particular session id.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return fmt.Errorf("session has expired")
	} else if err != nil {
		return err
	}

	// Unmarshal the jsonData.
	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return err
	}

	// Get the metadata for the object.
	var metaData miniio_models.Metadata
	metaData.SessionId = sessionData.SessionId
	metaData.IPAddress = sessionData.SessionId
	metaData.UserAgent = sessionData.UserAgent
	metaData.FileDetails = sessionData.FileDetails
	metaData.CreationTime = time.Now()

	// Get the permanent folder location for the session id.
	folderPermPath := os.Getenv("FOLDER_PERM_PATH")
	folderPermPath = filepath.Join(folderPermPath, sessionId)
	fileName := fmt.Sprintf("%s.%s", sessionId, sessionData.FileDetails.FileType)
	filePermPath := filepath.Join(folderPermPath, "/"+fileName)

	// Get the content type.
	contentType := mime.TypeByExtension(sessionData.FileDetails.FileType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set the object name.
	objectName := sessionId

	// Set metadata for the object.
	// Serialise the metadata into json.
	metaDataJson, err := json.Marshal(metaData)
	if err != nil {
		return err
	}
	// Convert json to string.
	metaDataMap := map[string]string{
		fmt.Sprintf("%s_metadata", sessionId): string(metaDataJson),
	}

	// Upload the file.
	_, err = miniIoClient.Client.FPutObjectWithContext(context.Background(), miniIoClient.BucketName, objectName, filePermPath, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metaDataMap,
	})
	if err != nil {
		return err
	}

	log.Printf("Message: Successfully uploaded object in Mini-Io server with session id \"%s\"", sessionId)

	return nil
}
