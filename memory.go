package gormtestutil

import (
	"fmt"

	testingi "github.com/mitchellh/go-testing-interface"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// sqliteConnectionString is used to create a named in-memory database that allows multiple clients
// to access the same in-memory instance. :memory: doesn't have that feature.
const sqliteConnectionString = "file:%s?mode=memory&cache=shared"

// MemoryDatabaseOption allows various options to be supplied to NewMemoryDatabase
type MemoryDatabaseOption func(*memoryDbConfig)

type memoryDbConfig struct {
	name               string
	disableForeignKeys bool
	singularConnection bool
}

const memoryConnectionString = ":memory:"

// NewMemoryDatabase returns a sqlite database that runs in-memory, allowing you to use a database
// without a running instance. Multiple options can be passed to configure the database.
func NewMemoryDatabase(t testingi.T, options ...MemoryDatabaseOption) *gorm.DB {
	t.Helper()

	config := &memoryDbConfig{}

	for _, option := range options {
		option(config)
	}

	connectionString := memoryConnectionString

	if config.name != "" {
		connectionString = fmt.Sprintf(sqliteConnectionString, config.name)
	}

	database, err := gorm.Open(sqlite.Open(connectionString), &gorm.Config{})
	if err != nil {
		t.Error(err)

		return nil
	}

	if !config.disableForeignKeys {
		if err := database.Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
			t.Error(err)

			return nil
		}
	}

	if config.singularConnection {
		db, _ := database.DB()
		db.SetMaxOpenConns(1)
	}

	// Set a 10-second database lock timeout and WAL journal mode, this allows async test to run more freely
	// in bigger tests
	database.Exec("PRAGMA journal_mode=WAL;")
	database.Exec("PRAGMA busy_timeout=10000;")

	return database
}

// WithoutForeignKeys disabled foreign keys for testing purposes
func WithoutForeignKeys() MemoryDatabaseOption {
	return func(config *memoryDbConfig) {
		config.disableForeignKeys = true
	}
}

// WithSingularConnection is useful when you're getting 'database table is locked' errors while working with
// goroutines. This wil make sure sqlite only uses 1 connection.
func WithSingularConnection() MemoryDatabaseOption {
	return func(config *memoryDbConfig) {
		config.singularConnection = true
	}
}

// WithName gives the database a name, allowing you to use NewMemoryDatabase multiple times to connect to
// the same database instance.
func WithName(name string) MemoryDatabaseOption {
	return func(config *memoryDbConfig) {
		config.name = name
	}
}
