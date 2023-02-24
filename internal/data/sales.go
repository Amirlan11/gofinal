package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/laldil/greenlight/internal/validator"
	"github.com/lib/pq"
	"time"
)

type Sale struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    Runtime   `json:"duration,omitempty"`
	Foodsale    []string  `json:"foodsale,omitempty"`
	Version     int32     `json:"version"`
}

type SaleModel struct {
	DB *sql.DB
}

func (m SaleModel) Insert(sale *Sale) error {
	query :=
		`INSERT INTO sales (title, description, duration, foodsale)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, version`
	args := []any{sale.Title, sale.Description, sale.Duration, pq.Array(sale.Foodsale)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&sale.ID, &sale.CreatedAt, &sale.Version)
}

func (m SaleModel) Get(id int64) (*Sale, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT * FROM sales WHERE id = $1`

	var sale Sale
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&sale.ID,
		&sale.CreatedAt,
		&sale.Title,
		&sale.Description,
		&sale.Duration,
		pq.Array(&sale.Foodsale),
		&sale.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &sale, nil
}

func (m SaleModel) GetAll(title string, foodsale []string, filters Filters) ([]*Sale, Metadata, error) {
	query := fmt.Sprintf(` 
		SELECT COUNT(*) OVER(), id, created_at, title, description, duration, foodsale, version
		FROM sales 
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (foodsale @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, pq.Array(foodsale), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	sales := []*Sale{}
	for rows.Next() {
		var sale Sale
		err := rows.Scan(
			&totalRecords,
			&sale.ID,
			&sale.CreatedAt,
			&sale.Title,
			&sale.Description,
			&sale.Duration,
			pq.Array(&sale.Foodsale),
			&sale.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		sales = append(sales, &sale)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return sales, metadata, nil
}

func (m SaleModel) Update(sale *Sale) error {
	query :=
		`UPDATE sales
		 SET title = $1, description = $2, duration = $3, foodsale = $4, version = version + 1
		 WHERE id = $5 AND version = $6
		 RETURNING version`

	args := []any{
		sale.Title,
		sale.Description,
		sale.Duration,
		pq.Array(sale.Foodsale),
		sale.ID,
		sale.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&sale.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m SaleModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM sales WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateSale(v *validator.Validator, sale *Sale) {
	v.Check(sale.Title != "", "title", "must be provided")
	v.Check(len(sale.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(len(sale.Description) <= 500, "description", "must not be more than 500 bytes long")
	v.Check(sale.Duration != 0, "duration", "must be provided")
	v.Check(sale.Duration > 0, "duration", "must be a positive integer")
	v.Check(sale.Foodsale != nil, "foodsale", "must be provided")
	v.Check(len(sale.Foodsale) >= 1, "foodsale", "must contain at least 1 food")
	v.Check(validator.Unique(sale.Foodsale), "foodsale", "must not contain duplicate values")
}
