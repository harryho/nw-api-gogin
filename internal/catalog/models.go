package catalog

import "time"

type Category struct {
	ID          int       `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Category) TableName() string { return "categories" }

type Supplier struct {
	ID           int       `gorm:"column:id;primaryKey"`
	CompanyName  string    `gorm:"column:company_name"`
	ContactName  *string   `gorm:"column:contact_name"`
	ContactTitle *string   `gorm:"column:contact_title"`
	Address      *string   `gorm:"column:address"`
	City         *string   `gorm:"column:city"`
	Region       *string   `gorm:"column:region"`
	PostalCode   *string   `gorm:"column:postal_code"`
	Country      *string   `gorm:"column:country"`
	Phone        *string   `gorm:"column:phone"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (Supplier) TableName() string { return "suppliers" }

type Product struct {
	ID              int       `gorm:"column:id;primaryKey"`
	CategoryID      int       `gorm:"column:category_id"`
	SupplierID      int       `gorm:"column:supplier_id"`
	Name            string    `gorm:"column:name"`
	QuantityPerUnit *string   `gorm:"column:quantity_per_unit"`
	UnitPrice       float64   `gorm:"column:unit_price"`
	UnitsInStock    int       `gorm:"column:units_in_stock"`
	UnitsOnOrder    int       `gorm:"column:units_on_order"`
	ReorderLevel    int       `gorm:"column:reorder_level"`
	Discontinued    bool      `gorm:"column:discontinued"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (Product) TableName() string { return "products" }

type Pagination struct {
	Page       int
	PageSize   int
	TotalPages int
	TotalItems int
}

type ListOptions struct {
	Page     int
	PageSize int
	Sort     string
}

type CategoryFilter struct {
	Name *string
}

type ProductFilter struct {
	Name         *string
	CategoryID   *int
	SupplierID   *int
	Discontinued *bool
}

type SupplierFilter struct {
	CompanyName *string
	Country     *string
}

type Page[T any] struct {
	Items []T
	Meta  Pagination
}

type CategoryInput struct {
	Name        string
	Description *string
}

type SupplierInput struct {
	CompanyName  string
	ContactName  *string
	ContactTitle *string
	Address      *string
	City         *string
	Region       *string
	PostalCode   *string
	Country      *string
	Phone        *string
}

type ProductInput struct {
	CategoryID      int
	SupplierID      int
	Name            string
	QuantityPerUnit *string
	UnitPrice       float64
	UnitsInStock    int
	UnitsOnOrder    int
	ReorderLevel    int
	Discontinued    bool
}
