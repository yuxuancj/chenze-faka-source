package database

import (
	"testing"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWipeTablesWhenDBNil(t *testing.T) {
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	err := WipeTables()
	assert.NoError(t, err)
}

func TestWipeTablesWhenDBHasTables(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)

	err = AutoMigrate()
	require.NoError(t, err)

	assert.True(t, DB.Migrator().HasTable(&model.User{}))

	err = WipeTables()
	assert.NoError(t, err)

	Close()
	DB = nil
}

func TestWipeTablesOnEmptyDB(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)

	err = WipeTables()
	assert.NoError(t, err)

	Close()
	DB = nil
}

func TestInitSQLiteSuccess(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)
	assert.True(t, IsAvailable())

	Close()
	DB = nil
}

func TestInitSQLiteDefaultPath(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "",
	}

	err := Init(cfg)
	require.NoError(t, err)
	assert.True(t, IsAvailable())

	Close()
	DB = nil
}

func TestInitWithInvalidSQLitePath(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "/nonexistent/path/chenze_test.db",
	}

	err := Init(cfg)
	assert.Error(t, err)
	assert.False(t, IsAvailable())

	DB = nil
}

func TestInitWithInvalidDriver(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "invalid_driver",
		Host:   "127.0.0.1",
		Port:   3306,
		User:   "root",
		DBName: "test",
	}

	err := Init(cfg)
	assert.Error(t, err)
	assert.False(t, IsAvailable())

	DB = nil
}

func TestAutoMigrateCreatesAllTables(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)

	err = AutoMigrate()
	require.NoError(t, err)

	assert.True(t, DB.Migrator().HasTable(&model.User{}))
	assert.True(t, DB.Migrator().HasTable(&model.Category{}))
	assert.True(t, DB.Migrator().HasTable(&model.Product{}))
	assert.True(t, DB.Migrator().HasTable(&model.Card{}))
	assert.True(t, DB.Migrator().HasTable(&model.Order{}))
	assert.True(t, DB.Migrator().HasTable(&model.PaymentChannel{}))
	assert.True(t, DB.Migrator().HasTable(&model.EmailConfig{}))
	assert.True(t, DB.Migrator().HasTable(&model.Node{}))
	assert.True(t, DB.Migrator().HasTable(&model.OperationLog{}))
	assert.True(t, DB.Migrator().HasTable(&model.OrderLog{}))
	assert.True(t, DB.Migrator().HasTable(&model.EmailLog{}))
	assert.True(t, DB.Migrator().HasTable(&model.UpgradeLog{}))
	assert.True(t, DB.Migrator().HasTable(&model.FileUpload{}))

	Close()
	DB = nil
}

func TestAutoMigrateWhenDBNil(t *testing.T) {
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	err := AutoMigrate()
	assert.NoError(t, err)
}

func TestTestConnectionSQLiteSuccess(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	result, err := TestConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "3.x", result.Version)
	assert.NotNil(t, result.HasData)
	assert.NotNil(t, result.TableCount)
}

func TestTestConnectionSQLiteWithTable(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)
	require.NoError(t, AutoMigrate())

	result, err := TestConnection(cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "3.x", result.Version)
	assert.True(t, result.TableCount > 0)

	Close()
	DB = nil
}

func TestTestConnectionMySQLVariousErrors(t *testing.T) {
	t.Run("connection refused on invalid port", func(t *testing.T) {
		cfg := &model.DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     19999,
			User:     "root",
			Password: "wrong",
			DBName:   "test",
		}

		result, err := TestConnection(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("connection with empty host", func(t *testing.T) {
		cfg := &model.DatabaseConfig{
			Driver:   "mysql",
			Host:     "",
			Port:     3306,
			User:     "root",
			Password: "",
			DBName:   "",
		}

		result, err := TestConnection(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("connection with invalid user", func(t *testing.T) {
		cfg := &model.DatabaseConfig{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     19999,
			User:     "nonexistent",
			Password: "wrongpass",
			DBName:   "test",
		}

		result, err := TestConnection(cfg)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCloseWhenDBNil(t *testing.T) {
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	assert.NotPanics(t, func() {
		Close()
	})
}

func TestIsAvailableAfterClose(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	require.NoError(t, err)
	assert.True(t, IsAvailable())

	Close()
	DB = nil
	assert.False(t, IsAvailable())
}