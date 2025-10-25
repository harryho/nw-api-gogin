package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/glb/nw-api-gogin/internal/auth"
	"github.com/glb/nw-api-gogin/internal/catalog"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_ListCategories_Success(t *testing.T) {
	stub := &catalogServiceStub{
		listCategoriesFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error) {
			page := catalog.Pagination{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1}
			items := []catalog.Category{{ID: 1, Name: "Beverages"}}
			return catalog.Page[catalog.Category]{Items: items, Meta: page}, nil
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	handler.ListCategories(ctx, ListCategoriesParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp CategoryListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected one category, got %d", len(resp.Data))
	}
	if resp.Data[0].Id != 1 {
		t.Fatalf("expected category id 1, got %d", resp.Data[0].Id)
	}
	if resp.Meta.TotalItems != 1 {
		t.Fatalf("expected total items 1, got %d", resp.Meta.TotalItems)
	}
}

func TestHandler_CreateCategory_ValidationError(t *testing.T) {
	stub := &catalogServiceStub{
		createCategoryFn: func(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error) {
			return catalog.Category{}, catalog.NewValidationError("name is required", nil)
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"name":"Beverages"}`)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/categories", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateCategory(ctx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", rec.Code)
	}

	var resp ValidationErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != string(catalog.ErrorValidation) {
		t.Fatalf("expected code %q, got %q", catalog.ErrorValidation, resp.Code)
	}
}

func TestHandler_CreateCategory_InvalidJSON(t *testing.T) {
	stub := &catalogServiceStub{}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("{"))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateCategory(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	if stub.createCategoryCalls != 0 {
		t.Fatalf("expected service not to be called, got %d calls", stub.createCategoryCalls)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "invalid_payload" {
		t.Fatalf("expected code 'invalid_payload', got %q", resp.Code)
	}
}

func TestHandler_GetCategory_NotFound(t *testing.T) {
	stub := &catalogServiceStub{
		getCategoryFn: func(ctx context.Context, id int) (catalog.Category, error) {
			return catalog.Category{}, catalog.NewNotFoundError("category not found", nil)
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/categories/99", nil)

	handler.GetCategory(ctx, CategoryIdParam(99))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != string(catalog.ErrorNotFound) {
		t.Fatalf("expected code %q, got %q", catalog.ErrorNotFound, resp.Code)
	}
}

func TestHandler_UpdateCategory_Success(t *testing.T) {
	stub := &catalogServiceStub{
		updateCategoryFn: func(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error) {
			return catalog.Category{ID: id, Name: input.Name}, nil
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"name":"Updated"}`)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/categories/1", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateCategory(ctx, CategoryIdParam(1))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp Category
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Id != 1 || resp.Name != "Updated" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandler_DeleteCategory_Success(t *testing.T) {
	stub := &catalogServiceStub{
		deleteCategoryFn: func(ctx context.Context, id int) error {
			return nil
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/categories/2", nil)

	handler.DeleteCategory(ctx, CategoryIdParam(2))
	ctx.Writer.WriteHeaderNow()

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	if stub.deleteCategoryCalls != 1 {
		t.Fatalf("expected delete to be called once, got %d", stub.deleteCategoryCalls)
	}
}

func TestHandler_ListProducts_Success(t *testing.T) {
	stub := &catalogServiceStub{
		listProductsFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error) {
			page := catalog.Pagination{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1}
			items := []catalog.Product{{ID: 10, Name: "Gadget"}}
			return catalog.Page[catalog.Product]{Items: items, Meta: page}, nil
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/products", nil)

	handler.ListProducts(ctx, ListProductsParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp ProductListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Id != 10 {
		t.Fatalf("unexpected response: %+v", resp.Data)
	}
}

func TestHandler_CreateProduct_ValidationError(t *testing.T) {
	stub := &catalogServiceStub{
		createProductFn: func(ctx context.Context, input catalog.ProductInput) (catalog.Product, error) {
			return catalog.Product{}, catalog.NewValidationError("unit price required", nil)
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":10}`)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/products", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProduct(ctx)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", rec.Code)
	}

	var resp ValidationErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != string(catalog.ErrorValidation) {
		t.Fatalf("expected validation error code, got %q", resp.Code)
	}
}

func TestHandler_CreateProduct_InvalidJSON(t *testing.T) {
	stub := &catalogServiceStub{}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("{"))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateProduct(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if stub.createProductCalls != 0 {
		t.Fatalf("expected product creation not to be called, got %d calls", stub.createProductCalls)
	}
}

func TestHandler_DeleteProduct_NotFound(t *testing.T) {
	stub := &catalogServiceStub{
		deleteProductFn: func(ctx context.Context, id int) error {
			return catalog.NewNotFoundError("product not found", nil)
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/products/9", nil)

	handler.DeleteProduct(ctx, ProductIdParam(9))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != string(catalog.ErrorNotFound) {
		t.Fatalf("expected not found code, got %q", resp.Code)
	}
}

func TestHandler_ListSuppliers_Success(t *testing.T) {
	stub := &catalogServiceStub{
		listSuppliersFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error) {
			page := catalog.Pagination{Page: 1, PageSize: 10, TotalItems: 1, TotalPages: 1}
			suppliers := []catalog.Supplier{{ID: 3, CompanyName: "Contoso"}}
			return catalog.Page[catalog.Supplier]{Items: suppliers, Meta: page}, nil
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/suppliers", nil)

	handler.ListSuppliers(ctx, ListSuppliersParams{})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp SupplierListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Id != 3 {
		t.Fatalf("unexpected supplier data: %+v", resp.Data)
	}
}

func TestHandler_CreateSupplier_InternalError(t *testing.T) {
	stub := &catalogServiceStub{
		createSupplierFn: func(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error) {
			return catalog.Supplier{}, errors.New("database offline")
		},
	}

	handler := NewHandler(stub, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"companyName":"Supplier"}`)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/suppliers", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateSupplier(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "internal_error" {
		t.Fatalf("expected internal error code, got %q", resp.Code)
	}
}

func TestHandler_IssueToken_NotImplemented(t *testing.T) {
	handler := NewHandler(&catalogServiceStub{}, nil)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/token", nil)

	handler.IssueToken(ctx)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != "not_implemented" {
		t.Fatalf("expected not_implemented code, got %q", resp.Code)
	}
}

func TestHandler_IssueToken_Success(t *testing.T) {
	issued := false
	tokenSvc := &tokenServiceStub{
		issueTokenFn: func(ctx context.Context, input auth.TokenIssueRequest) (auth.Token, error) {
			issued = true
			if input.Username != "user" {
				t.Fatalf("expected username 'user', got %q", input.Username)
			}
			if input.Password != "pass" {
				t.Fatalf("expected password 'pass', got %q", input.Password)
			}
			if len(input.Scopes) != 1 || input.Scopes[0] != "viewer" {
				t.Fatalf("expected scope 'viewer', got %v", input.Scopes)
			}
			return auth.Token{Value: "token123", ExpiresAt: time.Now().Add(time.Hour), Subject: "user", Scopes: []string{"viewer"}}, nil
		},
	}

	handler := NewHandler(&catalogServiceStub{}, tokenSvc)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"username":"user","password":"pass","scope":"viewer"}`)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/token", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.IssueToken(ctx)

	if !issued {
		t.Fatalf("expected IssueToken to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp TokenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.AccessToken != "token123" {
		t.Fatalf("expected access token 'token123', got %q", resp.AccessToken)
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("expected token type 'Bearer', got %q", resp.TokenType)
	}
	if resp.ExpiresIn <= 0 {
		t.Fatalf("expected positive expiresIn, got %d", resp.ExpiresIn)
	}
}

func TestHandler_IssueToken_InvalidCredentials(t *testing.T) {
	tokenSvc := &tokenServiceStub{
		issueTokenFn: func(ctx context.Context, input auth.TokenIssueRequest) (auth.Token, error) {
			return auth.Token{}, auth.NewError(auth.ErrorInvalidCredentials, "invalid username or password", nil)
		},
	}

	handler := NewHandler(&catalogServiceStub{}, tokenSvc)

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	body := bytes.NewBufferString(`{"username":"user","password":"bad"}`)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/auth/token", body)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.IssueToken(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != string(auth.ErrorInvalidCredentials) {
		t.Fatalf("expected error code %q, got %q", auth.ErrorInvalidCredentials, resp.Code)
	}
}

type catalogServiceStub struct {
	listCategoriesFn    func(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error)
	createCategoryFn    func(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error)
	getCategoryFn       func(ctx context.Context, id int) (catalog.Category, error)
	updateCategoryFn    func(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error)
	deleteCategoryFn    func(ctx context.Context, id int) error
	listProductsFn      func(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error)
	getProductFn        func(ctx context.Context, id int) (catalog.Product, error)
	createProductFn     func(ctx context.Context, input catalog.ProductInput) (catalog.Product, error)
	updateProductFn     func(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error)
	deleteProductFn     func(ctx context.Context, id int) error
	listSuppliersFn     func(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error)
	getSupplierFn       func(ctx context.Context, id int) (catalog.Supplier, error)
	createSupplierFn    func(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error)
	updateSupplierFn    func(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error)
	deleteSupplierFn    func(ctx context.Context, id int) error
	listCategoriesCalls int
	createCategoryCalls int
	getCategoryCalls    int
	deleteCategoryCalls int
	createProductCalls  int
	deleteProductCalls  int
	listProductsCalls   int
	createSupplierCalls int
	listSuppliersCalls  int
	deleteSupplierCalls int
}

func (s *catalogServiceStub) ListCategories(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error) {
	s.listCategoriesCalls++
	if s.listCategoriesFn == nil {
		return catalog.Page[catalog.Category]{}, nil
	}
	return s.listCategoriesFn(ctx, opts, filter)
}

func (s *catalogServiceStub) GetCategory(ctx context.Context, id int) (catalog.Category, error) {
	s.getCategoryCalls++
	if s.getCategoryFn == nil {
		return catalog.Category{}, nil
	}
	return s.getCategoryFn(ctx, id)
}

func (s *catalogServiceStub) CreateCategory(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error) {
	s.createCategoryCalls++
	if s.createCategoryFn == nil {
		return catalog.Category{}, nil
	}
	return s.createCategoryFn(ctx, input)
}

func (s *catalogServiceStub) UpdateCategory(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error) {
	if s.updateCategoryFn == nil {
		return catalog.Category{}, nil
	}
	return s.updateCategoryFn(ctx, id, input)
}

func (s *catalogServiceStub) DeleteCategory(ctx context.Context, id int) error {
	s.deleteCategoryCalls++
	if s.deleteCategoryFn == nil {
		return nil
	}
	return s.deleteCategoryFn(ctx, id)
}

func (s *catalogServiceStub) ListProducts(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error) {
	s.listProductsCalls++
	if s.listProductsFn == nil {
		return catalog.Page[catalog.Product]{}, nil
	}
	return s.listProductsFn(ctx, opts, filter)
}

func (s *catalogServiceStub) GetProduct(ctx context.Context, id int) (catalog.Product, error) {
	if s.getProductFn == nil {
		return catalog.Product{}, nil
	}
	return s.getProductFn(ctx, id)
}

func (s *catalogServiceStub) CreateProduct(ctx context.Context, input catalog.ProductInput) (catalog.Product, error) {
	s.createProductCalls++
	if s.createProductFn == nil {
		return catalog.Product{}, nil
	}
	return s.createProductFn(ctx, input)
}

func (s *catalogServiceStub) UpdateProduct(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error) {
	if s.updateProductFn == nil {
		return catalog.Product{}, nil
	}
	return s.updateProductFn(ctx, id, input)
}

func (s *catalogServiceStub) DeleteProduct(ctx context.Context, id int) error {
	s.deleteProductCalls++
	if s.deleteProductFn == nil {
		return nil
	}
	return s.deleteProductFn(ctx, id)
}

func (s *catalogServiceStub) ListSuppliers(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error) {
	s.listSuppliersCalls++
	if s.listSuppliersFn == nil {
		return catalog.Page[catalog.Supplier]{}, nil
	}
	return s.listSuppliersFn(ctx, opts, filter)
}

func (s *catalogServiceStub) GetSupplier(ctx context.Context, id int) (catalog.Supplier, error) {
	if s.getSupplierFn == nil {
		return catalog.Supplier{}, nil
	}
	return s.getSupplierFn(ctx, id)
}

func (s *catalogServiceStub) CreateSupplier(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error) {
	s.createSupplierCalls++
	if s.createSupplierFn == nil {
		return catalog.Supplier{}, nil
	}
	return s.createSupplierFn(ctx, input)
}

func (s *catalogServiceStub) UpdateSupplier(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error) {
	if s.updateSupplierFn == nil {
		return catalog.Supplier{}, nil
	}
	return s.updateSupplierFn(ctx, id, input)
}

func (s *catalogServiceStub) DeleteSupplier(ctx context.Context, id int) error {
	s.deleteSupplierCalls++
	if s.deleteSupplierFn == nil {
		return nil
	}
	return s.deleteSupplierFn(ctx, id)
}
