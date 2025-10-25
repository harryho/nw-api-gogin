package catalog

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(ctx context.Context, fn func(repo *Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &Repository{db: tx}
		return fn(txRepo)
	})
}

func (r *Repository) ListCategories(ctx context.Context, opts ListOptions, filter CategoryFilter) (Page[Category], error) {
	query := r.db.WithContext(ctx).Model(&Category{})

	if filter.Name != nil && strings.TrimSpace(*filter.Name) != "" {
		query = query.Where("LOWER(name) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(strings.TrimSpace(*filter.Name))))
	}

	return list[Category](query, opts, map[string]string{
		"name":       "name",
		"created_at": "created_at",
	})
}

func (r *Repository) GetCategory(ctx context.Context, id int) (Category, error) {
	var category Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return Category{}, err
	}
	return category, nil
}

func (r *Repository) SaveCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *Repository) DeleteCategory(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListProducts(ctx context.Context, opts ListOptions, filter ProductFilter) (Page[Product], error) {
	query := r.db.WithContext(ctx).Model(&Product{})

	if filter.Name != nil && strings.TrimSpace(*filter.Name) != "" {
		query = query.Where("LOWER(name) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(strings.TrimSpace(*filter.Name))))
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}
	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", *filter.SupplierID)
	}
	if filter.Discontinued != nil {
		query = query.Where("discontinued = ?", *filter.Discontinued)
	}

	return list[Product](query, opts, map[string]string{
		"name":           "name",
		"unit_price":     "unit_price",
		"created_at":     "created_at",
		"units_in_stock": "units_in_stock",
	})
}

func (r *Repository) GetProduct(ctx context.Context, id int) (Product, error) {
	var product Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return Product{}, err
	}
	return product, nil
}

func (r *Repository) SaveProduct(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *Repository) DeleteProduct(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) ListSuppliers(ctx context.Context, opts ListOptions, filter SupplierFilter) (Page[Supplier], error) {
	query := r.db.WithContext(ctx).Model(&Supplier{})

	if filter.CompanyName != nil && strings.TrimSpace(*filter.CompanyName) != "" {
		query = query.Where("LOWER(company_name) LIKE ?", fmt.Sprintf("%%%s%%", strings.ToLower(strings.TrimSpace(*filter.CompanyName))))
	}
	if filter.Country != nil && strings.TrimSpace(*filter.Country) != "" {
		query = query.Where("LOWER(country) = ?", strings.ToLower(strings.TrimSpace(*filter.Country)))
	}

	return list[Supplier](query, opts, map[string]string{
		"company_name": "company_name",
		"country":      "country",
		"created_at":   "created_at",
	})
}

func (r *Repository) GetSupplier(ctx context.Context, id int) (Supplier, error) {
	var supplier Supplier
	if err := r.db.WithContext(ctx).First(&supplier, id).Error; err != nil {
		return Supplier{}, err
	}
	return supplier, nil
}

func (r *Repository) SaveSupplier(ctx context.Context, supplier *Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *Repository) DeleteSupplier(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Supplier{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func list[T any](query *gorm.DB, opts ListOptions, sortFields map[string]string) (Page[T], error) {
	normalizedOpts := normalizeListOptions(opts)

	column, direction, err := parseSort(normalizedOpts.Sort, sortFields)
	if err != nil {
		return Page[T]{}, err
	}
	if column != "" {
		query = query.Order(fmt.Sprintf("%s %s", column, direction))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return Page[T]{}, err
	}

	var items []T
	offset := (normalizedOpts.Page - 1) * normalizedOpts.PageSize
	if err := query.Offset(offset).Limit(normalizedOpts.PageSize).Find(&items).Error; err != nil {
		return Page[T]{}, err
	}

	meta := Pagination{
		Page:       normalizedOpts.Page,
		PageSize:   normalizedOpts.PageSize,
		TotalItems: int(total),
		TotalPages: calculateTotalPages(total, int64(normalizedOpts.PageSize)),
	}

	return Page[T]{Items: items, Meta: meta}, nil
}

func normalizeListOptions(opts ListOptions) ListOptions {
	result := opts
	if result.Page < 1 {
		result.Page = 1
	}
	if result.PageSize <= 0 {
		result.PageSize = 20
	}
	if result.PageSize > 100 {
		result.PageSize = 100
	}
	return result
}

func parseSort(sort string, fields map[string]string) (column string, direction string, err error) {
	sort = strings.TrimSpace(sort)
	if sort == "" {
		return "", "", nil
	}
	direction = "ASC"
	if strings.HasPrefix(sort, "-") {
		direction = "DESC"
		sort = strings.TrimPrefix(sort, "-")
	}

	if columnName, ok := fields[sort]; ok {
		return columnName, direction, nil
	}

	return "", "", fmt.Errorf("%w: %s", ErrInvalidSortField, sort)
}

var ErrInvalidSortField = errors.New("invalid sort field")

func calculateTotalPages(totalItems int64, pageSize int64) int {
	if pageSize == 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(pageSize)))
}
