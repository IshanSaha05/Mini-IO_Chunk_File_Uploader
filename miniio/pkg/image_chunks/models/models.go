package models

import (
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

type RequestData struct {
	FileName      string `json:"file_name"`
	FileType      string `json:"file_type"`
	FileSizeUnit  string `json:"file_size_unit"`
	FileSize      int    `json:"file_size"`
	TotalChunks   int    `json:"total_chunks"`
	ChunkNumber   int    `json:"chunk_number"`
	CompileStatus bool   `json:"compile_status"`
}

type FileDetails struct {
	FileName     string `json:"file_name"`
	FileType     string `json:"file_type"`
	FileSizeUnit string `json:"file_size_unit"`
	FileSize     int    `json:"file_size"`
	TotalChunks  int    `json:"total_chunks"`
}

type SessionData struct {
	SessionId        string `json:"session_id"`
	IPAddress        string `json:"ip_address"`
	UserAgent        string `json:"user_agent"`
	FileDetails      `json:"file_details"`
	CreationTime     time.Time       `json:"creation_time"`
	ExpiryTime       time.Time       `json:"expiry_time"`
	FailedChunksInfo []int           `json:"failed_chunks_info"`
	ReceivedIds      mapset.Set[int] `json:"received_ids"`
}
