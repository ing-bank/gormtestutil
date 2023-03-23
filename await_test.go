package gormtestutil

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestEnsure_NilWaitGroupFails(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(testing.T)

	// Act
	ok := EnsureCompletion(testingObject, nil)

	// Assert
	assert.False(t, ok)
}

func TestEnsure_NegativeDurationFails(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(testing.T)

	expectation := &sync.WaitGroup{}
	expectation.Add(1)
	go func() {
		timeout := time.After(1 * time.Second)

		<-timeout
		expectation.Done()
	}()

	// Act
	ok := EnsureCompletion(testingObject, expectation, WithTimeout(-1*time.Second))

	// Assert
	assert.True(t, testingObject.Failed())
	assert.False(t, ok)
}

func TestEnsure_LongerTaskTimeThanEnsureDurationFails(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(testing.T)

	expectation := &sync.WaitGroup{}
	expectation.Add(1)
	go func() {
		timeout := time.After(5 * time.Second)

		<-timeout
		expectation.Done()
	}()

	// Act
	ok := EnsureCompletion(testingObject, expectation, WithTimeout(1*time.Second))

	// Assert
	assert.True(t, testingObject.Failed())
	assert.False(t, ok)
}

func TestEnsure_ShortTaskTimeThanEnsureDurationSucceeds(t *testing.T) {
	t.Parallel()
	// Arrange
	testingObject := new(testing.T)

	expectation := &sync.WaitGroup{}
	expectation.Add(1)
	go func() {
		timeout := time.After(1 * time.Second)

		<-timeout
		expectation.Done()
	}()

	// Act
	ok := EnsureCompletion(testingObject, expectation, WithTimeout(3*time.Second))

	// Assert
	assert.False(t, testingObject.Failed())
	assert.True(t, ok)
}
