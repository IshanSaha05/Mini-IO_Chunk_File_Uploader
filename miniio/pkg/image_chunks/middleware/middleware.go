package middleware

import (
	chunk_helpers "ImageUploadMiniIo/pkg/image_chunks/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check whether the session id has been passed inside the cookie or not.
		// If not set the empty string for session id and pass the control to the route handler.
		sessionId, err := c.Cookie("session_id")
		if err != nil || sessionId == "" {
			c.Set("sessionId", "")
		} else {
			// If session id is present inside the cookie, validate the session id.
			// If validation is not successful, return with a failure message.
			// If sessiond id does not exist, return an appropriate message.
			// If successful, set the session id and pass the control to the route handler.
			exists, err := chunk_helpers.ValidateSession(sessionId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error."})
				c.Abort()
				return
			}
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials."})
				c.Abort()
				return
			}

			c.Set("sessionId", sessionId)
		}

		c.Next()
	}
}
