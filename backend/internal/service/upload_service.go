package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

const (
	uploadDir  = "./uploads"
	maxFileLen = 10 * 1024 * 1024
)

var allowedImages = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
}

var allowedDocuments = map[string]bool{
	".pdf":  true,
	".txt":  true,
}

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (s *UploadService) SaveFile(filename string, data []byte, mimeType string, uploaderID uint) (*model.FileUpload, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	fileType := s.determineFileType(ext)

	if fileType == "" {
		return nil, errors.New("不支持的文件类型")
	}

	if len(data) > maxFileLen {
		return nil, errors.New("文件大小超出限制(10MB)")
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, err
	}
	storedName := hex.EncodeToString(randBytes) + ext

	savePath := filepath.Join(uploadDir, storedName)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	upload := &model.FileUpload{
		OriginalName: filename,
		StoredName:   storedName,
		Path:         savePath,
		Size:         int64(len(data)),
		MimeType:     mimeType,
		Type:         fileType,
		UploaderID:   uploaderID,
		CreatedAt:    time.Now(),
	}

	if err := database.DB.Create(upload).Error; err != nil {
		os.Remove(savePath)
		return nil, err
	}

	return upload, nil
}

func (s *UploadService) GetFile(id uint) (*model.FileUpload, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var upload model.FileUpload
	if err := database.DB.First(&upload, id).Error; err != nil {
		return nil, errors.New("文件不存在")
	}
	return &upload, nil
}

func (s *UploadService) DeleteFile(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}

	var upload model.FileUpload
	if err := database.DB.First(&upload, id).Error; err != nil {
		return errors.New("文件不存在")
	}

	if upload.Path != "" {
		os.Remove(upload.Path)
	}

	return database.DB.Delete(&upload).Error
}

func (s *UploadService) ListFiles(page, pageSize int, fileType string) (map[string]interface{}, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var files []model.FileUpload
	var total int64

	query := database.DB.Model(&model.FileUpload{})
	if fileType != "" {
		query = query.Where("type = ?", fileType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"files":     files,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}

func (s *UploadService) determineFileType(ext string) string {
	if allowedImages[ext] {
		return model.FileTypeImage
	}
	if allowedDocuments[ext] {
		return model.FileTypeFile
	}
	return ""
}