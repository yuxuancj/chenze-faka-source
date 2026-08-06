package model

import (
	"time"
)

type FileUpload struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	OriginalName string    `json:"original_name" gorm:"size:255;not null"`
	StoredName   string    `json:"stored_name" gorm:"size:255;not null"`
	Path         string    `json:"path" gorm:"size:500;not null"`
	Size         int64     `json:"size" gorm:"default:0"`
	MimeType     string    `json:"mime_type" gorm:"size:100"`
	Type         string    `json:"type" gorm:"size:20;not null;default:file"`
	UploaderID   uint      `json:"uploader_id" gorm:"index"`
	CreatedAt    time.Time `json:"created_at"`
}

const (
	FileTypeImage = "image"
	FileTypeFile  = "file"
)

func (FileUpload) TableName() string {
	return "file_uploads"
}