package postgres

import (
	"context"

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

func (db *TargetStore) GetAddress(ctx context.Context) (string, error) {
	q := `SELECT address FROM config WHERE NOT is_active LIMIT 1`

	var address string
	err := db.QueryRow(ctx, q).Scan(&address)
	if err != nil {
		return "", errors.Wrap(err, "err while getting address from db:")
	}

	return address, nil
}
