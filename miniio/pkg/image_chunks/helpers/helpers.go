package helpers

import (
	chunk_models "ImageUploadMiniIo/pkg/image_chunks/models"
	redis_database "ImageUploadMiniIo/pkg/redis"
	redis_models "ImageUploadMiniIo/pkg/redis/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// Function to get the total chunks from the redis for a particular session id.
func GetTotalChunks(sessionId string) (*int, error) {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Get the session data and check whether session has expired or not.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session has expired")
	} else if err != nil {
		return nil, err
	}

	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return nil, err
	}

	return &sessionData.TotalChunks, nil
}

// Function to update the chunk received list.
func UpdateReceivedIdSet(c *gin.Context, sessionId string) (mapset.Set[int], error) {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Check if the session exists or not.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session has expired")
	} else if err != nil {
		return nil, err
	}

	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return nil, err
	}

	// Get the chunk details from the client request.
	chunkDetails, err := GetChunkDetails(c)
	if err != nil {
		return nil, err
	}

	// Update the received chunk numbers.
	sessionData.ReceivedIds.Add(chunkDetails.ChunkNumber)

	// Get the TTL.
	ttl, err := redisClient.Client.TTL(redisClient.Ctx, sessionId).Result()
	if err != nil {
		return nil, err
	}

	// Set the updated list.
	jsonData, err = redisClient.Client.Set(redisClient.Ctx, sessionId, sessionData, ttl).Result()
	if err != nil {
		return nil, err
	}

	// Unmarshling the received data.
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return nil, err
	}

	return sessionData.ReceivedIds, nil
}

// Function to get the file details from the client request and save in the models.FileDetails struct.
func GetChunkDetails(c *gin.Context) (*chunk_models.RequestData, error) {
	// Get the file details from the client request.
	var requestData chunk_models.RequestData
	err := c.BindJSON(&requestData)
	if err != nil {
		return nil, err
	}

	return &requestData, nil
}

// Function to delete a session id from redis, if it exists.
func DeleteSessionIfExists(redisClient *redis_models.RedisClient, compositeKey string) (string, error) {
	// Search the value using the composite key if exists and delete it by returning the value.
	value, err := redisClient.Client.Do(redisClient.Ctx, "GETDEL", compositeKey).Result()
	if err != nil {
		return "", err
	}

	// If there were no sessions corresponding to that ip address and user agent, return null string and null error.
	if value == nil {
		return "", nil
	}

	// If it exists, convert it to string and deserialise from JSON to a pre-defined struct and return the session id.

	// Converting it into string.
	deserialisedData, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected value type: %T", deserialisedData)
	}

	// Deserialising from JSON to pre-defined struct.
	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(deserialisedData), &sessionData)
	if err != nil {
		return "", err
	}

	return sessionData.SessionId, nil
}

// Function to delete the temporary hidden folder containing the chunks for a particular session.
func DeleteTempFolderPaths(sessionId string) error {
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
func DeletePermFolderPaths(sessionId string) error {
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

// Function to create a cookie for the client session.
func CreateCookie(c *gin.Context) (*http.Cookie, error) {
	// Search if any session already exists in the system with the same user agent and ip address.
	// If yes delete the corresponding entry and corresponding temporary chunk folder and full file location
	// from the system with the help of the session id present in the redis for that user.

	// Getting the redis client.
	redisClient := redis_database.GetRedisClient()

	// Getting the ip address and user agent.
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Creating the composite key.
	compositeKey := ipAddress + ":" + userAgent

	// Deleting the session if exists.
	deletedSessionId, err := DeleteSessionIfExists(redisClient, compositeKey)
	if err != nil {
		return nil, err
	}

	// If session already existed, deleting the existing folders, if exists.
	if deletedSessionId != "" {
		err = DeleteTempFolderPaths(deletedSessionId)
		if err != nil {
			return nil, err
		}

		err = DeletePermFolderPaths(deletedSessionId)
		if err != nil {
			return nil, err
		}
	}

	// Then take the file details from the client request, session id, user agent, ip address, creation time,
	// expiry time and failed chunk ids and save it in the redis.

	// Create new session id.
	sessionId := uuid.NewString()

	// Create a cookie.
	currentTime := time.Now()
	cookie := &http.Cookie{
		Name:     "sessionId",
		Value:    sessionId,
		Expires:  currentTime.Add(time.Hour),
		HttpOnly: true,
	}

	// Create the session data to be stored in redis.

	// Getting the data passed in the client request.
	requestData, err := GetChunkDetails(c)
	if err != nil {
		return nil, err
	}

	var fileDetails chunk_models.FileDetails
	fileDetails.FileName = requestData.FileName
	fileDetails.FileSize = requestData.FileSize
	fileDetails.FileSizeUnit = requestData.FileSizeUnit
	fileDetails.FileType = requestData.FileType
	fileDetails.TotalChunks = requestData.TotalChunks

	// Setting the data for session data.
	var sessionData chunk_models.SessionData
	sessionData.SessionId = sessionId
	sessionData.IPAddress = ipAddress
	sessionData.UserAgent = userAgent
	sessionData.FileDetails = fileDetails
	sessionData.CreationTime = currentTime
	sessionData.ExpiryTime = currentTime.Add(time.Hour)
	sessionData.FailedChunksInfo = make([]int, 0)
	sessionData.ReceivedIds = mapset.NewSet[int]()

	// Write this data to the redis.
	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return nil, err
	}

	err = redisClient.Client.Set(redisClient.Ctx, sessionId, jsonData, 24*time.Hour).Err()
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

// Function to validate the session, whether it exists and is not expired stored in redis.
func ValidateSession(sessionId string) (bool, error) {
	// Getting the redis client.
	redisClient := redis_database.GetRedisClient()

	// Search in the redis with the sessionId as the key if it exists or not.
	// If exists returns true, otherwise, return false.
	result, err := redisClient.Client.Exists(redisClient.Ctx, sessionId).Result()
	if err != nil {
		return false, err
	}

	// In the above if case, if it is already an error, then session validation failed, otherwise, it either exists or not.
	// Result = 0 --> does not exist & Result = 1 --> exists.
	return result == 1, nil
}

// Function to create a temporary hidden folder for a particular session to save the uploaded chunks.
func createTempFolder(sessionId string) (string, error) {
	// Get the folder path.
	folderTempPath := os.Getenv("FOLDER_TEMP_PATH")

	// Updating the folder path specific to sessionId.
	folderTempPath = filepath.Join(folderTempPath, "."+sessionId)

	// Check whether the path already exists or not.
	// If no, create the folder and then return.
	if _, err := os.Stat(folderTempPath); os.IsNotExist(err) {
		err := os.Mkdir(folderTempPath, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return folderTempPath, nil
}

// Function to save the chunks in the temporary location for a particular session.
func saveChunkTempLocation(c *gin.Context, sessionId string, folderPath string, chunkDetails *chunk_models.RequestData) error {
	// Make the file name of form sessionId + chunk number form and make the chukn file path.
	fileName := fmt.Sprintf("%s_%d.%s", sessionId, chunkDetails.ChunkNumber, chunkDetails.FileType)
	filePath := filepath.Join(folderPath, fileName)

	// Get the file from the request.
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// Saving the file.
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return err
	}

	return nil
}

// Function to help in different processes of uploading the chunks for a particular session.
func UploadChunkHelper(c *gin.Context, sessionId string) (*int, error) {
	// Check whether the file location already exists or not. If not then make one.
	folderPath, err := createTempFolder(sessionId)
	if err != nil {
		return nil, err
	}

	// Get the chunk details from the client request.
	chunkDetails, err := GetChunkDetails(c)
	if err != nil {
		return nil, err
	}

	// Temporarily save the chunk in the location.
	err = saveChunkTempLocation(c, sessionId, folderPath, chunkDetails)
	if err != nil {
		return nil, err
	}

	return &chunkDetails.ChunkNumber, nil
}

// Function to update the redis failed list for a particular session, if any chunk upload activity fails.
func UpdateRedisFailedList(c *gin.Context, sessionId string, failedChunkNumber int) error {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Check if the session Id exists or it has expired.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return fmt.Errorf("session expired")
	} else if err != nil {
		return err
	}

	// The key exists and now you need to update the failed chunk number list.
	// Deserialise the json data into the models.Session structure.
	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return err
	}

	// Updating the failed chunk number list for the session id.
	sessionData.FailedChunksInfo = append(sessionData.FailedChunksInfo, failedChunkNumber)

	return nil
}

// Function to check whether for a particular session any chunk upload failed or not.
func CheckFailStatus(sessionId string) ([]int, error) {
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()

	// Get the session data.
	jsonData, err := redisClient.Client.Get(redisClient.Ctx, sessionId).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session expired")
	} else if err != nil {
		return nil, err
	}

	// The key exists and now you need to update the failed chunk number list.
	// Deserialise the json data into the models.Session structure.
	var sessionData chunk_models.SessionData
	err = json.Unmarshal([]byte(jsonData), &sessionData)
	if err != nil {
		return nil, err
	}

	// Check if the failed list is empty or not.
	// If the failed list is not empty return the whole list.
	if len(sessionData.FailedChunksInfo) != 0 {
		return sessionData.FailedChunksInfo, nil
	}

	return nil, nil
}

// Function to delete all information for a particular session id.
func DeleteAllForSession(sessionId string) []error {
	var errors []error

	// Deleting the permanent folder and its contents created for a particular session.
	err := DeletePermFolderPaths(sessionId)
	if err != nil {
		errors = append(errors, err)
	}

	// Deleting the temporary folder and its contents created for a particular session.
	err = DeleteTempFolderPaths(sessionId)
	if err != nil {
		errors = append(errors, err)
	}

	// Deleteing the session id details from the redis storage for a particular session.
	// Get the redis client.
	redisClient := redis_database.GetRedisClient()
	_, err = DeleteSessionIfExists(redisClient, sessionId)
	if err != nil {
		errors = append(errors, err)
	}

	return errors

}
