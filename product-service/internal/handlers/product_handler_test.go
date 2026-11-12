package handlers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"net/http/httptest"
	"shopflow/product-service/internal/models"
	"strings"
	"testing"
)

type mockProductRepository struct{ products []models.Product }

func (m *mockProductRepository) Create(ctx context.Context, p *models.Product) error {
	p.ID = int64(len(m.products) + 1)
	m.products = append(m.products, *p)
	return nil
}
func (m *mockProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	return m.products, nil
}
func (m *mockProductRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	for i := range m.products {
		if m.products[i].ID == id {
			return &m.products[i], nil
		}
	}
	return nil, pgx.ErrNoRows
}
func (m *mockProductRepository) Update(ctx context.Context, p *models.Product) error {
	for i := range m.products {
		if m.products[i].ID == p.ID {
			m.products[i] = *p
			return nil
		}
	}
	return pgx.ErrNoRows
}
func (m *mockProductRepository) Delete(ctx context.Context, id int64) error {
	for i := range m.products {
		if m.products[i].ID == id {
			m.products = append(m.products[:i], m.products[i+1:]...)
			return nil
		}
	}
	return pgx.ErrNoRows
}

func TestCreateProductEmptyName(t *testing.T) {
	h := NewProductHandler(&mockProductRepository{}, nil)
	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"","price":100,"stock":10}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateProduct(rr, req)
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
func TestCreateProductSuccess(t *testing.T) {
	repo := &mockProductRepository{}
	h := NewProductHandler(repo, nil)
	req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"Laptop","price":75000,"stock":10,"category":"Electronics"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateProduct(rr, req)
	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	if len(repo.products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(repo.products))
	}
}
func TestGetProductNotFound(t *testing.T) {
	h := NewProductHandler(&mockProductRepository{}, nil)
	req := httptest.NewRequest("GET", "/products/999", nil)
	rr := httptest.NewRecorder()
	h.GetProduct(rr, req)
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
