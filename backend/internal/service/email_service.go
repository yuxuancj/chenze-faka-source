package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

type EmailListResult struct {
	Items    []model.EmailConfig `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

func (s *EmailService) Create(cfg *model.EmailConfig) (*model.EmailConfig, error) {
	if cfg.SMTPHost == "" || cfg.Username == "" {
		return nil, errors.New("SMTP配置不完整")
	}
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 465
	}
	if err := database.DB.Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *EmailService) Update(id uint, cfg *model.EmailConfig) (*model.EmailConfig, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var existing model.EmailConfig
	if err := database.DB.First(&existing, id).Error; err != nil {
		return nil, errors.New("配置不存在")
	}
	if cfg.SMTPHost != "" {
		existing.SMTPHost = cfg.SMTPHost
	}
	if cfg.SMTPPort > 0 {
		existing.SMTPPort = cfg.SMTPPort
	}
	if cfg.Username != "" {
		existing.Username = cfg.Username
	}
	if cfg.Password != "" {
		existing.Password = cfg.Password
	}
	if cfg.Sender != "" {
		existing.Sender = cfg.Sender
	}
	existing.UseSSL = cfg.UseSSL
	existing.Status = cfg.Status
	if err := database.DB.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *EmailService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.EmailConfig{}, id).Error
}

func (s *EmailService) List(page, pageSize int) (*EmailListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &EmailListResult{Items: []model.EmailConfig{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.EmailConfig
	var total int64
	if err := database.DB.Model(&model.EmailConfig{}).Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := database.DB.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &EmailListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *EmailService) TestConnection(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	var cfg model.EmailConfig
	if err := database.DB.First(&cfg, id).Error; err != nil {
		return errors.New("配置不存在")
	}
	return s.sendTestEmail(&cfg)
}

func (s *EmailService) sendTestEmail(cfg *model.EmailConfig) error {
	subject := "晨泽发卡系统 - 邮件测试"
	body := "这是一封来自晨泽发卡系统的测试邮件，如果你收到此邮件，说明SMTP配置正确。"

	return s.send(cfg, []string{cfg.Username}, subject, body, "")
}

func (s *EmailService) SendCardEmail(recipients []string, subject string, body string, relatedNo string) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	var cfg model.EmailConfig
	if err := database.DB.Where("status = ?", 1).First(&cfg).Error; err != nil {
		return errors.New("没有可用的邮件配置")
	}

	err := s.send(&cfg, recipients, subject, body, relatedNo)
	status := model.EmailLogSent
	errMsg := ""
	if err != nil {
		status = model.EmailLogFailed
		errMsg = err.Error()
	}

	for _, r := range recipients {
		log := &model.EmailLog{
			ToEmail:   r,
			Subject:   subject,
			Content:   body,
			Status:    status,
			ErrorMsg:  errMsg,
			RelatedNo: relatedNo,
		}
		database.DB.Create(log)
	}
	return err
}

func (s *EmailService) send(cfg *model.EmailConfig, recipients []string, subject string, body string, relatedNo string) error {
	from := cfg.Sender
	if from == "" {
		from = cfg.Username
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, strings.Join(recipients, ","), subject, body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	if cfg.UseSSL {
		tlsConfig := &tls.Config{
			ServerName: cfg.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return err
		}
		if err := client.Mail(from); err != nil {
			return err
		}
		for _, r := range recipients {
			if err := client.Rcpt(r); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, from, recipients, []byte(msg))
}
