package repository

import "github.com/jackc/pgx/v5/pgxpool"

type PatternsRepository struct {
	pool *pgxpool.Pool
}

func NewPatternsRepository(pool *pgxpool.Pool) *PatternsRepository {
	return &PatternsRepository{
		pool: pool,
	}
}
