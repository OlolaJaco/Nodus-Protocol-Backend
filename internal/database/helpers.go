package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const defaultQueryTimeout = 10 * time.Second

// WithTimeout wraps a GORM DB instance with a query-scoped context timeout.
// Use instead of bare db.WithContext when the caller doesn't supply a timeout.
//
//	db := database.WithTimeout(s.db).Find(&users)
func WithTimeout(db *gorm.DB) *gorm.DB {
	ctx, _ := context.WithTimeout(context.Background(), defaultQueryTimeout) //nolint:govet
	return db.WithContext(ctx)
}

// Paginate applies LIMIT/OFFSET clauses from page and limit parameters.
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}

// SoftNotDeleted adds a WHERE clause excluding soft-deleted rows for models
// that use a deleted_at nullable column rather than GORM's DeletedAt type.
func SoftNotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}
