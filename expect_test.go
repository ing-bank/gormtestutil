package gormtestutil

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"
)

// TestModel which gets created in TestCreated calls
type TestModel struct{}

// TableName of the TestModel struct
func (tm TestModel) TableName() string {
	return "test_models"
}

func setup() (t *testing.T, _ *gorm.DB, _ chan struct{}) {
	// Create testobject
	testObject := new(testing.T)

	// Get database
	db := NewMemoryDatabase(t, WithoutForeignKeys())

	// Get channel to communicate async
	c := make(chan struct{})

	return testObject, db, c
}

func TestExpectCreated_ModelNeverCreatedReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, c := setup()
	expectation := ExpectCreated(testObject, db, TestModel{})

	// Act
	// do nothing!

	// Assert
	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		t.FailNow()
	case <-time.After(15 * time.Second):
		// succeeds
	}
}

func TestExpectCreated_CreatedOnceWithDefaultArgumentsSucceeds(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, c := setup()
	expectation := ExpectCreated(testObject, db, TestModel{})

	// Act
	_ = db.Create(&TestModel{})

	// Assert
	go func() {
		defer close(c)
		expectation.Wait()
	}()

	select {
	case <-c:
		assert.False(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.FailNow()
	}
}

func TestExpectCreated_DuplicatedCallbackThrowsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, _ := setup()
	model := TestModel{}
	tableName := model.TableName()
	hookName := fmt.Sprintf("assert_create_%v", tableName)
	mockCallback := func(tx *gorm.DB) {}

	// Populate GORM with conflicting callback with equivalent name
	err := db.Callback().Create().Before("gorm:create").Register(hookName, mockCallback)
	assert.Nil(t, err)

	// Act
	_ = ExpectCreated(testObject, db, &model)

	// Assert
	assert.True(t, testObject.Failed())
}

func TestExpectCreated_NilDatabaseReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)
	var database *gorm.DB

	// Act
	ExpectCreated(testObject, database, &TestModel{})

	// Assert
	assert.True(t, testObject.Failed())
}

func TestExpectCreated_NonStructModelReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject := new(testing.T)
	database := &gorm.DB{}

	// Act
	ExpectCreated(testObject, database, 5)

	// Assert
	assert.True(t, testObject.Failed())
}

func TestExpectCreated_TestFailsOnMoreThanTimesAllowed(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, c := setup()
	model := TestModel{}

	// Act
	wg := ExpectCreated(testObject, db, TestModel{})
	_ = db.Create(model)
	_ = db.Create(model)

	// Assert
	go func() {
		defer func() {
			close(c)
		}()
		// Wait in a goroutine so we can assume a bounded
		wg.Wait()
	}()

	select {
	case <-c:
		assert.True(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.Errorf("test did not complete in bounded time")
	}
}

func TestExpectCreated_TestDoesNotFailOnMoreThanTimesAllowedWithStrictOff(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, c := setup()
	model := TestModel{}

	// Act
	wg := ExpectCreated(testObject, db, TestModel{}, WithoutMaximum())
	_ = db.Create(model)
	_ = db.Create(model)

	// Assert
	go func() {
		defer func() {
			close(c)
		}()
		// Wait in a goroutine so we can assume a bounded
		wg.Wait()
	}()

	select {
	case <-c:
		assert.False(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.Errorf("test did not complete in bounded time")
	}
}

func TestExpectCreated_CalledTwiceReusingExistingExpectationSucceeds(t *testing.T) {
	t.Parallel()
	// Arrange
	testObject, db, c := setup()
	model := TestModel{}

	expectation1 := &sync.WaitGroup{}
	expectation2 := ExpectCreated(testObject, db, TestModel{}, WithCalls(2), WithExpectation(expectation1))

	// Act
	_ = db.Create(model)
	_ = db.Create(model)

	// Assert
	go func() {
		defer func() {
			close(c)
		}()
		// Wait in a goroutine so we can assume a bounded
		expectation2.Wait()
	}()

	assert.Equal(t, expectation1, expectation2)

	select {
	case <-c:
		assert.False(t, testObject.Failed())
	case <-time.After(15 * time.Second):
		t.FailNow()
	}
}
