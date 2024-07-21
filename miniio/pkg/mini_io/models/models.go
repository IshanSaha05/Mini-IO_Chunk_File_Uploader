package models

import (
	chunk_models "ImageUploadMiniIo/pkg/image_chunks/models"
	"context"
	"sync"
	"time"

	minio "github.com/minio/minio-go"
)

type MiniIoClient struct {
	Ctx   context.Context
	Cancel     context.CancelFunc
	Client     *minio.Client
	BucketName string
	Location   string
	Once       sync.Once
}

type Metadata struct {
	SessionId    string                   `json:"session_id"`
	IPAddress    string                   `json:"ip_address"`
	UserAgent    string                   `json:"user_agent"`
	FileDetails  chunk_models.FileDetails `json:"file_details"`
	CreationTime time.Time                `json:"creation_time"`
}
