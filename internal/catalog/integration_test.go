package catalog_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/harryho/nw-api-gogin/internal/api"
	"github.com/harryho/nw-api-gogin/internal/auth"
	"github.com/harryho/nw-api-gogin/internal/catalog"
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

	authenticator, err := auth.NewStaticAuthenticator(map[string]struct {
		PasswordHash []byte
		Principal    auth.Principal
	}{
		"admin": {
			PasswordHash: auth.MustHashPassword("secret"),
			Principal:    auth.Principal{Subject: "admin", Scopes: []string{"admin", "manager", "viewer"}},
		},
	})
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	keyManager, err := auth.NewHMACKeyManager([]byte("integration-secret"), "integration")
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	tokenSvc, err := auth.NewService(auth.Config{
		Issuer:         "integration-test",
		Audience:       []string{"integration-audience"},
		AccessTokenTTL: time.Hour,
	}, authenticator, keyManager)
	if err != nil {
		t.Fatalf("failed to create token service: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	handler := api.NewHandler(svc, tokenSvc)
	api.RegisterHandlersWithOptions(r, handler, api.GinServerOptions{
		Middlewares: []api.MiddlewareFunc{api.AuthMiddleware(tokenSvc)},
	})

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

	// Request token for authenticated operations
	initialTokenReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	initialTokenReq.Header.Set("Content-Type", binding.MIMEJSON)
	initialTokenResp := httptest.NewRecorder()
	router.ServeHTTP(initialTokenResp, initialTokenReq)
	if initialTokenResp.Code != http.StatusOK {
		t.Fatalf("expected initial token 200, got %d", initialTokenResp.Code)
	}

	var issued api.TokenResponse
	if err := json.NewDecoder(initialTokenResp.Body).Decode(&issued); err != nil {
		t.Fatalf("failed to decode initial token: %v", err)
	}
	if issued.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	authHeader := "Bearer " + issued.AccessToken

	viewerTokenReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewBufferString(`{"username":"admin","password":"secret","scope":"viewer"}`))
	viewerTokenReq.Header.Set("Content-Type", binding.MIMEJSON)
	viewerTokenResp := httptest.NewRecorder()
	router.ServeHTTP(viewerTokenResp, viewerTokenReq)
	if viewerTokenResp.Code != http.StatusOK {
		t.Fatalf("expected viewer token 200, got %d", viewerTokenResp.Code)
	}

	var viewerToken api.TokenResponse
	if err := json.NewDecoder(viewerTokenResp.Body).Decode(&viewerToken); err != nil {
		t.Fatalf("failed to decode viewer token: %v", err)
	}
	if viewerToken.AccessToken == "" {
		t.Fatalf("expected viewer token to contain access token")
	}
	viewerAuthHeader := "Bearer " + viewerToken.AccessToken

	forbiddenReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":"Scoped"}`))
	forbiddenReq.Header.Set("Content-Type", binding.MIMEJSON)
	forbiddenReq.Header.Set("Authorization", viewerAuthHeader)
	forbiddenResp := httptest.NewRecorder()
	router.ServeHTTP(forbiddenResp, forbiddenReq)
	if forbiddenResp.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for viewer scope, got %d", forbiddenResp.Code)
	}

	// Create category
	categoryReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":"Beverages"}`))
	categoryReq.Header.Set("Content-Type", binding.MIMEJSON)
	categoryReq.Header.Set("Authorization", authHeader)
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
	updateReq.Header.Set("Authorization", authHeader)
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected category update 200, got %d", updateResp.Code)
	}

	// List categories
	listReq := httptest.NewRequest(http.MethodGet, "/categories", nil)
	listReq.Header.Set("Authorization", authHeader)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected category list 200, got %d", listResp.Code)
	}

	// Create supplier
	supplierReq := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewBufferString(`{"companyName":"Supply Co"}`))
	supplierReq.Header.Set("Content-Type", binding.MIMEJSON)
	supplierReq.Header.Set("Authorization", authHeader)
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
	productReq.Header.Set("Authorization", authHeader)
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
	getProductReq.Header.Set("Authorization", authHeader)
	getProductResp := httptest.NewRecorder()
	router.ServeHTTP(getProductResp, getProductReq)
	if getProductResp.Code != http.StatusOK {
		t.Fatalf("expected get product 200, got %d", getProductResp.Code)
	}

	// Update product
	updateProductPayload := `{"name":"Updated Chai","categoryId":` + strconv.Itoa(category.Id) + `,"supplierId":` + strconv.Itoa(supplier.Id) + `,"unitPrice":20,"unitsInStock":8}`
	updateProductReq := httptest.NewRequest(http.MethodPut, "/products/"+strconv.Itoa(product.Id), bytes.NewBufferString(updateProductPayload))
	updateProductReq.Header.Set("Content-Type", binding.MIMEJSON)
	updateProductReq.Header.Set("Authorization", authHeader)
	updateProductResp := httptest.NewRecorder()
	router.ServeHTTP(updateProductResp, updateProductReq)
	if updateProductResp.Code != http.StatusOK {
		t.Fatalf("expected update product 200, got %d", updateProductResp.Code)
	}

	// List products
	listProductsReq := httptest.NewRequest(http.MethodGet, "/products", nil)
	listProductsReq.Header.Set("Authorization", authHeader)
	listProductsResp := httptest.NewRecorder()
	router.ServeHTTP(listProductsResp, listProductsReq)
	if listProductsResp.Code != http.StatusOK {
		t.Fatalf("expected list products 200, got %d", listProductsResp.Code)
	}

	// Delete product
	deleteProductReq := httptest.NewRequest(http.MethodDelete, "/products/"+strconv.Itoa(product.Id), nil)
	deleteProductReq.Header.Set("Authorization", authHeader)
	deleteProductResp := httptest.NewRecorder()
	router.ServeHTTP(deleteProductResp, deleteProductReq)
	if deleteProductResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete product 204, got %d", deleteProductResp.Code)
	}

	// Delete supplier
	deleteSupplierReq := httptest.NewRequest(http.MethodDelete, "/suppliers/"+strconv.Itoa(supplier.Id), nil)
	deleteSupplierReq.Header.Set("Authorization", authHeader)
	deleteSupplierResp := httptest.NewRecorder()
	router.ServeHTTP(deleteSupplierResp, deleteSupplierReq)
	if deleteSupplierResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete supplier 204, got %d", deleteSupplierResp.Code)
	}

	// Delete category
	deleteCategoryReq := httptest.NewRequest(http.MethodDelete, "/categories/"+strconv.Itoa(category.Id), nil)
	deleteCategoryReq.Header.Set("Authorization", authHeader)
	deleteCategoryResp := httptest.NewRecorder()
	router.ServeHTTP(deleteCategoryResp, deleteCategoryReq)
	if deleteCategoryResp.Code != http.StatusNoContent {
		t.Fatalf("expected delete category 204, got %d", deleteCategoryResp.Code)
	}

	// Request token (covers IssueToken)
	tokenReq := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewBufferString(`{"username":"admin","password":"secret","scope":"viewer"}`))
	tokenReq.Header.Set("Content-Type", binding.MIMEJSON)
	tokenResp := httptest.NewRecorder()
	router.ServeHTTP(tokenResp, tokenReq)
	if tokenResp.Code != http.StatusOK {
		t.Fatalf("expected token endpoint 200, got %d", tokenResp.Code)
	}

	var token api.TokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&token); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if token.AccessToken == "" {
		t.Fatalf("expected access token value")
	}
	if token.TokenType != "Bearer" {
		t.Fatalf("expected token type Bearer, got %q", token.TokenType)
	}
	if token.ExpiresIn <= 0 {
		t.Fatalf("expected positive expiresIn, got %d", token.ExpiresIn)
	}
}
