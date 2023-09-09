package postgres

import (
	"context"

	model "github.com/Pizhlo/balancer/model/balancer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type TargetStore struct {
	*pgxpool.Pool
}

func New(conn *pgxpool.Pool) *TargetStore {
	db := &TargetStore{conn}
	return db
}

func (db *TargetStore) Close() {
	db.Pool.Close()
}

func (db *TargetStore) GetConfig(ctx context.Context) ([]model.ConfigDB, error) {
	q := `SELECT * FROM config`
	configs := []model.ConfigDB{}

	rows, err := db.Query(ctx, q)
	if err != nil {
		return configs, errors.Wrap(err, "error while executing query")
	}

	for rows.Next() {
		row := model.ConfigDB{}
		err := rows.Scan(&row.ID, &row.Address, &row.RRWeight, &row.IsActive)
		if err != nil {
			return configs, errors.Wrap(err, "error while scanning row")
		}
		configs = append(configs, row)
	}

	return configs, nil
}
