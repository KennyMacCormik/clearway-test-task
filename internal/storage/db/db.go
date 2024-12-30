package db

import (
	"clearway-test-task/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Db struct {
	sql *sql.DB
}

func NewDb(dsn string) (*Db, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &Db{sql: db}, nil
}

func (d *Db) Close() error {
	return d.sql.Close()
}

func (d *Db) GetUserPwdHashByLogin(login string) (string, error) {
	var hash string
	if err := d.sql.QueryRow("SELECT pwd FROM \"user\" WHERE login = $1;", login).Scan(&hash); err != nil {
		return "", helpers.NewNotFound(fmt.Sprintf("user %s not found", login))
	}
	return hash, nil
}

func (d *Db) GetDataByAssetName(assetName, login string) ([]byte, string, error) {
	var data []byte
	var ct string
	if err := d.sql.QueryRow("SELECT data, content_type FROM \"files\" WHERE asset_id = $1 and user_login = $2;", assetName, login).Scan(&data, &ct); err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, "", helpers.NewNotFound(fmt.Sprintf("combination of user %s and asset-id %s not found", login, assetName))
		}
		return nil, "", err
	}
	return data, ct, nil
}

func (d *Db) SetDataByAssetName(ctx context.Context, assetName, login, contentType string, data []byte) error {
	q := `
        INSERT INTO "files" (asset_id, user_login, content_type, data, createdAt, updatedAt)
        VALUES ($1, $2, $3, $4, EXTRACT(EPOCH FROM now()), EXTRACT(EPOCH FROM now()))
        ON CONFLICT (asset_id, user_login)
        DO UPDATE SET 
            content_type = EXCLUDED.content_type,
            data = EXCLUDED.data,
            updatedAt = EXCLUDED.updatedAt;
    `
	_, err := d.sql.ExecContext(ctx, q, assetName, login, contentType, data)
	if err != nil {
		return fmt.Errorf("failed to insert or update file: %w", err)
	}
	return nil
}
