package catalog

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestValidateProductInput_CategoryMustBePositive(t *testing.T) {
	err := validateProductInput(ProductInput{Name: "Name", CategoryID: 0, SupplierID: 1, UnitPrice: 1})
	if err == nil || err.Code != ErrorValidation {
		t.Fatalf("expected validation error for category, got %v", err)
	}
}

func TestValidateProductInput_SupplierMustBePositive(t *testing.T) {
	err := validateProductInput(ProductInput{Name: "Name", CategoryID: 1, SupplierID: 0, UnitPrice: 1})
	if err == nil || err.Code != ErrorValidation {
		t.Fatalf("expected validation error for supplier, got %v", err)
	}
}

func TestValidateProductInput_UnitPriceNonNegative(t *testing.T) {
	err := validateProductInput(ProductInput{Name: "Name", CategoryID: 1, SupplierID: 1, UnitPrice: -1})
	if err == nil || err.Code != ErrorValidation {
		t.Fatalf("expected validation error for unit price, got %v", err)
	}
}

func TestValidateProductInput_InventoryNonNegative(t *testing.T) {
	err := validateProductInput(ProductInput{Name: "Name", CategoryID: 1, SupplierID: 1, UnitPrice: 1, UnitsInStock: -1})
	if err == nil || err.Code != ErrorValidation {
		t.Fatalf("expected validation error for inventory counts, got %v", err)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	if !isUniqueViolation(pgErr) {
		t.Fatalf("expected unique violation to be detected")
	}
	if isUniqueViolation(errors.New("other")) {
		t.Fatalf("expected non-pg error to not be treated as unique violation")
	}
}
