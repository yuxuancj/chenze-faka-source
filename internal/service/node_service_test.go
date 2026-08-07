package service

import (
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllNodeTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupNodeTestDB(t *testing.T) *gorm.DB {
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

func TestNodeService_Create(t *testing.T) {
	setupNodeTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	t.Run("create node successfully", func(t *testing.T) {
		node, err := svc.Create("节点1", "https://node1.example.com", 5)
		require.NoError(t, err)
		assert.NotNil(t, node)
		assert.Equal(t, "节点1", node.Name)
		assert.Equal(t, "https://node1.example.com", node.URL)
		assert.Equal(t, 5, node.Weight)
		assert.Equal(t, model.NodeOnline, node.Status)
	})

	t.Run("create node with empty name fails", func(t *testing.T) {
		node, err := svc.Create("", "https://node.example.com", 1)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "名称和URL不能为空")
	})

	t.Run("create node with empty URL fails", func(t *testing.T) {
		node, err := svc.Create("节点", "", 1)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "名称和URL不能为空")
	})

	t.Run("create node with zero weight uses default 1", func(t *testing.T) {
		node, err := svc.Create("节点2", "https://node2.example.com", 0)
		require.NoError(t, err)
		assert.Equal(t, 1, node.Weight)
	})
}

func TestNodeService_Update(t *testing.T) {
	setupNodeTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	node, err := svc.Create("节点1", "https://node1.example.com", 5)
	require.NoError(t, err)

	t.Run("update node name and url", func(t *testing.T) {
		updated, err := svc.Update(node.ID, "节点1更新", "https://node1-new.example.com", 10, model.NodeOnline)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "节点1更新", updated.Name)
		assert.Equal(t, "https://node1-new.example.com", updated.URL)
		assert.Equal(t, 10, updated.Weight)
		assert.Equal(t, model.NodeOnline, updated.Status)
	})

	t.Run("update with empty name keeps old name", func(t *testing.T) {
		updated, err := svc.Update(node.ID, "", "", 3, model.NodeOffline)
		require.NoError(t, err)
		assert.Equal(t, "节点1更新", updated.Name)
		assert.Equal(t, "https://node1-new.example.com", updated.URL)
		assert.Equal(t, 3, updated.Weight)
		assert.Equal(t, model.NodeOffline, updated.Status)
	})

	t.Run("update nonexistent node", func(t *testing.T) {
		updated, err := svc.Update(9999, "不存在", "", 1, model.NodeOnline)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "节点不存在")
	})
}

func TestNodeService_Delete(t *testing.T) {
	setupNodeTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	node, err := svc.Create("待删除节点", "https://delete.example.com", 1)
	require.NoError(t, err)

	err = svc.Delete(node.ID)
	require.NoError(t, err)

	_, err = svc.Update(node.ID, "test", "", 1, model.NodeOnline)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")
}

func TestNodeService_List(t *testing.T) {
	setupNodeTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	_, err := svc.Create("节点A", "https://a.example.com", 1)
	require.NoError(t, err)
	_, err = svc.Create("节点B", "https://b.example.com", 2)
	require.NoError(t, err)
	_, err = svc.Create("节点C", "https://c.example.com", 3)
	require.NoError(t, err)

	t.Run("list all nodes", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 3)
	})

	t.Run("list with keyword", func(t *testing.T) {
		result, err := svc.List(1, 10, "节点A")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list with URL keyword", func(t *testing.T) {
		result, err := svc.List(1, 10, "b.example")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)

		result2, err := svc.List(2, 2, "")
		require.NoError(t, err)
		assert.Len(t, result2.Items, 1)
	})

	t.Run("list with default page", func(t *testing.T) {
		result, err := svc.List(0, 10, "")
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with default pageSize", func(t *testing.T) {
		result, err := svc.List(1, 0, "")
		require.NoError(t, err)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestNodeService_GetBestNode(t *testing.T) {
	setupNodeTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewNodeService()

	t.Run("no available nodes", func(t *testing.T) {
		node, err := svc.GetBestNode()
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "没有可用节点")
	})

	node1, err := svc.Create("节点A", "https://a.example.com", 5)
	require.NoError(t, err)
	_, err = svc.Update(node1.ID, "", "", 5, model.NodeOnline)
	require.NoError(t, err)

	node2, err := svc.Create("节点B", "https://b.example.com", 10)
	require.NoError(t, err)
	_, err = svc.Update(node2.ID, "", "", 10, model.NodeOnline)
	require.NoError(t, err)

	t.Run("returns node with highest weight", func(t *testing.T) {
		best, err := svc.GetBestNode()
		require.NoError(t, err)
		assert.NotNil(t, best)
		assert.Equal(t, "节点B", best.Name)
		assert.Equal(t, 10, best.Weight)
	})
}

func TestNodeService_DBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewNodeService()

	t.Run("create with nil DB", func(t *testing.T) {
		node, err := svc.Create("test", "https://example.com", 1)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("update with nil DB", func(t *testing.T) {
		node, err := svc.Update(1, "test", "", 1, 1)
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("delete with nil DB", func(t *testing.T) {
		err := svc.Delete(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("list with nil DB", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Items, 0)
	})

	t.Run("getBestNode with nil DB", func(t *testing.T) {
		node, err := svc.GetBestNode()
		assert.Error(t, err)
		assert.Nil(t, node)
		assert.Contains(t, err.Error(), "数据库未连接")
	})
}