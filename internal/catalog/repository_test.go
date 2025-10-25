package catalog

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestRepository_WithTxRollback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&Category{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to access sql DB: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	repo := NewRepository(db)
	sentinel := errors.New("tx failure")

	err = repo.WithTx(context.Background(), func(txRepo *Repository) error {
		if txRepo == repo {
			t.Fatalf("expected transactional repository to be distinct instance")
		}
		category := Category{Name: "Transactional"}
		if saveErr := txRepo.SaveCategory(context.Background(), &category); saveErr != nil {
			t.Fatalf("failed to save category inside tx: %v", saveErr)
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	var count int64
	if err := db.Model(&Category{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count categories: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected rollback to remove inserted rows, found %d", count)
	}
}
