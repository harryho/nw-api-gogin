package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) IssueToken(c *gin.Context) {
	notImplemented(c)
}

func (h *Handler) ListCategories(c *gin.Context, _ ListCategoriesParams) {
	notImplemented(c)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	notImplemented(c)
}

func (h *Handler) DeleteCategory(c *gin.Context, _ CategoryIdParam) {
	notImplemented(c)
}

func (h *Handler) GetCategory(c *gin.Context, _ CategoryIdParam) {
	notImplemented(c)
}

func (h *Handler) UpdateCategory(c *gin.Context, _ CategoryIdParam) {
	notImplemented(c)
}

func (h *Handler) ListProducts(c *gin.Context, _ ListProductsParams) {
	notImplemented(c)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	notImplemented(c)
}

func (h *Handler) DeleteProduct(c *gin.Context, _ ProductIdParam) {
	notImplemented(c)
}

func (h *Handler) GetProduct(c *gin.Context, _ ProductIdParam) {
	notImplemented(c)
}

func (h *Handler) UpdateProduct(c *gin.Context, _ ProductIdParam) {
	notImplemented(c)
}

func (h *Handler) ListSuppliers(c *gin.Context, _ ListSuppliersParams) {
	notImplemented(c)
}

func (h *Handler) CreateSupplier(c *gin.Context) {
	notImplemented(c)
}

func (h *Handler) DeleteSupplier(c *gin.Context, _ SupplierIdParam) {
	notImplemented(c)
}

func (h *Handler) GetSupplier(c *gin.Context, _ SupplierIdParam) {
	notImplemented(c)
}

func (h *Handler) UpdateSupplier(c *gin.Context, _ SupplierIdParam) {
	notImplemented(c)
}

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Code: "not_implemented", Message: "not implemented"})
}
