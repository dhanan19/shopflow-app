package main

import (
    "encoding/json"
    "log"
    "net/http"
)

func main() {

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

        w.Header().Set("Content-Type", "application/json")

        json.NewEncoder(w).Encode(map[string]string{
            "service": "product-service",
            "status":  "healthy",
        })
    })

    log.Println("Product Service running on port 8080")

    log.Fatal(http.ListenAndServe(":8080", nil))
}
