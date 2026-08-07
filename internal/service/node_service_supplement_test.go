package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupNodeSupplementTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllNodeTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
	return db
}

func TestNodeService_PingSuccess(t *testing.T) {
	setupNodeSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	node, err := svc.Create("健康节点", server.URL, 5)
	require.NoError(t, err)

	result, err := svc.Ping(node.ID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.NodeOnline, result.Status)
	assert.NotNil(t, result.LastPing)
	assert.True(t, result.PingTime >= 0)
}

func TestNodeService_PingNonExistent(t *testing.T) {
	setupNodeSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	result, err := svc.Ping(9999)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "节点不存在")
}

func TestNodeService_PingHTTPError(t *testing.T) {
	setupNodeSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	node, err := svc.Create("错误节点", server.URL, 5)
	require.NoError(t, err)

	result, err := svc.Ping(node.ID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.NodeOffline, result.Status)
	assert.NotNil(t, result.LastPing)
}

func TestNodeService_PingTimeout(t *testing.T) {
	setupNodeSupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	server.Config.WriteTimeout = 6 * time.Second
	server.Start()

	node, err := svc.Create("超时节点", server.URL, 5)
	require.NoError(t, err)

	start := time.Now()
	result, err := svc.Ping(node.ID)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, model.NodeOffline, result.Status)
	assert.True(t, elapsed < 8*time.Second, fmt.Sprintf("Ping should have timed out quickly, took %v", elapsed))

	server.Close()
}

func TestNodeService_PingDatabaseNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewNodeService()

	result, err := svc.Ping(1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "数据库未连接")
}