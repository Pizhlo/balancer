package postgres

import (
	"context"

	model "github.com/Pizhlo/balancer/model/balancer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type BalancerStore struct {
	*pgxpool.Pool
}

func New(conn *pgxpool.Pool) *BalancerStore {
	db := &BalancerStore{conn}
	return db
}

func (db *BalancerStore) Close() {
	db.Pool.Close()
}

func (db *BalancerStore) GetTargets(ctx context.Context) ([]model.Target, error) {
	q := `SELECT address, rr_weight, is_active FROM config WHERE is_active`

	targets := []model.Target{}
	rows, err := db.Query(ctx, q)
	if err != nil {
		return targets, errors.Wrap(err, "err while getting targets from db:")
	}

	for rows.Next() {
		var target model.Target

		err := rows.Scan(&target.Address, &target.RRWeight, &target.IsActive)
		if err != nil {
			return targets, errors.Wrap(err, "err while scanning targets:")
		}

		targets = append(targets, target)
	}

	return targets, nil
}
