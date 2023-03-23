package gormtestutil

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func ExampleNewMemoryDatabase() {
	type MyObject struct {
		ID int
	}

	t := new(testing.T)
	database := NewMemoryDatabase(t)

	var count int64
	database.Model(&MyObject{}).Count(&count)
	fmt.Printf("There are %d objects\n", count)
}

func ExampleNewMemoryDatabase_ignoringForeignKeys() {
	type MyObject struct {
		ID int
	}

	t := new(testing.T)
	database := NewMemoryDatabase(t, WithoutForeignKeys())

	var count int64
	database.Model(&MyObject{}).Count(&count)
	fmt.Printf("There are %d objects\n", count)
}

func ExampleNewMemoryDatabase_withSingularConnection() {
	type MyObject struct {
		ID int
	}

	t := new(testing.T)

	database1 := NewMemoryDatabase(t, WithName(t.Name()))
	database2 := NewMemoryDatabase(t, WithName(t.Name()))

	database1.Create(&MyObject{ID: 2})

	var result MyObject
	database2.First(&result)
	fmt.Println(result.ID)
}

// example where the default arguments are used (expect created once and no previous expectations)
// with an upper limit of test time
func ExampleExpectCreated_defaults() {
	var t *testing.T

	// arrange
	type SomeModel struct {
		Name string
	}

	database := NewMemoryDatabase(t)
	expectation := ExpectCreated(t, database, &SomeModel{})

	// Act
	_ = database.Create(SomeModel{Name: "Hello, world!"})

	// Assert
	if ok := EnsureCompletion(t, expectation); !ok {
		t.FailNow()
	}
}

// example where the default arguments are used (expect created once and no previous expectations)
// with an upper limit of test time
func ExampleExpectCreated_withVarArgs() {
	var t *testing.T
	var exp *sync.WaitGroup

	// arrange
	type SomeModel struct {
		Name string
	}

	database := NewMemoryDatabase(t)
	expectation := ExpectCreated(t, database, &SomeModel{}, WithCalls(2), WithExpectation(exp), WithoutMaximum())

	// Act
	_ = database.Create(SomeModel{Name: "Hello, world!"})
	_ = database.Create(SomeModel{Name: "And another time"})

	// Assert
	if ok := EnsureCompletion(t, expectation, WithTimeout(15*time.Second)); !ok {
		t.FailNow()
	}
}
