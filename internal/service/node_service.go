package service

import (
	"errors"
	"net/http"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type NodeService struct{}

func NewNodeService() *NodeService {
	return &NodeService{}
}

type NodeListResult struct {
	Items    []model.Node `json:"items"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

func (s *NodeService) Create(name, url string, weight int) (*model.Node, error) {
	if name == "" || url == "" {
		return nil, errors.New("名称和URL不能为空")
	}
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	node := &model.Node{Name: name, URL: url, Weight: weight, Status: model.NodeOffline}
	if err := database.DB.Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func (s *NodeService) Update(id uint, name, url string, weight, status int) (*model.Node, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, errors.New("节点不存在")
	}
	if name != "" {
		node.Name = name
	}
	if url != "" {
		node.URL = url
	}
	node.Weight = weight
	node.Status = status
	if err := database.DB.Save(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *NodeService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.Node{}, id).Error
}

func (s *NodeService) Ping(id uint) (*model.Node, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var node model.Node
	if err := database.DB.First(&node, id).Error; err != nil {
		return nil, errors.New("节点不存在")
	}

	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(node.URL)
	pingTime := time.Since(start).Milliseconds()

	now := time.Now()
	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		node.Status = model.NodeOffline
	} else {
		node.Status = model.NodeOnline
	}
	node.LastPing = &now
	node.PingTime = pingTime
	if err := database.DB.Save(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *NodeService) List(page, pageSize int, keyword string) (*NodeListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &NodeListResult{Items: []model.Node{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.Node
	var total int64
	query := database.DB.Model(&model.Node{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR url LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &NodeListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *NodeService) GetBestNode() (*model.Node, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var node model.Node
	err := database.DB.Where("status = ?", model.NodeOnline).
		Order("weight DESC, ping_time ASC").
		First(&node).Error
	if err != nil {
		return nil, errors.New("没有可用节点")
	}
	return &node, nil
}
