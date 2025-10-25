package catalog_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/glb/nw-api-gogin/internal/api"
	"github.com/glb/nw-api-gogin/internal/catalog"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&catalog.Category{}, &catalog.Supplier{}, &catalog.Product{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	repo := catalog.NewRepository(db)
	svc := catalog.NewService(repo, zaptest.NewLogger(t))

	r := gin.New()
	r.Use(gin.Recovery())

	handler := api.NewHandler(svc)
	api.RegisterHandlers(r, handler)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to access sql db: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return r
}

func TestIntegration_FullCatalogLifecycle(t *testing.T) {
	router := setupRouter(t)

	// Create category
	categoryReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":"Beverages"}`))
	categoryReq.Header.Set("Content-Type", binding.MIMEJSON)
	categoryResp := httptest.NewRecorder()
	router.ServeHTTP(categoryResp, categoryReq)
	if categoryResp.Code != http.StatusCreated {
		t.Fatalf("expected category create 201, got %d", categoryResp.Code)
	}

	var category api.Category
	if err := json.NewDecoder(categoryResp.Body).Decode(&category); err != nil {
		t.Fatalf("failed to decode category: %v", err)
	}

	// Update category
	updateReq := httptest.NewRequest(http.MethodPut, "/categories/"+strconv.Itoa(category.Id), bytes.NewBufferString(`{"name":"Hot Beverages"}`))
	updateReq.Header.Set("Content-Type", binding.MIMEJSON)
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected category update 200, got %d", updateResp.Code)
	}

	// List categories
	listReq := httptest.NewRequest(http.MethodGet, "/categories", nil)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected category list 200, got %d", listResp.Code)
	}

	// Create supplier
	supplierReq := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewBufferString(`{"companyName":"Supply Co"}`))
	supplierReq.Header.Set("Content-Type", binding.MIMEJSON)
	supplierResp := httptest.NewRecorder()
	router.ServeHTTP(supplierResp, supplierReq)
	if supplierResp.Code != http.StatusCreated {
		t.Fatalf("expected supplier create 201, got %d", supplierResp.Code)
	}

	var supplier api.Supplier
	if err := json.NewDecoder(supplierResp.Body).Decode(&supplier); err != nil {
		t.Fatalf("failed to decode supplier: %v", err)
	}

	// Create product
	productPayload := `{"name":"Chai","categoryId":` + strconv.Itoa(category.Id) + `,"supplierId":` + strconv.Itoa(supplier.Id) + `,"unitPrice":18,"unitsInStock":10}`
	productReq := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(productPayload))
	productReq.Header.Set("Content-Type", binding.MIMEJSON)
	productResp := httptest.NewRecorder()

	router.ServeHTTP(productResp, productReq)
	if productResp.Code != http.StatusCreated {
		t.Fatalf("expected product create 201, got %d", productResp.Code)
	}

	var product api.Product
	if err := json.NewDecoder(productResp.Body).Decode(&product); err != nil {
		t.Fatalf("failed to decode product: %v", err)
	}

	// Get product
	getProductReq := httptest.NewRequest(http.MethodGet, "/products/"+strconv.Itoa(product.Id), nil)
	getProductResp := httptest.NewRecorder()
	router.ServeHTTP(getProductResp, getProductReq)
	if getProductResp.Code != http.StatusOK {
		t.Fatalf("expected get product 200, got %d", getProductResp.Code)
	}

	// Update product
	updateProductPayload := `{"name":"Updated Chai","categoryId":` + strconv.Itoa(category.Id) + `,"supplierId":` + strconv.Itoa(supplier.Id) + `,"unitPrice":20,"unitsInStock":8}`
	updateProductReq := httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(product.Id), bytes.NewBufferString(updateProductPayload))
	updateProductReq.Header.Set("Content-Type", binding.MIMEJSON)
	updateProductResp := httptest.NewRecorder()
	router.ServeHTTP(updateProductResp, updateProductReq)
	if updateProductResp.Code != http.StatusOK {
		t.Fatalf("expected update product 200, got %d", updateProductResp.Code)
	}

	// List products
	listProductsReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	listProductsResp := httptest.NewRecorder()
	router.ServeHTTP(listProductsResp, listProductsReq)
	if listProductsResp.Code != http.StatusOK {
		t.Fatalf("expected list products 200, got %d", listProductsResp.Code)
	}

	// Delete product
	deleteProductReq := httptest.NewRequest(http.MethodDelete, "/products/"+strconv.Itoa(product.Id), nil)
	deleteProductResp := httptest.NewRecorder()
	router.ServeHTTP(deleteProductResp, deleteProductReq)
	if deleteProductResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete product 204, got %d", deleteProductResp.Code)
	}

	// Delete supplier
	deleteSupplierReq := httptest.NewRequest(http.MethodDelete, "/suppliers/"+strconv.Itoa(supplier.Id), nil)
	deleteSupplierResp := httptest.NewRecorder()
	router.ServeHTTP(deleteSupplierResp, deleteSupplierReq)
	if deleteSupplierResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete supplier 204, got %d", deleteSupplierResp.Code)
	}

	// Delete category
	deleteCategoryReq := httptest.NewRequest(http.MethodDelete, "/categories/"+strconv.Itoa(category.Id), nil)
	deleteCategoryResp := httptest.NewRecorder()
	router.ServeHTTP(deleteCategoryResp, deleteCategoryReq)
	if deleteCategoryResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete category 204, got %d", deleteCategoryResp.Code)
	}

	// Request token (covers IssueToken)
	tokenReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewBufferString(`{"username":"a","password":"b","scope":"viewer"}`))
	tokenReq.Header.Set("Content-Type", binding.MIMEJSON)
	tokenResp := httptest.NewRecorder()
	router.ServeHTTP(tokenResp, tokenReq)
	if tokenResp.Code != http.StatusNotImplemented {
		t.Fatalf("expected token endpoint 501, got %d", tokenResp.Code)
	}
}
