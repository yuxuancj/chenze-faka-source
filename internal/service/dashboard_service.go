package service

import (
	"errors"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type DashboardService struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

type DashboardStats struct {
	ProductCount    int64   `json:"product_count"`
	OrderCount      int64   `json:"order_count"`
	CardSoldCount   int64   `json:"card_sold_count"`
	TotalRevenue    float64 `json:"total_revenue"`
	TodayOrders     int64   `json:"today_orders"`
	TodayRevenue    float64 `json:"today_revenue"`
	MonthOrders     int64   `json:"month_orders"`
	MonthRevenue    float64 `json:"month_revenue"`
	UserCount       int64   `json:"user_count"`
	RecentOrders    []model.Order `json:"recent_orders"`
	TopProducts     []TopProduct  `json:"top_products"`
	PayMethodStats  []PayStat     `json:"pay_method_stats"`
	SalesTrend      []TrendPoint  `json:"sales_trend"`
}

type TopProduct struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	SoldQty  int64   `json:"sold_qty"`
	Revenue  float64 `json:"revenue"`
}

type PayStat struct {
	Method  string  `json:"method"`
	Count   int64   `json:"count"`
	Amount  float64 `json:"amount"`
}

type TrendPoint struct {
	Date    string  `json:"date"`
	Orders  int64   `json:"orders"`
	Revenue float64 `json:"revenue"`
}

func (s *DashboardService) GetStats() (*DashboardStats, error) {
	if database.DB == nil {
		return &DashboardStats{}, nil
	}

	stats := &DashboardStats{}

	database.DB.Model(&model.Product{}).Count(&stats.ProductCount)
	database.DB.Model(&model.Order{}).Count(&stats.OrderCount)
	database.DB.Model(&model.Card{}).Where("status = ?", model.CardStatusSold).Count(&stats.CardSoldCount)

	var totalRev struct{ Total float64 }
	database.DB.Model(&model.Order{}).
		Where("status >= ?", model.OrderStatusPaid).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&totalRev)
	stats.TotalRevenue = totalRev.Total

	today := time.Now().Format("2006-01-02")
	database.DB.Model(&model.Order{}).
		Where("DATE(created_at) = ? AND status >= ?", today, model.OrderStatusPaid).
		Count(&stats.TodayOrders)
	var todayRev struct{ Total float64 }
	database.DB.Model(&model.Order{}).
		Where("DATE(created_at) = ? AND status >= ?", today, model.OrderStatusPaid).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&todayRev)
	stats.TodayRevenue = todayRev.Total

	monthStart := time.Now().Format("2006-01-01")
	database.DB.Model(&model.Order{}).
		Where("DATE(created_at) >= ? AND status >= ?", monthStart, model.OrderStatusPaid).
		Count(&stats.MonthOrders)
	var monthRev struct{ Total float64 }
	database.DB.Model(&model.Order{}).
		Where("DATE(created_at) >= ? AND status >= ?", monthStart, model.OrderStatusPaid).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Scan(&monthRev)
	stats.MonthRevenue = monthRev.Total

	database.DB.Model(&model.User{}).Count(&stats.UserCount)

	var recentOrders []model.Order
	database.DB.Where("status >= ?", model.OrderStatusPaid).
		Order("id DESC").Limit(10).Find(&recentOrders)
	stats.RecentOrders = recentOrders

	var topProducts []TopProduct
	database.DB.Model(&model.Order{}).
		Select("product_id, product_name, SUM(quantity) as sold_qty, SUM(total_amount) as revenue").
		Where("status >= ?", model.OrderStatusPaid).
		Group("product_id, product_name").
		Order("sold_qty DESC").
		Limit(10).
		Scan(&topProducts)
	stats.TopProducts = topProducts

	var payStats []PayStat
	database.DB.Model(&model.Order{}).
		Select("pay_method, COUNT(*) as count, SUM(total_amount) as amount").
		Where("status >= ?", model.OrderStatusPaid).
		Group("pay_method").
		Scan(&payStats)
	stats.PayMethodStats = payStats

	stats.SalesTrend = s.getSalesTrend()

	return stats, nil
}

func (s *DashboardService) getSalesTrend() []TrendPoint {
	var points []TrendPoint
	now := time.Now()
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		var p TrendPoint
		p.Date = dateStr
		database.DB.Model(&model.Order{}).
			Where("DATE(created_at) = ? AND status >= ?", dateStr, model.OrderStatusPaid).
			Select("COUNT(*) as orders, COALESCE(SUM(total_amount), 0) as revenue").
			Scan(&p)
		points = append(points, p)
	}
	return points
}

func (s *DashboardService) GetOrderStatusCounts() (map[string]int64, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var pending, paid, complete, cancel int64
	database.DB.Model(&model.Order{}).Where("status = ?", model.OrderStatusPending).Count(&pending)
	database.DB.Model(&model.Order{}).Where("status = ?", model.OrderStatusPaid).Count(&paid)
	database.DB.Model(&model.Order{}).Where("status = ?", model.OrderStatusComplete).Count(&complete)
	database.DB.Model(&model.Order{}).Where("status = ?", model.OrderStatusCancel).Count(&cancel)
	return map[string]int64{
		"pending":  pending,
		"paid":     paid,
		"complete": complete,
		"cancel":   cancel,
	}, nil
}
