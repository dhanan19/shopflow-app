package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"shopflow/product-service/internal/cache"
	"shopflow/product-service/internal/config"
	"shopflow/product-service/internal/handlers"
	"shopflow/product-service/internal/repository"
	"time"
)

func main() {
	ctx := context.Background()
	db, err := config.ConnectDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	repo := repository.NewProductRepository(db)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	productCache := cache.NewProductCache(redisAddr, 5*time.Minute)
	defer productCache.Close()
	h := handlers.NewProductHandler(repo, productCache)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"ok","service":"product-service"}`))
	})
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetProducts(w, r)
		case http.MethodPost:
			h.CreateProduct(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetProduct(w, r)
		case http.MethodPut:
			h.UpdateProduct(w, r)
		case http.MethodDelete:
			h.DeleteProduct(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	log.Println("product-service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
