package redis

import (
	redis_models "ImageUploadMiniIo/pkg/redis/models"
	"log"
	"os"
)

// Function to get the redis client.
func GetRedisClient() *redis_models.RedisClient {
	return &redisClient
}

// Function to delete all the temporary folders for a particular session id.
func deleteTempFolderPaths(sessionId string) error {
	// Loading the environment variables.
	tempFolderPath := os.Getenv("FOLDER_TEMP_PATH")

	// Updating the folder paths by appending the session id.
	tempFolderPath = tempFolderPath + "/" + sessionId

	// Check if the folders exists.
	if _, err := os.Stat(tempFolderPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Remove the folders and its contents.
	err := os.RemoveAll(tempFolderPath)
	if err != nil {
		return err
	}

	return nil
}

// Function to delete the permanent folder containing the whole compiled file.
func deletePermFolderPaths(sessionId string) error {
	// Loading the environment variables.
	permFolderPath := os.Getenv("FOLDER_PERM_PATH")

	// Updating the folder paths by appending the session id.
	permFolderPath = permFolderPath + "/" + sessionId

	// Check if the folders exists.
	if _, err := os.Stat(permFolderPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	// Remove the folders and its contents.
	err := os.RemoveAll(permFolderPath)
	if err != nil {
		return err
	}

	return nil
}

// Function to handle the function which is to be done, when a session gets expired and deleted from the redis.
func handleExpiredKey(sessionId string) error {
	err := deletePermFolderPaths(sessionId)
	if err != nil {
		return err
	}

	err = deleteTempFolderPaths(sessionId)
	if err != nil {
		return err
	}

	return nil
}

// Function to handle the redis expired channel messages.
func handleMessages() {
	defer redisClient.Wg.Done()

	for msg := range redisClient.ExpireChannel.Channel() {
		log.Printf("Message: Session with id \"%s\" expired.\n         Deleting all folders if exists.", msg.Payload)
		err := handleExpiredKey(msg.Payload)
		if err != nil {
			log.Fatalf("Error: %s", err.Error())
		}
	}
}
