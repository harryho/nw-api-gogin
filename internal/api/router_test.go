package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/glb/nw-api-gogin/internal/auth"
	"github.com/glb/nw-api-gogin/internal/catalog"
)

func TestRegisterHandlers_CoversRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	tokenValue := "router-token"
	tokenSvc := &tokenServiceStub{
		issueTokenFn: func(ctx context.Context, input auth.TokenIssueRequest) (auth.Token, error) {
			return auth.Token{Value: tokenValue, ExpiresAt: time.Now().Add(time.Minute)}, nil
		},
		validateTokenFn: func(ctx context.Context, token string) (*auth.Claims, error) {
			if token != tokenValue {
				return nil, auth.NewError(auth.ErrorInvalidToken, "invalid token", nil)
			}
			return &auth.Claims{
				Scopes:           []string{"admin", "manager", "viewer"},
				RegisteredClaims: jwt.RegisteredClaims{Subject: "router-user"},
			}, nil
		},
	}
	handler := NewHandler(&fullCatalogStub{}, tokenSvc)
	RegisterHandlersWithOptions(router, handler, GinServerOptions{
		Middlewares: []MiddlewareFunc{AuthMiddleware(tokenSvc)},
	})

	testCases := []struct {
		method       string
		path         string
		body         string
		status       int
		requiresAuth bool
	}{
		{method: http.MethodGet, path: "/categories", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodGet, path: "/categories?page=abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodGet, path: "/categories?pageSize=abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPost, path: "/categories", body: `{"name":"New"}`, status: http.StatusCreated, requiresAuth: true},
		{method: http.MethodGet, path: "/categories/1", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodGet, path: "/categories/not-a-number", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/categories/not-a-number", body: `{"name":"Updated"}`, status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodDelete, path: "/categories/not-a-number", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/categories/1", body: `{"name":"Updated"}`, status: http.StatusOK, requiresAuth: true},
		{method: http.MethodDelete, path: "/categories/1", status: http.StatusNoContent, requiresAuth: true},
		{method: http.MethodGet, path: "/products", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodGet, path: "/products?categoryId=abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodGet, path: "/products?supplierId=abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodGet, path: "/products?discontinued=maybe", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPost, path: "/products", body: `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":10}`, status: http.StatusCreated, requiresAuth: true},
		{method: http.MethodGet, path: "/products/1", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodGet, path: "/products/abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/products/abc", body: `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":12}`, status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodDelete, path: "/products/abc", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/products/1", body: `{"name":"Prod","categoryId":1,"supplierId":1,"unitPrice":12}`, status: http.StatusOK, requiresAuth: true},
		{method: http.MethodDelete, path: "/products/1", status: http.StatusNoContent, requiresAuth: true},
		{method: http.MethodGet, path: "/suppliers", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodPost, path: "/suppliers", body: `{"companyName":"Supplier"}`, status: http.StatusCreated, requiresAuth: true},
		{method: http.MethodGet, path: "/suppliers/1", status: http.StatusOK, requiresAuth: true},
		{method: http.MethodGet, path: "/suppliers/bad", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/suppliers/bad", body: `{"companyName":"Supplier"}`, status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodDelete, path: "/suppliers/bad", status: http.StatusBadRequest, requiresAuth: true},
		{method: http.MethodPut, path: "/suppliers/1", body: `{"companyName":"Supplier"}`, status: http.StatusOK, requiresAuth: true},
		{method: http.MethodDelete, path: "/suppliers/1", status: http.StatusNoContent, requiresAuth: true},
		{method: http.MethodPost, path: "/auth/token", body: `{"username":"u","password":"p","scope":"viewer"}`, status: http.StatusOK, requiresAuth: false},
		{method: http.MethodGet, path: "/categories", status: http.StatusUnauthorized, requiresAuth: false},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if tc.requiresAuth {
			req.Header.Set("Authorization", "Bearer "+tokenValue)
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != tc.status {
			t.Fatalf("%s %s: expected status %d, got %d", tc.method, tc.path, tc.status, resp.Code)
		}
		if tc.status != http.StatusNoContent {
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

	handler := NewHandler(failStub, nil)
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
