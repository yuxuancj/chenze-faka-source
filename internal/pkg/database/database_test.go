package database

import (
	"testing"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestInitSQLite(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	assert.NoError(t, err)
	assert.True(t, IsAvailable())

	Close()
	DB = nil
}

func TestAutoMigrate(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	assert.NoError(t, err)

	err = AutoMigrate()
	assert.NoError(t, err)

	assert.True(t, DB.Migrator().HasTable(&model.User{}))
	assert.True(t, DB.Migrator().HasTable(&model.Order{}))
	assert.True(t, DB.Migrator().HasTable(&model.Card{}))
	assert.True(t, DB.Migrator().HasTable(&model.Product{}))

	Close()
	DB = nil
}

func TestClose(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	assert.NoError(t, err)

	Close()

	sqlDB, err := DB.DB()
	if err == nil {
		err = sqlDB.Ping()
		assert.Error(t, err)
	}

	DB = nil
}

func TestIsAvailable(t *testing.T) {
	DB = nil
	assert.False(t, IsAvailable())

	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	assert.NoError(t, err)
	assert.True(t, IsAvailable())

	Close()
	DB = nil
}

func TestWipeTables(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	err := Init(cfg)
	assert.NoError(t, err)

	err = AutoMigrate()
	assert.NoError(t, err)

	err = WipeTables()
	assert.NoError(t, err)

	Close()
	DB = nil
}

func TestTestConnectionSQLite(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver: "sqlite",
		SQLite: "file::memory:?cache=shared",
	}

	result, err := TestConnection(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "3.x", result.Version)
}

func TestTestConnectionMySQLNonExistent(t *testing.T) {
	cfg := &model.DatabaseConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     9999,
		User:     "nonexistent",
		Password: "wrong",
		DBName:   "nonexistent",
	}

	result, err := TestConnection(cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
}