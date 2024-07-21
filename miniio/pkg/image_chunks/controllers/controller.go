package controllers

import (
	"ImageUploadMiniIo/pkg/chunk_manager"
	chunk_helpers "ImageUploadMiniIo/pkg/image_chunks/helpers"
	miniio "ImageUploadMiniIo/pkg/mini_io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Upload controller.
func UploadChunks() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if session id has been passed from the middleware and extract it.
		sessionIdVal, exists := c.Get("sessionId")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error."})
			c.Abort()
			return
		}
		sessionId, ok := sessionIdVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error."})
			c.Abort()
			return
		}

		// Check whether the session id is empty or not.
		// If yes creates a new session and get the cookie which is further added in the response from the server side.
		if sessionId == "" {
			cookie, err := chunk_helpers.CreateCookie(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "error_details": err.Error()})
				c.Abort()
				return
			}

			http.SetCookie(c.Writer, cookie)

			// Retrieving the session id from the created cookie.
			sessionId, err = c.Cookie(cookie.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "error_details": err.Error()})
				c.Abort()
				return
			}
		}

		// Update the send list in the redis.
		receivedIdsSet, err := chunk_helpers.UpdateReceivedIdSet(c, sessionId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_details": err.Error()})
			c.Abort()
			return
		}

		// Upload the chunk and check the status if it has succeeded or failed.
		chunkNumber, err := chunk_helpers.UploadChunkHelper(c, sessionId)

		// If upload is unsuccessful, update the redis status unsuccessful list.
		if err != nil {
			chunk_helpers.UpdateRedisFailedList(c, sessionId, *chunkNumber)
		}

		// Check first whether all the chunks have been forwared or not.
		// If yes, check whether all the chunks have received or not.
		// If no, then get the all the chunks which has failed to upload & send those chunk numbers as a response to client.
		// Get the client request data.
		requestData, err := chunk_helpers.GetChunkDetails(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_details": err.Error()})
			c.Abort()
			return
		}

		// Get the total chunk number saved in redis for the particular session.
		totalChunksPointer, err := chunk_helpers.GetTotalChunks(sessionId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_details": err.Error()})
			c.Abort()
			return
		}

		// Check whether we can initiate the compilation process of the chunks for a particular session id.
		totalChunks := *totalChunksPointer
		if requestData.CompileStatus && receivedIdsSet.Cardinality() == totalChunks {
			// Check whether any of chunks have failed or not.
			// If yes then end the process for the paritcular session id or else go with compiling the chunks.
			failedList, err := chunk_helpers.CheckFailStatus(sessionId)
			if failedList != nil {
				response := gin.H{"error": "Few chunks have failed.", "failed_chunk_list": failedList}

				// Delete redis row item, temp folder and perm folder for that session id.
				errors := chunk_helpers.DeleteAllForSession(sessionId)
				if len(errors) > 0 {
					response["error_list"] = errors
				}

				c.JSON(http.StatusPartialContent, response)
				c.Abort()
				return
			} else if err != nil {
				response := gin.H{"error": "Internal server error.", "error_details": err.Error()}

				// Delete redis row item, temp folder and perm folder for that session id.
				errors := chunk_helpers.DeleteAllForSession(sessionId)
				if len(errors) > 0 {
					response["error_list"] = errors
				}

				c.JSON(http.StatusBadRequest, response)
				c.Abort()
				return
			}

			// If all the chunks have been successfully uploaded, call to chunk manager to initiate the process
			// of merging the chunks and saving it as a whole file.
			// And also, send in the server response file uploaded successfully.
			err = chunk_manager.CompileChunks(c, sessionId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_details": err.Error()})
				c.Abort()
				return
			}

			// Run the mini-io and transfer the files into s3 buckets.
			err = miniio.UploadSessionFilesToMiniIoBucket(sessionId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_details": err.Error()})
				c.Abort()
				return
			}

			// Delete redis row item, temp folder and perm folder for that session id.
			errors := chunk_helpers.DeleteAllForSession(sessionId)
			if len(errors) > 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error.", "error_list": errors})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "File successfully uploaded."})
		}
	}
}
