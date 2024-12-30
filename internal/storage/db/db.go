package db

import (
	"clearway-test-task/pkg"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"time"
)

const queryGetUserPwdHash = `
    SELECT pwd FROM "users" WHERE login = $1;
`
const queryGetDataByAssetName = `
    SELECT data, content_type FROM "files" 
    WHERE asset_id = $1 AND user_login = $2;
`
const querySetDataByAssetName = `
    INSERT INTO "files" (asset_id, user_login, content_type, data, created_at)
    VALUES ($1, $2, $3, $4, EXTRACT(EPOCH FROM NOW()))
    ON CONFLICT (asset_id, user_login)
    DO UPDATE SET 
        content_type = EXCLUDED.content_type,
        data = EXCLUDED.data,
        updated_at = EXCLUDED.created_at;
`

type Db struct {
	sql         *sql.DB
	stmtGetPwd  *sql.Stmt
	stmtGetData *sql.Stmt
	stmtSetData *sql.Stmt
}

func NewDb(dsn string, ConnMax, ConnMaxIdle int, ConnMaxReuse time.Duration) (*Db, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	db.SetMaxOpenConns(ConnMax)         // Maximum number of open connections
	db.SetMaxIdleConns(ConnMaxIdle)     // Maximum number of idle connections
	db.SetConnMaxLifetime(ConnMaxReuse) // Maximum connection lifetime

	stmtGetPwd, err := db.Prepare(queryGetUserPwdHash)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stmtGetPwd: %w", err)
	}

	stmtGetData, err := db.Prepare(queryGetDataByAssetName)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stmtGetData: %w", err)
	}

	stmtSetData, err := db.Prepare(querySetDataByAssetName)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stmtSetData: %w", err)
	}

	return &Db{
		sql:         db,
		stmtGetPwd:  stmtGetPwd,
		stmtGetData: stmtGetData,
		stmtSetData: stmtSetData,
	}, nil
}

func (d *Db) Close(lg *slog.Logger) error {
	if d.stmtGetPwd != nil {
		_ = d.stmtGetPwd.Close()
	}
	if d.stmtGetData != nil {
		_ = d.stmtGetData.Close()
	}
	if d.stmtSetData != nil {
		_ = d.stmtSetData.Close()
	}
	err := d.sql.Close()
	if err != nil {
		lg.Error("failed to close the database", "error", err)
		return err
	}
	lg.Debug("closed db")
	return nil
}

func (d *Db) GetUserPwdHashByLogin(ctx context.Context, login string) (string, error) {
	var hash string
	if err := d.stmtGetPwd.QueryRowContext(ctx, login).Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: user %s not found", pkg.ErrNotFound, login)
		}
		return "", err
	}
	return hash, nil
}

func (d *Db) GetDataByAssetName(ctx context.Context, assetName, login string) ([]byte, string, error) {
	var data []byte
	var ct string
	if err := d.stmtGetData.QueryRowContext(ctx, login).Scan(&data, &ct); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("%w: combination of user %s and asset-id %s not found", pkg.ErrNotFound, login, assetName)
		}
		return nil, "", err
	}
	return data, ct, nil
}

func (d *Db) SetDataByAssetName(ctx context.Context, assetName, login, contentType string, data []byte) error {
	_, err := d.stmtSetData.ExecContext(ctx, assetName, login, contentType, data)
	if err != nil {
		return fmt.Errorf("failed upsert: %w", err)
	}
	return nil
}
