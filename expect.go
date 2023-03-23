package gormtestutil

import (
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"sync"
	"testing"
)

const (
	defaultTimesCalled = 1
	defaultStrict      = true
)

// ExpectOption allows various options to be supplied to Expect* functions
type ExpectOption func(*expectConfig)

// WithCalls is used to expect an invocation an X amount of times
func WithCalls(times int) ExpectOption {
	return func(config *expectConfig) {
		config.Times = times
	}
}

// WithExpectation allows you to chain wait groups and expectations together
func WithExpectation(expectation *sync.WaitGroup) ExpectOption {
	return func(config *expectConfig) {
		config.Expectation = expectation
	}
}

// WithoutMaximum instructs the expectation to ignore an excess amount of calls. By default, any more calls
// than the expected 'times' cause an error.
func WithoutMaximum() ExpectOption {
	return func(config *expectConfig) {
		config.Strict = false
	}
}

// ExpectCreated asserts that an insert statement has at least been executed on the model.
func ExpectCreated(t *testing.T, database *gorm.DB, model any, options ...ExpectOption) *sync.WaitGroup {
	t.Helper()
	return expectHook(t, database, model, "create", options...)
}

// ExpectDeleted asserts that a delete statement has at least been executed on the model.
func ExpectDeleted(t *testing.T, database *gorm.DB, model any, options ...ExpectOption) *sync.WaitGroup {
	t.Helper()
	return expectHook(t, database, model, "delete", options...)
}

// ExpectUpdated asserts that an update statement has at least been executed on the model.
func ExpectUpdated(t *testing.T, database *gorm.DB, model any, options ...ExpectOption) *sync.WaitGroup {
	t.Helper()
	return expectHook(t, database, model, "update", options...)
}

type expectConfig struct {
	Times       int
	Strict      bool
	Expectation *sync.WaitGroup
}

// expectHook asserts that a hook has at least been executed on the model.
func expectHook(t *testing.T, database *gorm.DB, model any, hook string, options ...ExpectOption) *sync.WaitGroup {
	t.Helper()

	if database == nil {
		t.Error("database cannot be nil")
		return nil
	}

	kind := reflect.ValueOf(model).Kind()
	if kind != reflect.Struct {
		t.Error("model must be a struct")
		return nil
	}

	// Default values
	config := &expectConfig{
		Times:       defaultTimesCalled,
		Strict:      defaultStrict,
		Expectation: &sync.WaitGroup{},
	}

	for _, option := range options {
		option(config)
	}

	// Set waitgroup for amount of times
	config.Expectation.Add(config.Times)

	// Get table name of model to use in register hook
	stmt := &gorm.Statement{DB: database}
	if err := stmt.Parse(model); err != nil {
		t.Error(err)
		return nil
	}
	tableName := stmt.Table

	var timesCalled int
	assertHook := func(tx *gorm.DB) {
		t.Helper()

		if tx.Statement.Table == tableName {
			timesCalled++
			if timesCalled <= config.Times {
				config.Expectation.Done()
			} else {
				message := fmt.Sprintf("%s hook asserts called %d times but called at least %d times\n", tableName, config.Times, timesCalled)
				if config.Strict {
					t.Errorf(message)
				} else {
					t.Log(message)
				}
				return
			}
		}
	}

	hookName := fmt.Sprintf("assert_%s_%v", hook, tableName)
	switch hook {
	case "create":
		gormHook := "gorm:after_create"
		if err := database.Callback().Create().After(gormHook).Register(hookName, assertHook); err != nil {
			t.Error(err)
			return nil
		}
	case "delete":
		gormHook := "gorm:after_delete"
		if err := database.Callback().Delete().After(gormHook).Register(hookName, assertHook); err != nil {
			t.Error(err)
			return nil
		}
	case "update":
		gormHook := "gorm:after_update"
		if err := database.Callback().Update().After(gormHook).Register(hookName, assertHook); err != nil {
			t.Error(err)
			return nil
		}
	}

	return config.Expectation
}
