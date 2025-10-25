package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

	handler := NewHandler(stub)

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

type catalogServiceStub struct {
	listCategoriesFn    func(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error)
	createCategoryFn    func(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error)
	getCategoryFn       func(ctx context.Context, id int) (catalog.Category, error)
	updateCategoryFn    func(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error)
	deleteCategoryFn    func(ctx context.Context, id int) error
	listProductsFn      func(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error)
	listCategoriesCalls int
	createCategoryCalls int
	getCategoryCalls    int
	deleteCategoryCalls int
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
	if s.listProductsFn == nil {
		return catalog.Page[catalog.Product]{}, nil
	}
	return s.listProductsFn(ctx, opts, filter)
}

func (s *catalogServiceStub) GetProduct(ctx context.Context, id int) (catalog.Product, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) CreateProduct(ctx context.Context, input catalog.ProductInput) (catalog.Product, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) UpdateProduct(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) DeleteProduct(ctx context.Context, id int) error {
	panic("not implemented")
}

func (s *catalogServiceStub) ListSuppliers(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error) {
	panic("not implemented")
}

func (s *catalogServiceStub) GetSupplier(ctx context.Context, id int) (catalog.Supplier, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) CreateSupplier(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) UpdateSupplier(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error) {
	panic("not implemented")
}

func (s *catalogServiceStub) DeleteSupplier(ctx context.Context, id int) error {
	panic("not implemented")
}
