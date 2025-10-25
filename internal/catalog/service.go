package catalog

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, logger: log}
}

func (s *Service) ListCategories(ctx context.Context, opts ListOptions, filter CategoryFilter) (Page[Category], error) {
	page, err := s.repo.ListCategories(ctx, opts, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidSortField) {
			return Page[Category]{}, NewValidationError("invalid sort field", err)
		}
		s.loggerForContext(ctx).Error("list categories failed", zap.Error(err))
		return Page[Category]{}, NewInternalError("failed to list categories", err)
	}
	return page, nil
}

func (s *Service) GetCategory(ctx context.Context, id int) (Category, error) {
	category, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Category{}, NewNotFoundError("category not found", err)
		}
		s.loggerForContext(ctx).Error("get category failed", zap.Int("id", id), zap.Error(err))
		return Category{}, NewInternalError("failed to fetch category", err)
	}
	return category, nil
}

func (s *Service) CreateCategory(ctx context.Context, input CategoryInput) (Category, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Category{}, NewValidationError("name is required", nil)
	}
	category := Category{Name: strings.TrimSpace(input.Name), Description: trimPointer(input.Description)}
	if err := s.repo.SaveCategory(ctx, &category); err != nil {
		if isUniqueViolation(err) {
			return Category{}, NewConflictError("category name already exists", err)
		}
		s.loggerForContext(ctx).Error("create category failed", zap.Error(err))
		return Category{}, NewInternalError("failed to create category", err)
	}
	return category, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id int, input CategoryInput) (Category, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Category{}, NewValidationError("name is required", nil)
	}
	category, err := s.repo.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Category{}, NewNotFoundError("category not found", err)
		}
		s.loggerForContext(ctx).Error("load category for update failed", zap.Int("id", id), zap.Error(err))
		return Category{}, NewInternalError("failed to load category", err)
	}

	category.Name = strings.TrimSpace(input.Name)
	category.Description = trimPointer(input.Description)

	if err := s.repo.SaveCategory(ctx, &category); err != nil {
		if isUniqueViolation(err) {
			return Category{}, NewConflictError("category name already exists", err)
		}
		s.loggerForContext(ctx).Error("update category failed", zap.Int("id", id), zap.Error(err))
		return Category{}, NewInternalError("failed to update category", err)
	}
	return category, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id int) error {
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("category not found", err)
		}
		s.loggerForContext(ctx).Error("delete category failed", zap.Int("id", id), zap.Error(err))
		return NewInternalError("failed to delete category", err)
	}
	return nil
}

func (s *Service) ListProducts(ctx context.Context, opts ListOptions, filter ProductFilter) (Page[Product], error) {
	page, err := s.repo.ListProducts(ctx, opts, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidSortField) {
			return Page[Product]{}, NewValidationError("invalid sort field", err)
		}
		s.loggerForContext(ctx).Error("list products failed", zap.Error(err))
		return Page[Product]{}, NewInternalError("failed to list products", err)
	}
	return page, nil
}

func (s *Service) GetProduct(ctx context.Context, id int) (Product, error) {
	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewNotFoundError("product not found", err)
		}
		s.loggerForContext(ctx).Error("get product failed", zap.Int("id", id), zap.Error(err))
		return Product{}, NewInternalError("failed to fetch product", err)
	}
	return product, nil
}

func (s *Service) CreateProduct(ctx context.Context, input ProductInput) (Product, error) {
	if err := validateProductInput(input); err != nil {
		return Product{}, err
	}

	if _, err := s.repo.GetCategory(ctx, input.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewValidationError("invalid category id", err)
		}
		return Product{}, NewInternalError("failed to verify category", err)
	}

	if _, err := s.repo.GetSupplier(ctx, input.SupplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewValidationError("invalid supplier id", err)
		}
		return Product{}, NewInternalError("failed to verify supplier", err)
	}

	product := Product{
		CategoryID:      input.CategoryID,
		SupplierID:      input.SupplierID,
		Name:            strings.TrimSpace(input.Name),
		QuantityPerUnit: trimPointer(input.QuantityPerUnit),
		UnitPrice:       input.UnitPrice,
		UnitsInStock:    input.UnitsInStock,
		UnitsOnOrder:    input.UnitsOnOrder,
		ReorderLevel:    input.ReorderLevel,
		Discontinued:    input.Discontinued,
	}

	if err := s.repo.SaveProduct(ctx, &product); err != nil {
		s.loggerForContext(ctx).Error("create product failed", zap.Error(err))
		return Product{}, NewInternalError("failed to create product", err)
	}
	return product, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id int, input ProductInput) (Product, error) {
	if err := validateProductInput(input); err != nil {
		return Product{}, err
	}

	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewNotFoundError("product not found", err)
		}
		return Product{}, NewInternalError("failed to load product", err)
	}

	if _, err := s.repo.GetCategory(ctx, input.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewValidationError("invalid category id", err)
		}
		return Product{}, NewInternalError("failed to verify category", err)
	}

	if _, err := s.repo.GetSupplier(ctx, input.SupplierID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, NewValidationError("invalid supplier id", err)
		}
		return Product{}, NewInternalError("failed to verify supplier", err)
	}

	product.CategoryID = input.CategoryID
	product.SupplierID = input.SupplierID
	product.Name = strings.TrimSpace(input.Name)
	product.QuantityPerUnit = trimPointer(input.QuantityPerUnit)
	product.UnitPrice = input.UnitPrice
	product.UnitsInStock = input.UnitsInStock
	product.UnitsOnOrder = input.UnitsOnOrder
	product.ReorderLevel = input.ReorderLevel
	product.Discontinued = input.Discontinued

	if err := s.repo.SaveProduct(ctx, &product); err != nil {
		s.loggerForContext(ctx).Error("update product failed", zap.Int("id", id), zap.Error(err))
		return Product{}, NewInternalError("failed to update product", err)
	}
	return product, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int) error {
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("product not found", err)
		}
		s.loggerForContext(ctx).Error("delete product failed", zap.Int("id", id), zap.Error(err))
		return NewInternalError("failed to delete product", err)
	}
	return nil
}

func (s *Service) ListSuppliers(ctx context.Context, opts ListOptions, filter SupplierFilter) (Page[Supplier], error) {
	page, err := s.repo.ListSuppliers(ctx, opts, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidSortField) {
			return Page[Supplier]{}, NewValidationError("invalid sort field", err)
		}
		s.loggerForContext(ctx).Error("list suppliers failed", zap.Error(err))
		return Page[Supplier]{}, NewInternalError("failed to list suppliers", err)
	}
	return page, nil
}

func (s *Service) GetSupplier(ctx context.Context, id int) (Supplier, error) {
	supplier, err := s.repo.GetSupplier(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Supplier{}, NewNotFoundError("supplier not found", err)
		}
		s.loggerForContext(ctx).Error("get supplier failed", zap.Int("id", id), zap.Error(err))
		return Supplier{}, NewInternalError("failed to fetch supplier", err)
	}
	return supplier, nil
}

func (s *Service) CreateSupplier(ctx context.Context, input SupplierInput) (Supplier, error) {
	if strings.TrimSpace(input.CompanyName) == "" {
		return Supplier{}, NewValidationError("companyName is required", nil)
	}
	supplier := Supplier{
		CompanyName:  strings.TrimSpace(input.CompanyName),
		ContactName:  trimPointer(input.ContactName),
		ContactTitle: trimPointer(input.ContactTitle),
		Address:      trimPointer(input.Address),
		City:         trimPointer(input.City),
		Region:       trimPointer(input.Region),
		PostalCode:   trimPointer(input.PostalCode),
		Country:      trimPointer(input.Country),
		Phone:        trimPointer(input.Phone),
	}

	if err := s.repo.SaveSupplier(ctx, &supplier); err != nil {
		s.loggerForContext(ctx).Error("create supplier failed", zap.Error(err))
		return Supplier{}, NewInternalError("failed to create supplier", err)
	}
	return supplier, nil
}

func (s *Service) UpdateSupplier(ctx context.Context, id int, input SupplierInput) (Supplier, error) {
	if strings.TrimSpace(input.CompanyName) == "" {
		return Supplier{}, NewValidationError("companyName is required", nil)
	}
	supplier, err := s.repo.GetSupplier(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Supplier{}, NewNotFoundError("supplier not found", err)
		}
		s.loggerForContext(ctx).Error("load supplier for update failed", zap.Int("id", id), zap.Error(err))
		return Supplier{}, NewInternalError("failed to load supplier", err)
	}

	supplier.CompanyName = strings.TrimSpace(input.CompanyName)
	supplier.ContactName = trimPointer(input.ContactName)
	supplier.ContactTitle = trimPointer(input.ContactTitle)
	supplier.Address = trimPointer(input.Address)
	supplier.City = trimPointer(input.City)
	supplier.Region = trimPointer(input.Region)
	supplier.PostalCode = trimPointer(input.PostalCode)
	supplier.Country = trimPointer(input.Country)
	supplier.Phone = trimPointer(input.Phone)

	if err := s.repo.SaveSupplier(ctx, &supplier); err != nil {
		s.loggerForContext(ctx).Error("update supplier failed", zap.Int("id", id), zap.Error(err))
		return Supplier{}, NewInternalError("failed to update supplier", err)
	}
	return supplier, nil
}

func (s *Service) DeleteSupplier(ctx context.Context, id int) error {
	if err := s.repo.DeleteSupplier(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewNotFoundError("supplier not found", err)
		}
		s.loggerForContext(ctx).Error("delete supplier failed", zap.Int("id", id), zap.Error(err))
		return NewInternalError("failed to delete supplier", err)
	}
	return nil
}

func (s *Service) loggerForContext(ctx context.Context) *zap.Logger {
	if s.logger == nil {
		return zap.NewNop()
	}
	if ctx == nil {
		return s.logger
	}
	return s.logger.With(zap.String("component", "catalog"))
}

func validateProductInput(input ProductInput) *Error {
	if strings.TrimSpace(input.Name) == "" {
		return NewValidationError("name is required", nil)
	}
	if input.CategoryID <= 0 {
		return NewValidationError("categoryId must be positive", nil)
	}
	if input.SupplierID <= 0 {
		return NewValidationError("supplierId must be positive", nil)
	}
	if input.UnitPrice < 0 {
		return NewValidationError("unitPrice must be non-negative", nil)
	}
	if input.UnitsInStock < 0 || input.UnitsOnOrder < 0 || input.ReorderLevel < 0 {
		return NewValidationError("inventory counts must be non-negative", nil)
	}
	return nil
}

func trimPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
