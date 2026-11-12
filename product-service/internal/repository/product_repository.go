package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"shopflow/product-service/internal/models"
)

type ProductRepository struct{ db *pgxpool.Pool }

func NewProductRepository(db *pgxpool.Pool) *ProductRepository { return &ProductRepository{db: db} }

func (r *ProductRepository) Create(ctx context.Context, p *models.Product) error {
	return r.db.QueryRow(ctx, `INSERT INTO products (name,description,price,stock,category) VALUES ($1,$2,$3,$4,$5) RETURNING id,created_at,updated_at`, p.Name, p.Description, p.Price, p.Stock, p.Category).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}
func (r *ProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,description,price,stock,category,created_at,updated_at FROM products ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRow(ctx, `SELECT id,name,description,price,stock,category,created_at,updated_at FROM products WHERE id=$1`, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *ProductRepository) Update(ctx context.Context, p *models.Product) error {
	return r.db.QueryRow(ctx, `UPDATE products SET name=$1,description=$2,price=$3,stock=$4,category=$5,updated_at=CURRENT_TIMESTAMP WHERE id=$6 RETURNING created_at,updated_at`, p.Name, p.Description, p.Price, p.Stock, p.Category, p.ID).Scan(&p.CreatedAt, &p.UpdatedAt)
}
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
