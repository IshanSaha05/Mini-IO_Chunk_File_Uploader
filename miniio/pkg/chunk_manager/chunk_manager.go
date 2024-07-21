package chunk_manager

import (
	chunk_models "ImageUploadMiniIo/pkg/image_chunks/models"
	redis_database "ImageUploadMiniIo/pkg/redis"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func getChunkDetails(sessionId string) (*chunk_models.FileDetails, error) {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Check if the session Id exists or it has expired.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session expired")
	} else if err != nil {
		return nil, err
	}

	// Deserialise the json data into the models.Session structure.
	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData.FileDetails, nil
}

func createPermFolder(sessionId string) (string, error) {
	// Get the folder path.
	folderPermPath := os.Getenv("FOLDER_PERM_PATH")

	// Updating the folder path specific to sessionId.
	folderPermPath = filepath.Join(folderPermPath, sessionId)

	// Check whether the path already exists or not.
	// If no, create the folder and then return.
	if _, err := os.Stat(folderPermPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPermPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return folderPermPath, nil
}

func getTempFolderPath(sessionId string) (string, error) {
	// Get the environment variables.
	tempFolderPath := os.Getenv("FOLDER_TEMP_PATH")

	// Getting the folder path for the corresponding session id.
	tempFolderPath = filepath.Join(tempFolderPath, "."+sessionId)

	// Checking whether the folder exists or not.
	if _, err := os.Stat(tempFolderPath); os.IsNotExist(err) {
		return "", err
	}

	return tempFolderPath, nil
}

func saveChunkPermLocation(sessionId string, permFolderPathh string, tempFolderPath string, chunkDetails *chunk_models.FileDetails) error {
	// Make the file name for permanent file.
	fileName := fmt.Sprintf("%s.%s", sessionId, chunkDetails.FileType)
	filePermPath := filepath.Join(permFolderPathh, "/"+fileName)

	// Open the final file for writing.
	permFile, err := os.Create(filePermPath)
	if err != nil {
		return err
	}
	defer permFile.Close()

	// Reading each chunk file and writing it to the output file.
	for i := 1; i <= chunkDetails.TotalChunks; i++ {
		// Getting the chunk file name and path.
		chunkFileName := fmt.Sprintf("%s_%d.%s", sessionId, i, chunkDetails.FileType)
		chunkFilePath := filepath.Join(tempFolderPath, chunkFileName)

		// Reading the chunk file.
		chunkData, err := os.ReadFile(chunkFilePath)
		if err != nil {
			return err
		}

		// Writing the chunk file into the final file.
		_, err = permFile.Write(chunkData)
		if err != nil {
			return err
		}
	}

	return nil
}

func CompileChunks(c *gin.Context, sessionId string) error {
	// Create the permanent folder.
	permFolderPath, err := createPermFolder(sessionId)
	if err != nil {
		return err
	}

	// Get the chunk details from the client request.
	chunkDetails, err := getChunkDetails(sessionId)
	if err != nil {
		return err
	}

	// Permanently save the full file in the location.
	// Get the temp folder path.
	tempFolderPath, err := getTempFolderPath(sessionId)
	if err != nil {
		return err
	}
	err = saveChunkPermLocation(sessionId, permFolderPath, tempFolderPath, chunkDetails)
	if err != nil {
		return err
	}

	return nil
}
