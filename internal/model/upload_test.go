package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUpload_TableName(t *testing.T) {
	assert.Equal(t, "file_uploads", FileUpload{}.TableName())
}

func TestFileUpload_TableName_NotEmpty(t *testing.T) {
	name := FileUpload{}.TableName()
	assert.NotEmpty(t, name)
	assert.IsType(t, "", name)
}

func TestFileUpload_FileTypeConstants(t *testing.T) {
	assert.Equal(t, "image", FileTypeImage)
	assert.Equal(t, "file", FileTypeFile)
}

func TestFileUpload_StructDefaults(t *testing.T) {
	f := FileUpload{}
	assert.Equal(t, uint(0), f.ID)
	assert.Equal(t, "", f.OriginalName)
	assert.Equal(t, "", f.StoredName)
	assert.Equal(t, "", f.Path)
	assert.Equal(t, int64(0), f.Size)
	assert.Equal(t, "", f.MimeType)
	assert.Equal(t, "", f.Type)
	assert.Equal(t, uint(0), f.UploaderID)
}

func TestFileUpload_StructAssignment(t *testing.T) {
	f := FileUpload{
		ID:           1,
		OriginalName: "test.jpg",
		StoredName:   "abc123.jpg",
		Path:         "./uploads/abc123.jpg",
		Size:         1024,
		MimeType:     "image/jpeg",
		Type:         FileTypeImage,
		UploaderID:   42,
	}

	assert.Equal(t, uint(1), f.ID)
	assert.Equal(t, "test.jpg", f.OriginalName)
	assert.Equal(t, "abc123.jpg", f.StoredName)
	assert.Equal(t, "./uploads/abc123.jpg", f.Path)
	assert.Equal(t, int64(1024), f.Size)
	assert.Equal(t, "image/jpeg", f.MimeType)
	assert.Equal(t, FileTypeImage, f.Type)
	assert.Equal(t, uint(42), f.UploaderID)
}