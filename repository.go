package main

import (
	"github.com/mpetavy/common"
	"strings"
)

type Repository[T any] struct {
	Database *Database
}

func NewRepository[T any](connection *Database) (*Repository[T], error) {
	common.DebugFunc()

	return &Repository[T]{
		Database: connection,
	}, nil
}

func (repository *Repository[T]) Save(records []T) error {
	common.DebugFunc()

	tx := repository.Database.Gorm.Create(records)
	if IsDuplicateKeyError(tx.Error) {
		return ErrDuplicateFound
	}

	if common.Error(tx.Error) {
		return tx.Error
	}

	return nil
}

func (repository *Repository[T]) FindById(id int) (*T, error) {
	common.DebugFunc()

	var record T

	tx := repository.Database.Gorm.First(&record, id)
	if tx.Error != nil && strings.Contains(tx.Error.Error(), "not found") {
		return nil, ErrNotFound
	}

	if common.Error(tx.Error) {
		return nil, tx.Error
	}

	return &record, nil
}

func (repository *Repository[T]) Find(where string) (*T, error) {
	common.DebugFunc()

	var record T

	tx := repository.Database.Gorm.Where(where).First(&record)
	if tx.Error != nil && strings.Contains(tx.Error.Error(), "not found") {
		return nil, ErrNotFound
	}

	if common.Error(tx.Error) {
		return nil, tx.Error
	}

	return &record, nil
}

func (repository *Repository[T]) Delete(id int) error {
	common.DebugFunc()

	var record T

	tx := repository.Database.Gorm.Delete(&record, id)
	if common.Error(tx.Error) {
		return tx.Error
	}

	if tx.RowsAffected != 1 {
		return ErrNotFound
	}

	return nil
}

func (repository *Repository[T]) FindAll(offset int, limit int) ([]T, error) {
	common.DebugFunc()

	var records []T

	tx := repository.Database.Gorm.Order("ID")

	if offset > 0 {
		tx = tx.Offset(offset)
	}

	if limit > 0 {
		tx = tx.Limit(limit)
	}

	tx = tx.Find(&records)
	if tx.Error != nil && strings.Contains(tx.Error.Error(), "not found") {
		return nil, ErrNotFound
	}

	if common.Error(tx.Error) {
		return nil, tx.Error
	}

	return records, nil
}

func (repository *Repository[T]) Update(record T) error {
	common.DebugFunc()

	tx := repository.Database.Gorm.Model(&record).Updates(record)
	if tx.RowsAffected != 1 {
		return ErrNotFound
	}
	if common.Error(tx.Error) {
		return tx.Error
	}

	return nil
}
