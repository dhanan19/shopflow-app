package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"shopflow/product-service/internal/cache"
	"shopflow/product-service/internal/models"
)

type ProductStore interface {
	Create(context.Context, *models.Product) error
	GetAll(context.Context) ([]models.Product, error)
	GetByID(context.Context, int64) (*models.Product, error)
	Update(context.Context, *models.Product) error
	Delete(context.Context, int64) error
}

type ProductHandler struct {
	repo  ProductStore
	cache *cache.ProductCache
}

func NewProductHandler(repo ProductStore, productCache *cache.ProductCache) *ProductHandler {
	return &ProductHandler{repo: repo, cache: productCache}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/products/"), 10, 64)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(p.Name) == "" || p.Price < 0 || p.Stock < 0 {
		writeJSON(w, 400, map[string]string{"error": "name is required and price/stock must be non-negative"})
		return
	}
	if err := h.repo.Create(r.Context(), &p); err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to create product"})
		return
	}
	writeJSON(w, http.StatusCreated, p)
}
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to get products"})
		return
	}
	writeJSON(w, 200, products)
}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid product id"})
		return
	}
	var p *models.Product
	if h.cache != nil {
		if cached, cacheErr := h.cache.Get(r.Context(), id); cacheErr == nil {
			writeJSON(w, 200, cached)
			return
		} else if cacheErr != redis.Nil {
			// Cache failures are non-fatal; fall back to PostgreSQL.
		}
	}
	p, err = h.repo.GetByID(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 404, map[string]string{"error": "product not found"})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to get product"})
		return
	}
	if h.cache != nil {
		_ = h.cache.Set(r.Context(), p)
	}
	writeJSON(w, 200, p)
}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid product id"})
		return
	}
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(p.Name) == "" || p.Price < 0 || p.Stock < 0 {
		writeJSON(w, 400, map[string]string{"error": "name is required and price/stock must be non-negative"})
		return
	}
	p.ID = id
	if err := h.repo.Update(r.Context(), &p); errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 404, map[string]string{"error": "product not found"})
		return
	} else if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to update product"})
		return
	}
	if h.cache != nil {
		_ = h.cache.Set(r.Context(), &p)
	}
	writeJSON(w, 200, p)
}
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid product id"})
		return
	}
	if err := h.repo.Delete(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 404, map[string]string{"error": "product not found"})
		return
	} else if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to delete product"})
		return
	}
	if h.cache != nil {
		_ = h.cache.Delete(r.Context(), id)
	}
	writeJSON(w, 200, map[string]string{"message": "product deleted"})
}
