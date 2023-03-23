# 🦁 Gorm Test Utils

[![Go package](https://github.com/ing-bank/gormtestutil/actions/workflows/test.yaml/badge.svg)](https://github.com/ing-bank/gormtestutil/actions/workflows/test.yaml)
![GitHub](https://img.shields.io/github/license/ing-bank/gormtestutil)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ing-bank/gormtestutil)

Small utility functions for testing Gorm-related code.
Such as sqlite database instantiation and wait groups with callbacks.

## ⬇️ Installation

`go get github.com/ing-bank/gormtestutil`

## 📋 Usage

### Database Instantiation

```go
package main

import (
	"testing"
	"github.com/ing-bank/gormtestutil"
)

func TestProductService_FetchAll_ReturnsAllProducts(t *testing.T) {
	// Arrange
    db := gormtestutil.NewMemoryDatabase(t,
		gormtestutil.WithName(t.Name()),
		gormtestutil.WithoutForeignKeys(),
		gormtestutil.WithSingularConnection())
	
    // [...]
}
```

### Hooks

```go
package main

import (
	"github.com/ing-bank/gormtestutil"
	"time"
	"testing"
)

func TestProductService_Create_CreatesProduct(t *testing.T) {
	// Arrange
	db := gormtestutil.NewMemoryDatabase(t)

	expectation := gormtestutil.ExpectCreated(t, db, Product{}, gormtestutil.WithCalls(1))

	// Act
	go Create(db, Product{Name: "Test"})

	// Assert
	gormtestutil.EnsureCompletion(t, expectation, gormtestutil.WithTimeout(30*time.Second))
}
```

## 🚀 Development

1. Clone the repository
2. Run `make t` to run unit tests
3. Run `make fmt` to format code

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

Nothing here yet!
