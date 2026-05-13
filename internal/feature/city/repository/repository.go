package repository

import (
	"context"
	"database/sql"
)

type City struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CityRepository interface {
	GetAll(ctx context.Context) ([]City, error)
}

type cityRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) CityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) GetAll(ctx context.Context) ([]City, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM cities ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}
