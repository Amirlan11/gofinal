package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/laldil/greenlight/internal/validator"
	"github.com/lib/pq"
)

type Food struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Price     int32     `json:"price"`
	Waittime  int32     `json:"waittime"`
	Recipe    []string  `json:"recipe"`
	Version   int32     `json:"version"`
}

type FoodModel struct {
	DB *sql.DB
}

func (f FoodModel) Insert(food *Food) error {
	query :=
		`INSERT INTO foods (title, price, waittime, recipe)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, version`
	args := []any{food.Title, food.Price, food.Waittime, pq.Array(food.Recipe)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return f.DB.QueryRowContext(ctx, query, args...).Scan(&food.ID, &food.CreatedAt, &food.Version)
}

func (f FoodModel) Get(id int64) (*Food, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT * FROM foods WHERE id = $1`

	var food Food
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := f.DB.QueryRowContext(ctx, query, id).Scan(
		&food.ID,
		&food.CreatedAt,
		&food.Title,
		&food.Price,
		&food.Waittime,
		pq.Array(&food.Recipe),
		&food.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &food, nil
}

func (f FoodModel) GetAll(title string, recipe []string, filters Filters) ([]*Food, Metadata, error) {
	query := fmt.Sprintf(` 
		SELECT COUNT(*) OVER(), id, created_at, title, price, waittime, recipe, version
		FROM foods 
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (recipe @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, pq.Array(recipe), filters.limit(), filters.offset()}

	rows, err := f.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	foods := []*Food{}
	for rows.Next() {
		var food Food
		err := rows.Scan(
			&totalRecords,
			&food.ID,
			&food.CreatedAt,
			&food.Title,
			&food.Price,
			&food.Waittime,
			pq.Array(&food.Recipe),
			&food.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		foods = append(foods, &food)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return foods, metadata, nil
}

func (f FoodModel) Update(food *Food) error {
	query :=
		`UPDATE foods
		 SET title = $1, price = $2, waittime = $3, recipe = $4, version = version + 1
		 WHERE id = $5 AND version = $6
		 RETURNING version`

	args := []any{
		food.Title,
		food.Price,
		food.Waittime,
		pq.Array(food.Recipe),
		food.ID,
		food.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := f.DB.QueryRowContext(ctx, query, args...).Scan(&food.Version)
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

func (m FoodModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM foods WHERE id = $1`

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

func ValidateFood(v *validator.Validator, food *Food) {
	v.Check(food.Title != "", "title", "must be provided")
	v.Check(len(food.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(food.Price != 0, "price", "must be provided")
	v.Check(food.Waittime != 0, "waittime", "must be provided")
	v.Check(len(food.Recipe) > 0, "recipe", "must be provided")
	v.Check(food.Waittime != 0, "waittime", "must be provided")
	v.Check(food.Waittime > 0, "waittime", "must be a positive integer")
	v.Check(len(food.Recipe) >= 1, "recipe", "must contain at least 1 food")
	v.Check(validator.Unique(food.Recipe), "recipe", "must not contain duplicate values")
}
