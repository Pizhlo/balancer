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

// GetAddress возвращает доступный незанятый адрес
func (db *TargetStore) GetAddress(ctx context.Context) (string, error) {
	q := `SELECT address FROM config WHERE NOT is_active LIMIT 1`

	var address string
	err := db.QueryRow(ctx, q).Scan(&address)
	if err != nil {
		return "", errors.Wrap(err, "err while getting address from db:")
	}

	return address, nil
}

// UpdateStatus обновляет в базе статус адреса
func (db *TargetStore) UpdateStatus(ctx context.Context, status bool, address string) error {
	q := `UPDATE config SET is_active = $1 WHERE address = $2`

	_, err := db.Exec(ctx, q, status, address)
	if err != nil {
		return errors.Wrap(err, "err while updating status:")
	}

	return nil
}
