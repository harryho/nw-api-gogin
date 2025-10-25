package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/glb/nw-api-gogin/internal/catalog"
	httpmw "github.com/glb/nw-api-gogin/internal/http/middleware"
)

type CatalogService interface {
	ListCategories(ctx context.Context, opts catalog.ListOptions, filter catalog.CategoryFilter) (catalog.Page[catalog.Category], error)
	GetCategory(ctx context.Context, id int) (catalog.Category, error)
	CreateCategory(ctx context.Context, input catalog.CategoryInput) (catalog.Category, error)
	UpdateCategory(ctx context.Context, id int, input catalog.CategoryInput) (catalog.Category, error)
	DeleteCategory(ctx context.Context, id int) error

	ListProducts(ctx context.Context, opts catalog.ListOptions, filter catalog.ProductFilter) (catalog.Page[catalog.Product], error)
	GetProduct(ctx context.Context, id int) (catalog.Product, error)
	CreateProduct(ctx context.Context, input catalog.ProductInput) (catalog.Product, error)
	UpdateProduct(ctx context.Context, id int, input catalog.ProductInput) (catalog.Product, error)
	DeleteProduct(ctx context.Context, id int) error

	ListSuppliers(ctx context.Context, opts catalog.ListOptions, filter catalog.SupplierFilter) (catalog.Page[catalog.Supplier], error)
	GetSupplier(ctx context.Context, id int) (catalog.Supplier, error)
	CreateSupplier(ctx context.Context, input catalog.SupplierInput) (catalog.Supplier, error)
	UpdateSupplier(ctx context.Context, id int, input catalog.SupplierInput) (catalog.Supplier, error)
	DeleteSupplier(ctx context.Context, id int) error
}

type Handler struct {
	catalog CatalogService
}

func NewHandler(catalogService CatalogService) *Handler {
	return &Handler{catalog: catalogService}
}

func (h *Handler) IssueToken(c *gin.Context) {
	h.respondNotImplemented(c)
}

func (h *Handler) ListCategories(c *gin.Context, params ListCategoriesParams) {
	opts := listOptions(params.Page, params.PageSize, params.Sort)
	filter := catalog.CategoryFilter{Name: params.Name}

	page, err := h.catalog.ListCategories(c.Request.Context(), opts, filter)
	if err != nil {
		h.respondError(c, err)
		return
	}

	resp := CategoryListResponse{
		Data: mapCategories(page.Items),
		Meta: toPaginationMeta(page.Meta),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	category, err := h.catalog.CreateCategory(c.Request.Context(), catalog.CategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toCategory(category))
}

func (h *Handler) DeleteCategory(c *gin.Context, id CategoryIdParam) {
	if err := h.catalog.DeleteCategory(c.Request.Context(), int(id)); err != nil {
		h.respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetCategory(c *gin.Context, id CategoryIdParam) {
	category, err := h.catalog.GetCategory(c.Request.Context(), int(id))
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCategory(category))
}

func (h *Handler) UpdateCategory(c *gin.Context, id CategoryIdParam) {
	var req CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	category, err := h.catalog.UpdateCategory(c.Request.Context(), int(id), catalog.CategoryInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCategory(category))
}

func (h *Handler) ListProducts(c *gin.Context, params ListProductsParams) {
	opts := listOptions(params.Page, params.PageSize, params.Sort)
	filter := catalog.ProductFilter{
		Name:         params.Name,
		CategoryID:   params.CategoryId,
		SupplierID:   params.SupplierId,
		Discontinued: params.Discontinued,
	}

	page, err := h.catalog.ListProducts(c.Request.Context(), opts, filter)
	if err != nil {
		h.respondError(c, err)
		return
	}

	resp := ProductListResponse{
		Data: mapProducts(page.Items),
		Meta: toPaginationMeta(page.Meta),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	product, err := h.catalog.CreateProduct(c.Request.Context(), toProductInput(req))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toProduct(product))
}

func (h *Handler) DeleteProduct(c *gin.Context, id ProductIdParam) {
	if err := h.catalog.DeleteProduct(c.Request.Context(), int(id)); err != nil {
		h.respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetProduct(c *gin.Context, id ProductIdParam) {
	product, err := h.catalog.GetProduct(c.Request.Context(), int(id))
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProduct(product))
}

func (h *Handler) UpdateProduct(c *gin.Context, id ProductIdParam) {
	var req ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	product, err := h.catalog.UpdateProduct(c.Request.Context(), int(id), toProductInput(req))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toProduct(product))
}

func (h *Handler) ListSuppliers(c *gin.Context, params ListSuppliersParams) {
	opts := listOptions(params.Page, params.PageSize, params.Sort)
	filter := catalog.SupplierFilter{
		CompanyName: params.CompanyName,
		Country:     params.Country,
	}

	page, err := h.catalog.ListSuppliers(c.Request.Context(), opts, filter)
	if err != nil {
		h.respondError(c, err)
		return
	}

	resp := SupplierListResponse{
		Data: mapSuppliers(page.Items),
		Meta: toPaginationMeta(page.Meta),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateSupplier(c *gin.Context) {
	var req SupplierCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	supplier, err := h.catalog.CreateSupplier(c.Request.Context(), toSupplierInput(req))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSupplier(supplier))
}

func (h *Handler) DeleteSupplier(c *gin.Context, id SupplierIdParam) {
	if err := h.catalog.DeleteSupplier(c.Request.Context(), int(id)); err != nil {
		h.respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetSupplier(c *gin.Context, id SupplierIdParam) {
	supplier, err := h.catalog.GetSupplier(c.Request.Context(), int(id))
	if err != nil {
		h.respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toSupplier(supplier))
}

func (h *Handler) UpdateSupplier(c *gin.Context, id SupplierIdParam) {
	var req SupplierUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.decodeError(c)
		return
	}

	supplier, err := h.catalog.UpdateSupplier(c.Request.Context(), int(id), toSupplierInput(req))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSupplier(supplier))
}

func (h *Handler) decodeError(c *gin.Context) {
	traceID := httpmw.RequestIDFromContext(c.Request.Context())
	resp := ErrorResponse{Code: "invalid_payload", Message: "invalid request payload"}
	if traceID != "" {
		resp.TraceId = &traceID
	}
	c.JSON(http.StatusBadRequest, resp)
}

func (h *Handler) respondNotImplemented(c *gin.Context) {
	traceID := httpmw.RequestIDFromContext(c.Request.Context())
	resp := ErrorResponse{Code: "not_implemented", Message: "not implemented"}
	if traceID != "" {
		resp.TraceId = &traceID
	}
	c.JSON(http.StatusNotImplemented, resp)
}

func (h *Handler) respondError(c *gin.Context, err error) {
	traceID := httpmw.RequestIDFromContext(c.Request.Context())

	if appErr, ok := catalog.AsError(err); ok {
		status := appErr.Status
		if status == 0 {
			status = http.StatusInternalServerError
		}

		switch appErr.Code {
		case catalog.ErrorValidation:
			resp := ValidationErrorResponse{
				Code:    string(appErr.Code),
				Message: appErr.Message,
			}
			if traceID != "" {
				resp.TraceId = &traceID
			}
			c.JSON(status, resp)
		default:
			resp := ErrorResponse{Code: string(appErr.Code), Message: appErr.Message}
			if traceID != "" {
				resp.TraceId = &traceID
			}
			c.JSON(status, resp)
		}
		return
	}

	resp := ErrorResponse{Code: "internal_error", Message: "internal server error"}
	if traceID != "" {
		resp.TraceId = &traceID
	}
	c.JSON(http.StatusInternalServerError, resp)
}

func listOptions(page *PageParam, pageSize *PageSizeParam, sort *SortParam) catalog.ListOptions {
	opts := catalog.ListOptions{}
	if page != nil {
		opts.Page = *page
	}
	if pageSize != nil {
		opts.PageSize = *pageSize
	}
	if sort != nil {
		opts.Sort = *sort
	}
	return opts
}

func toPaginationMeta(meta catalog.Pagination) PaginationMeta {
	return PaginationMeta{
		Page:       meta.Page,
		PageSize:   meta.PageSize,
		TotalItems: meta.TotalItems,
		TotalPages: meta.TotalPages,
	}
}

func mapCategories(categories []catalog.Category) []Category {
	out := make([]Category, len(categories))
	for i, category := range categories {
		out[i] = toCategory(category)
	}
	return out
}

func mapProducts(products []catalog.Product) []Product {
	out := make([]Product, len(products))
	for i, product := range products {
		out[i] = toProduct(product)
	}
	return out
}

func mapSuppliers(suppliers []catalog.Supplier) []Supplier {
	out := make([]Supplier, len(suppliers))
	for i, supplier := range suppliers {
		out[i] = toSupplier(supplier)
	}
	return out
}

func toCategory(category catalog.Category) Category {
	return Category{
		Id:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

func toProduct(product catalog.Product) Product {
	return Product{
		Id:              product.ID,
		CategoryId:      product.CategoryID,
		SupplierId:      product.SupplierID,
		Name:            product.Name,
		QuantityPerUnit: product.QuantityPerUnit,
		UnitPrice:       product.UnitPrice,
		UnitsInStock:    product.UnitsInStock,
		UnitsOnOrder:    toOptionalInt(product.UnitsOnOrder),
		ReorderLevel:    toOptionalInt(product.ReorderLevel),
		Discontinued:    product.Discontinued,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

func toSupplier(supplier catalog.Supplier) Supplier {
	return Supplier{
		Id:           supplier.ID,
		CompanyName:  supplier.CompanyName,
		ContactName:  supplier.ContactName,
		ContactTitle: supplier.ContactTitle,
		Address:      supplier.Address,
		City:         supplier.City,
		Region:       supplier.Region,
		PostalCode:   supplier.PostalCode,
		Country:      supplier.Country,
		Phone:        supplier.Phone,
		CreatedAt:    supplier.CreatedAt,
		UpdatedAt:    supplier.UpdatedAt,
	}
}

func toProductInput(req ProductCreateRequest) catalog.ProductInput {
	discontinued := false
	if req.Discontinued != nil {
		discontinued = *req.Discontinued
	}

	unitsInStock := 0
	if req.UnitsInStock != nil {
		unitsInStock = *req.UnitsInStock
	}

	unitsOnOrder := 0
	if req.UnitsOnOrder != nil {
		unitsOnOrder = *req.UnitsOnOrder
	}

	reorderLevel := 0
	if req.ReorderLevel != nil {
		reorderLevel = *req.ReorderLevel
	}

	return catalog.ProductInput{
		CategoryID:      req.CategoryId,
		SupplierID:      req.SupplierId,
		Name:            req.Name,
		QuantityPerUnit: req.QuantityPerUnit,
		UnitPrice:       req.UnitPrice,
		UnitsInStock:    unitsInStock,
		UnitsOnOrder:    unitsOnOrder,
		ReorderLevel:    reorderLevel,
		Discontinued:    discontinued,
	}
}

func toSupplierInput(req SupplierCreateRequest) catalog.SupplierInput {
	return catalog.SupplierInput{
		CompanyName:  req.CompanyName,
		ContactName:  req.ContactName,
		ContactTitle: req.ContactTitle,
		Address:      req.Address,
		City:         req.City,
		Region:       req.Region,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		Phone:        req.Phone,
	}
}

func toOptionalInt(value int) *int {
	return &value
}
