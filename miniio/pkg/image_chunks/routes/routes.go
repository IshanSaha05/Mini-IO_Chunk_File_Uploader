package routes

import (
	chunk_controller "ImageUploadMiniIo/pkg/image_chunks/controllers"
	chunk_middleware "ImageUploadMiniIo/pkg/image_chunks/middleware"

	"github.com/gin-gonic/gin"
)

func ChunkRoutes(chunkRouter *gin.Engine) {
	chunkRouter.Use(chunk_middleware.Authenticate())
	chunkRouter.POST("/api/v1/upload_chunk", chunk_controller.UploadChunks())
}
