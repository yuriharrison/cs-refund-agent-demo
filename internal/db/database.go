package db

import (
	"fmt"
	"log/slog"

	"github.com/yuriharrison/empirical-proj/internal/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("database ready", "dsn", dsn)
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Customer{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.RefundPolicy{},
		&domain.Refund{},
	)
}
