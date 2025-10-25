//go:build integration

package catalog_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/glb/nw-api-gogin/internal/catalog"
	"github.com/glb/nw-api-gogin/internal/db"
)

func TestCatalog_PostgresIntegration(t *testing.T) {
	cfg := db.LoadConfig()
	gormDB, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("failed to access sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	repo := catalog.NewRepository(gormDB)
	svc := catalog.NewService(repo, zaptest.NewLogger(t))
	ctx := context.Background()

	catInput := catalog.CategoryInput{Name: fmt.Sprintf("Integration %d", time.Now().UnixNano())}
	category, err := svc.CreateCategory(ctx, catInput)
	if err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.DeleteCategory(ctx, category.ID)
	})

	fetchedCategory, err := svc.GetCategory(ctx, category.ID)
	if err != nil {
		t.Fatalf("get category failed: %v", err)
	}
	if fetchedCategory.Name != catInput.Name {
		t.Fatalf("expected category name %q, got %q", catInput.Name, fetchedCategory.Name)
	}

	supplierInput := catalog.SupplierInput{CompanyName: fmt.Sprintf("Integration Supplier %d", time.Now().UnixNano())}
	supplier, err := svc.CreateSupplier(ctx, supplierInput)
	if err != nil {
		t.Fatalf("create supplier failed: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.DeleteSupplier(ctx, supplier.ID)
	})

	productInput := catalog.ProductInput{
		CategoryID:   category.ID,
		SupplierID:   supplier.ID,
		Name:         fmt.Sprintf("Integration Product %d", time.Now().UnixNano()),
		UnitPrice:    9.99,
		UnitsInStock: 10,
	}
	product, err := svc.CreateProduct(ctx, productInput)
	if err != nil {
		t.Fatalf("create product failed: %v", err)
	}
	t.Cleanup(func() {
		_ = svc.DeleteProduct(ctx, product.ID)
	})

	fetchedProduct, err := svc.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatalf("get product failed: %v", err)
	}
	if fetchedProduct.Name != productInput.Name {
		t.Fatalf("expected product name %q, got %q", productInput.Name, fetchedProduct.Name)
	}

	page, err := svc.ListProducts(ctx, catalog.ListOptions{Page: 1, PageSize: 10}, catalog.ProductFilter{CategoryID: &category.ID})
	if err != nil {
		t.Fatalf("list products failed: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatalf("expected product list to contain items")
	}
}
