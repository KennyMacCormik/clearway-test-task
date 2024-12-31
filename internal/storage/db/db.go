package db

import (
	myerrors "clearway-test-task/internal/errors"
	"clearway-test-task/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log/slog"
	"time"
)

const queryGetUserPwdHash = `
    SELECT pwd FROM "users"
    WHERE login = $1;
`
const queryGetDataByAssetName = `
    SELECT data, content_type FROM "files" 
    WHERE asset_name = $1 AND user_login = $2 AND deleted_at is NULL;
`
const querySetDataByAssetName = `
    INSERT INTO "files" (asset_name, user_login, content_type, data, created_at)
    VALUES ($1, $2, $3, $4, EXTRACT(EPOCH FROM NOW()))
    ON CONFLICT (asset_name, user_login)
    WHERE deleted_at IS NULL
    DO UPDATE SET 
        content_type = EXCLUDED.content_type,
        data = EXCLUDED.data,
        updated_at = EXCLUDED.created_at;
`
const queryDeleteDataByAssetName = `
    UPDATE "files"
    SET deleted_at = EXTRACT(EPOCH FROM NOW())
    WHERE asset_name = $1 AND user_login = $2 AND deleted_at is NULL;
`

const queryGetActiveSession = `
    SELECT user_login, token, exp FROM "sessions"
    WHERE deleted_at is NULL;
`

const querySetSessionUpdate = `
    UPDATE "sessions"
	SET deleted_at = EXTRACT(EPOCH FROM NOW())
	WHERE user_login = $1 AND deleted_at IS NULL;
`

const querySetSessionInsert = `
    INSERT INTO "sessions" (user_login, token, iat, exp, created_at)
	VALUES ($1, $2, $3, $4, EXTRACT(EPOCH FROM NOW()));
`

const queryDeleteSessionByLogin = `
    UPDATE "sessions"
    SET deleted_at = EXTRACT(EPOCH FROM NOW())
    WHERE user_login = $1 AND deleted_at is NULL;
`

type Db struct {
	sql               *sql.DB
	stmtGetPwd        *sql.Stmt
	stmtGetData       *sql.Stmt
	stmtSetData       *sql.Stmt
	stmtDeleteData    *sql.Stmt
	stmtDeleteSession *sql.Stmt
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

	stmtDeleteData, err := db.Prepare(queryDeleteDataByAssetName)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stmtSetData: %w", err)
	}

	stmtDeleteSession, err := db.Prepare(queryDeleteSessionByLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stmtSetData: %w", err)
	}

	return &Db{
		sql:               db,
		stmtGetPwd:        stmtGetPwd,
		stmtGetData:       stmtGetData,
		stmtSetData:       stmtSetData,
		stmtDeleteData:    stmtDeleteData,
		stmtDeleteSession: stmtDeleteSession,
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
	if d.stmtDeleteData != nil {
		_ = d.stmtDeleteData.Close()
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
			return "", myerrors.NewErrUserNotFound(login)
		}
		return "", err
	}
	return hash, nil
}

func (d *Db) GetDataByAssetName(ctx context.Context, assetName, login string) ([]byte, string, error) {
	var data []byte
	var ct string
	if err := d.stmtGetData.QueryRowContext(ctx, assetName, login).Scan(&data, &ct); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", myerrors.NewErrAssetNotFound(login, assetName)
		}
		return nil, "", err
	}
	return data, ct, nil
}

func (d *Db) SetDataByAssetName(ctx context.Context, assetName, login, contentType string, data []byte) error {
	if _, err := d.stmtSetData.ExecContext(ctx, assetName, login, contentType, data); err != nil {
		return fmt.Errorf("failed to upsert data by asset name: %w", err)
	}
	return nil
}

func (d *Db) DeleteDataByAssetName(ctx context.Context, assetName, login string) error {
	if _, err := d.stmtDeleteData.ExecContext(ctx, assetName, login); err != nil {
		return fmt.Errorf("failed to delete data by asset name: %w", err)
	}
	return nil
}

func (d *Db) UpdateSession(ctx context.Context, login, token string, iat, exp int64) error {
	tx, err := d.sql.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmtUpdate, err := tx.PrepareContext(ctx, querySetSessionUpdate)
	if err != nil {
		return fmt.Errorf("failed to prepare tx update statement: %w", err)
	}
	defer func() { _ = stmtUpdate.Close() }()

	stmtInsert, err := tx.PrepareContext(ctx, querySetSessionInsert)
	if err != nil {
		return fmt.Errorf("failed to prepare tx insert statement: %w", err)
	}
	defer func() { _ = stmtInsert.Close() }()

	if _, err = stmtUpdate.ExecContext(ctx, login); err != nil {
		return fmt.Errorf("failed to execute tx update statement: %w", err)
	}
	if _, err = stmtInsert.ExecContext(ctx, login, token, iat, exp); err != nil {
		return fmt.Errorf("failed to execute tx insert statement: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (d *Db) DeleteSessionByLogin(ctx context.Context, login string) error {
	if _, err := d.stmtDeleteSession.ExecContext(ctx, login); err != nil {
		return fmt.Errorf("failed to delete session by login: %w", err)
	}
	return nil
}

func (d *Db) GetActiveSessions(ctx context.Context) (map[string]storage.Token, error) {
	cache := make(map[string]storage.Token)

	stmt, err := d.sql.Prepare(queryGetActiveSession)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare queryGetActiveSession: %w", err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query queryGetActiveSession: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var userLogin, token string
		var exp int64
		if err = rows.Scan(&userLogin, &token, &exp); err != nil {
			return nil, fmt.Errorf("failed to scan row of queryGetActiveSession: %w", err)
		}
		cache[userLogin] = storage.Token{Token: token, ExpireAt: exp}
	}

	return cache, nil
}
