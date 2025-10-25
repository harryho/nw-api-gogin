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

func TestRegisterHandlers_CoversRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := NewHandler(&fullCatalogStub{})
	RegisterHandlers(router, handler)

	testCases := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/categories", "", http.StatusOK},
		{http.MethodGet, "/categories?page=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/categories?pageSize=abc", "", http.StatusBadRequest},
		{http.MethodPost, "/categories", `{"name":"New"}`, http.StatusCreated},
		{http.MethodGet, "/categories/1", "", http.StatusOK},
		{http.MethodGet, "/categories/not-a-number", "", http.StatusBadRequest},
		{http.MethodPut, "/categories/not-a-number", `{"name":"Updated"}`, http.StatusBadRequest},
		{http.MethodDelete, "/categories/not-a-number", "", http.StatusBadRequest},
		{http.MethodPut, "/categories/1", `{"name":"Updated"}`, http.StatusOK},
		{http.MethodDelete, "/categories/1", "", http.StatusNoContent},
		{http.MethodGet, "/products", "", http.StatusOK},
		{http.MethodGet, "/products?categoryId=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/products?supplierId=abc", "", http.StatusBadRequest},
		{http.MethodGet, "/products?discontinued=maybe", "", http.StatusBadRequest},
		{http.MethodPost, "/products", `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":10}`, http.StatusCreated},
		{http.MethodGet, "/products/1", "", http.StatusOK},
		{http.MethodGet, "/products/abc", "", http.StatusBadRequest},
		{http.MethodPut, "/products/abc", `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":12}`, http.StatusBadRequest},
		{http.MethodDelete, "/products/abc", "", http.StatusBadRequest},
		{http.MethodPut, "/products/1", `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":12}`, http.StatusOK},
		{http.MethodDelete, "/products/1", "", http.StatusNoContent},
		{http.MethodGet, "/suppliers", "", http.StatusOK},
		{http.MethodPost, "/suppliers", `{"companyName":"Supplier"}`, http.StatusCreated},
		{http.MethodGet, "/suppliers/1", "", http.StatusOK},
		{http.MethodGet, "/suppliers/bad", "", http.StatusBadRequest},
		{http.MethodPut, "/suppliers/bad", `{"companyName":"Supplier"}`, http.StatusBadRequest},
		{http.MethodDelete, "/suppliers/bad", "", http.StatusBadRequest},
		{http.MethodPut, "/suppliers/1", `{"companyName":"Supplier"}`, http.StatusOK},
		{http.MethodDelete, "/suppliers/1", "", http.StatusNoContent},
		{http.MethodPost, "/auth/token", `{"username":"u","password":"p","scope":"viewer"}`, http.StatusNotImplemented},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tc.status {
			t.Fatalf("%s %s: expected status %d, got %d", tc.method, tc.path, tc.status, resp.Code)
		}
		if tc.status != http.StatusNoContent && tc.status != http.StatusNotImplemented {
			var data map[string]any
			_ = json.Unmarshal(resp.Body.Bytes(), &data)
		}
	}
}

type fullCatalogStub struct{}

func (s *fullCatalogStub) ListCategories(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error) {
	page := catalog.Pagination{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1}
	items := []catalog.Category{{ID: 1, Name: "Cat"}}
	return catalog.Page[catalog.Category]{Items: items, Meta: page}, nil
}

func (s *fullCatalogStub) GetCategory(ctx context.Context, id int) (catalog.Category, error) {
	return catalog.Category{ID: id, Name: "Cat"}, nil
}

func (s *fullCatalogStub) CreateCategory(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error) {
	return catalog.Category{ID: 1, Name: input.Name}, nil
}

func (s *fullCatalogStub) UpdateCategory(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error) {
	return catalog.Category{ID: id, Name: input.Name}, nil
}

func (s *fullCatalogStub) DeleteCategory(ctx context.Context, id int) error {
	return nil
}

func (s *fullCatalogStub) ListProducts(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error) {
	page := catalog.Pagination{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1}
	items := []catalog.Product{{ID: 1, Name: "Prod"}}
	return catalog.Page[catalog.Product]{Items: items, Meta: page}, nil
}

func (s *fullCatalogStub) GetProduct(ctx context.Context, id int) (catalog.Product, error) {
	return catalog.Product{ID: id, Name: "Prod"}, nil
}

func (s *fullCatalogStub) CreateProduct(ctx context.Context, input catalog.ProductInput) (catalog.Product, error) {
	return catalog.Product{ID: 1, Name: input.Name}, nil
}

func (s *fullCatalogStub) UpdateProduct(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error) {
	return catalog.Product{ID: id, Name: input.Name}, nil
}

func (s *fullCatalogStub) DeleteProduct(ctx context.Context, id int) error {
	return nil
}

func (s *fullCatalogStub) ListSuppliers(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error) {
	page := catalog.Pagination{Page: 1, PageSize: 20, TotalItems: 1, TotalPages: 1}
	items := []catalog.Supplier{{ID: 1, CompanyName: "Supplier"}}
	return catalog.Page[catalog.Supplier]{Items: items, Meta: page}, nil
}

func (s *fullCatalogStub) GetSupplier(ctx context.Context, id int) (catalog.Supplier, error) {
	return catalog.Supplier{ID: id, CompanyName: "Supplier"}, nil
}

func (s *fullCatalogStub) CreateSupplier(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error) {
	return catalog.Supplier{ID: 1, CompanyName: input.CompanyName}, nil
}

func (s *fullCatalogStub) UpdateSupplier(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error) {
	return catalog.Supplier{ID: id, CompanyName: input.CompanyName}, nil
}

func (s *fullCatalogStub) DeleteSupplier(ctx context.Context, id int) error {
	return nil
}

func TestRegisterHandlersWithOptions_MiddlewareAbort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	failStub := &catalogServiceStub{
		listCategoriesFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error) {
			t.Fatalf("list categories should not be called")
			return catalog.Page[catalog.Category]{}, nil
		},
		createCategoryFn: func(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error) {
			t.Fatalf("create category should not be called")
			return catalog.Category{}, nil
		},
		deleteCategoryFn: func(ctx context.Context, id int) error {
			t.Fatalf("delete category should not be called")
			return nil
		},
		getCategoryFn: func(ctx context.Context, id int) (catalog.Category, error) {
			t.Fatalf("get category should not be called")
			return catalog.Category{}, nil
		},
		updateCategoryFn: func(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error) {
			t.Fatalf("update category should not be called")
			return catalog.Category{}, nil
		},
		listProductsFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error) {
			t.Fatalf("list products should not be called")
			return catalog.Page[catalog.Product]{}, nil
		},
		createProductFn: func(ctx context.Context, input catalog.ProductInput) (catalog.Product, error) {
			t.Fatalf("create product should not be called")
			return catalog.Product{}, nil
		},
		deleteProductFn: func(ctx context.Context, id int) error {
			t.Fatalf("delete product should not be called")
			return nil
		},
		getProductFn: func(ctx context.Context, id int) (catalog.Product, error) {
			t.Fatalf("get product should not be called")
			return catalog.Product{}, nil
		},
		updateProductFn: func(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error) {
			t.Fatalf("update product should not be called")
			return catalog.Product{}, nil
		},
		listSuppliersFn: func(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error) {
			t.Fatalf("list suppliers should not be called")
			return catalog.Page[catalog.Supplier]{}, nil
		},
		createSupplierFn: func(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error) {
			t.Fatalf("create supplier should not be called")
			return catalog.Supplier{}, nil
		},
		deleteSupplierFn: func(ctx context.Context, id int) error {
			t.Fatalf("delete supplier should not be called")
			return nil
		},
		getSupplierFn: func(ctx context.Context, id int) (catalog.Supplier, error) {
			t.Fatalf("get supplier should not be called")
			return catalog.Supplier{}, nil
		},
		updateSupplierFn: func(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error) {
			t.Fatalf("update supplier should not be called")
			return catalog.Supplier{}, nil
		},
	}

	handler := NewHandler(failStub)
	router := gin.New()
	RegisterHandlersWithOptions(router, handler, GinServerOptions{
		Middlewares: []MiddlewareFunc{
			func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusTeapot, gin.H{"msg": "blocked by middleware"})
			},
		},
	})

	testCases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/categories", ""},
		{http.MethodPost, "/categories", `{"name":"New"}`},
		{http.MethodDelete, "/categories/1", ""},
		{http.MethodGet, "/categories/1", ""},
		{http.MethodPut, "/categories/1", `{"name":"Updated"}`},
		{http.MethodGet, "/products", ""},
		{http.MethodPost, "/products", `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":10}`},
		{http.MethodDelete, "/products/1", ""},
		{http.MethodGet, "/products/1", ""},
		{http.MethodPut, "/products/1", `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":12}`},
		{http.MethodGet, "/suppliers", ""},
		{http.MethodPost, "/suppliers", `{"companyName":"Supplier"}`},
		{http.MethodDelete, "/suppliers/1", ""},
		{http.MethodGet, "/suppliers/1", ""},
		{http.MethodPut, "/suppliers/1", `{"companyName":"Supplier"}`},
		{http.MethodPost, "/auth/token", `{"username":"u","password":"p","scope":"viewer"}`},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusTeapot {
			t.Fatalf("%s %s: expected middleware to stop request with status %d, got %d", tc.method, tc.path, http.StatusTeapot, resp.Code)
		}
	}
}
