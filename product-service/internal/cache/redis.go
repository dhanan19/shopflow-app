package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"shopflow/product-service/internal/models"
)

type ProductCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewProductCache(addr string, ttl time.Duration) *ProductCache {
	return &ProductCache{client: redis.NewClient(&redis.Options{Addr: addr}), ttl: ttl}
}

func (c *ProductCache) Close() error { return c.client.Close() }

func (c *ProductCache) Get(ctx context.Context, id int64) (*models.Product, error) {
	value, err := c.client.Get(ctx, key(id)).Result()
	if err != nil {
		return nil, err
	}
	var product models.Product
	if err := json.Unmarshal([]byte(value), &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (c *ProductCache) Set(ctx context.Context, product *models.Product) error {
	value, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(product.ID), value, c.ttl).Err()
}

func (c *ProductCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, key(id)).Err()
}

func key(id int64) string { return "shopflow:product:" + strconv.FormatInt(id, 10) }
