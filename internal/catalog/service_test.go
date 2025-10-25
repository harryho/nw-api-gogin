package catalog

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&Category{}, &Supplier{}, &Product{}); err != nil {
		t.Fatalf("failed to migrate test schema: %v", err)
	}

	repo := NewRepository(db)
	svc := NewService(repo, zaptest.NewLogger(t))

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to access sql DB: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return svc, db
}

func TestService_CreateCategory_ValidationError(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.CreateCategory(context.Background(), CategoryInput{Name: "   "})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	appErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected catalog error, got %v", err)
	}
	if appErr.Code != ErrorValidation {
		t.Fatalf("expected validation error code, got %s", appErr.Code)
	}
}

func TestService_CreateCategory_Success(t *testing.T) {
	svc, db := newTestService(t)

	category, err := svc.CreateCategory(context.Background(), CategoryInput{Name: "Beverages"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if category.ID == 0 {
		t.Fatalf("expected persisted category with ID")
	}

	var stored Category
	if err := db.WithContext(context.Background()).First(&stored, category.ID).Error; err != nil {
		t.Fatalf("failed to load stored category: %v", err)
	}
	if stored.Name != "Beverages" {
		t.Fatalf("expected stored name 'Beverages', got %s", stored.Name)
	}
}

func TestService_GetCategory_NotFound(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.GetCategory(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected not found error")
	}

	appErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected catalog error, got %v", err)
	}
	if appErr.Code != ErrorNotFound {
		t.Fatalf("expected error code %s, got %s", ErrorNotFound, appErr.Code)
	}
}

func TestService_ListCategories_InvalidSort(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.ListCategories(context.Background(), ListOptions{Sort: "-unknown"}, CategoryFilter{})
	if err == nil {
		t.Fatalf("expected validation error for invalid sort field")
	}

	appErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected catalog error, got %v", err)
	}
	if appErr.Code != ErrorValidation {
		t.Fatalf("expected validation error code, got %s", appErr.Code)
	}
}

func TestService_UpdateCategory_Success(t *testing.T) {
	svc, db := newTestService(t)

	category, err := svc.CreateCategory(context.Background(), CategoryInput{Name: "Beverages"})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	updated, err := svc.UpdateCategory(context.Background(), category.ID, CategoryInput{Name: "Hot Beverages", Description: strPtr("  Tea ")})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Name != "Hot Beverages" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
	if updated.Description == nil || *updated.Description != "Tea" {
		t.Fatalf("expected trimmed description 'Tea', got %#v", updated.Description)
	}

	var stored Category
	if err := db.First(&stored, category.ID).Error; err != nil {
		t.Fatalf("failed to reload category: %v", err)
	}
	if stored.Name != "Hot Beverages" {
		t.Fatalf("expected stored name 'Hot Beverages', got %s", stored.Name)
	}
}

func TestService_UpdateCategory_NotFound(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.UpdateCategory(context.Background(), 123, CategoryInput{Name: "Missing"})
	if err == nil {
		t.Fatalf("expected not found error")
	}

	appErr, ok := AsError(err)
	if !ok {
		t.Fatalf("expected catalog error, got %v", err)
	}
	if appErr.Code != ErrorNotFound {
		t.Fatalf("expected not found code, got %s", appErr.Code)
	}
}

func TestService_DeleteCategory_Success(t *testing.T) {
	svc, db := newTestService(t)
	category, err := svc.CreateCategory(context.Background(), CategoryInput{Name: "ToDelete"})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if err := svc.DeleteCategory(context.Background(), category.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	var count int64
	if err := db.Model(&Category{}).Count(&count).Error; err != nil {
		t.Fatalf("failed to count categories: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected category removed, found %d", count)
	}
}

func TestService_DeleteCategory_NotFound(t *testing.T) {
	svc, _ := newTestService(t)

	if err := svc.DeleteCategory(context.Background(), 321); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestService_CreateProduct_Validation(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.CreateProduct(context.Background(), ProductInput{Name: "  ", CategoryID: 1, SupplierID: 1, UnitPrice: -1})
	if err == nil {
		t.Fatalf("expected validation error for name and price")
	}

	appErr, _ := AsError(err)
	if appErr == nil || appErr.Code != ErrorValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestService_CreateProduct_Success(t *testing.T) {
	svc, _ := newTestService(t)

	category, err := svc.CreateCategory(context.Background(), CategoryInput{Name: "Condiments"})
	if err != nil {
		t.Fatalf("unexpected category error: %v", err)
	}
	supplier, err := svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Acme"})
	if err != nil {
		t.Fatalf("unexpected supplier error: %v", err)
	}

	product, err := svc.CreateProduct(context.Background(), ProductInput{
		Name:         "Sauce",
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    10,
		UnitsInStock: 5,
	})
	if err != nil {
		t.Fatalf("unexpected product error: %v", err)
	}
	if product.ID == 0 {
		t.Fatalf("expected persisted product")
	}
}

func TestService_UpdateProduct_Success(t *testing.T) {
	svc, _ := newTestService(t)
	category, _ := svc.CreateCategory(context.Background(), CategoryInput{Name: "Produce"})
	supplier, _ := svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Growers"})
	product, err := svc.CreateProduct(context.Background(), ProductInput{
		Name:         "Apple",
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    1.5,
		UnitsInStock: 10,
	})
	if err != nil {
		t.Fatalf("unexpected product create error: %v", err)
	}

	updated, err := svc.UpdateProduct(context.Background(), product.ID, ProductInput{
		Name:         "Red Apple",
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    2,
		UnitsInStock: 20,
	})
	if err != nil {
		t.Fatalf("unexpected product update error: %v", err)
	}
	if updated.Name != "Red Apple" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
}

func TestService_DeleteProduct_Success(t *testing.T) {
	svc, _ := newTestService(t)
	category, _ := svc.CreateCategory(context.Background(), CategoryInput{Name: "Snacks"})
	supplier, _ := svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Snack Co"})
	product, err := svc.CreateProduct(context.Background(), ProductInput{
		Name:         "Chips",
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    2.5,
		UnitsInStock: 15,
	})
	if err != nil {
		t.Fatalf("unexpected product create error: %v", err)
	}

	if err := svc.DeleteProduct(context.Background(), product.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

func TestService_DeleteProduct_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.DeleteProduct(context.Background(), 999); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestService_ListProducts_Success(t *testing.T) {
	svc, _ := newTestService(t)
	category, _ := svc.CreateCategory(context.Background(), CategoryInput{Name: "Produce"})
	supplier, _ := svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Fresh"})
	_, _ = svc.CreateProduct(context.Background(), ProductInput{
		Name:         "Apple",
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		UnitPrice:    1.0,
		UnitsInStock: 5,
	})

	page, err := svc.ListProducts(context.Background(), ListOptions{}, ProductFilter{})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatalf("expected items in list")
	}
}

func TestService_ListSuppliers_Success(t *testing.T) {
	svc, _ := newTestService(t)
	_, _ = svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Global Supply"})

	page, err := svc.ListSuppliers(context.Background(), ListOptions{}, SupplierFilter{})
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatalf("expected suppliers in list")
	}
}

func TestService_GetSupplier_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.GetSupplier(context.Background(), 500); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestService_DeleteSupplier_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	if err := svc.DeleteSupplier(context.Background(), 400); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestService_loggerForContext(t *testing.T) {
	svc, _ := newTestService(t)
	logger := svc.loggerForContext(context.Background())
	if logger == nil {
		t.Fatalf("expected logger instance")
	}
}

func TestService_CreateSupplier_Success(t *testing.T) {
	svc, _ := newTestService(t)

	supplier, err := svc.CreateSupplier(context.Background(), SupplierInput{
		CompanyName: "Supply Co",
		Phone:       strPtr(" 123-456 "),
	})
	if err != nil {
		t.Fatalf("unexpected supplier error: %v", err)
	}
	if supplier.Phone == nil || *supplier.Phone != "123-456" {
		t.Fatalf("expected trimmed phone, got %#v", supplier.Phone)
	}
}

func TestService_UpdateSupplier_Success(t *testing.T) {
	svc, _ := newTestService(t)
	supplier, _ := svc.CreateSupplier(context.Background(), SupplierInput{CompanyName: "Old"})

	updated, err := svc.UpdateSupplier(context.Background(), supplier.ID, SupplierInput{CompanyName: "New Name"})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.CompanyName != "New Name" {
		t.Fatalf("expected updated company name, got %s", updated.CompanyName)
	}
}

func strPtr(value string) *string {
	return &value
}
