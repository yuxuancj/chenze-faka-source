package service

import (
	"errors"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

const currentVersion = "1.0.0"

type UpgradeService struct{}

func NewUpgradeService() *UpgradeService {
	return &UpgradeService{}
}

func (s *UpgradeService) GetVersion() map[string]interface{} {
	return map[string]interface{}{
		"version":     currentVersion,
		"name":        "Chenze Faka",
		"description": "自动发卡系统",
		"check_url":   "https://example.com/api/upgrade/check",
	}
}

func (s *UpgradeService) CheckUpdate() map[string]interface{} {
	return map[string]interface{}{
		"has_update":    false,
		"current":       currentVersion,
		"latest":        currentVersion,
		"download_url":  "",
		"release_notes": "",
	}
}

func (s *UpgradeService) UploadPackage(filename string, data []byte) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}

	if filename == "" {
		return errors.New("文件名不能为空")
	}

	log := &model.UpgradeLog{
		Version:     currentVersion,
		Description: "上传升级包: " + filename,
		Status:      model.UpgradeStatusPending,
		FileName:    filename,
		Size:        int64(len(data)),
		Operator:    "system",
	}

	return database.DB.Create(log).Error
}

func (s *UpgradeService) ApplyUpgrade() error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}

	var log model.UpgradeLog
	if err := database.DB.Where("status = ?", model.UpgradeStatusPending).
		Order("id DESC").First(&log).Error; err != nil {
		return errors.New("没有待升级的包")
	}

	now := time.Now()
	log.Status = model.UpgradeStatusSuccess
	log.UpdatedAt = now

	return database.DB.Save(&log).Error
}

func (s *UpgradeService) ListUpgradeLogs(page, pageSize int) (map[string]interface{}, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var logs []model.UpgradeLog
	var total int64

	if err := database.DB.Model(&model.UpgradeLog{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := database.DB.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, nil
}