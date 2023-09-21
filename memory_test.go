package gormtestutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/clause"
)

type A struct {
	ID int
}

type B struct {
	ID  int
	A   *A `gorm:"foreignKey:AID;"`
	AID int
}

func TestNewMemoryDatabase_ReturnsDatabase(t *testing.T) {
	t.Parallel()
	// Act
	db := NewMemoryDatabase(t)

	// Assert
	assert.Equal(t, "sqlite", db.Name())
}

func TestNewMemoryDatabase_DisablesPragmaOnTrue(t *testing.T) {
	t.Parallel()
	// Act
	db := NewMemoryDatabase(t, WithoutForeignKeys())

	// Assert
	_ = db.AutoMigrate(A{}, B{})

	// This A does not exist, but pragma is off, so we should not have an error
	c := []B{{ID: 1, AID: 23}, {ID: 2, AID: 24}, {ID: 3, AID: 25}}
	assert.Nil(t, db.Omit(clause.Associations).Create(&c).Error)
}

func TestNewMemoryDatabase_EnablesPragmaByDefault(t *testing.T) {
	t.Parallel()
	// Act
	db := NewMemoryDatabase(t)

	// Assert
	_ = db.AutoMigrate(A{}, B{})

	// This A does not exist, so we expect an error
	c := []B{{ID: 1, AID: 23}, {ID: 2, AID: 24}, {ID: 3, AID: 25}}

	resultError := db.Omit(clause.Associations).Create(&c).Error

	if assert.NotNil(t, resultError) {
		assert.Equal(t, "FOREIGN KEY constraint failed", resultError.Error())
	}
}

func TestNewMemoryDatabase_ReturnsNamedDatabase(t *testing.T) {
	t.Parallel()
	// Act
	db := NewMemoryDatabase(t, WithName("abc"))

	// Assert
	_ = db.AutoMigrate(A{})
	db.Create(&A{ID: 1})

	// Get another connection and check the item
	checkDb := NewMemoryDatabase(t, WithName("abc"))

	var result *A
	checkDb.First(&result)
	assert.Equal(t, &A{ID: 1}, result)
}
