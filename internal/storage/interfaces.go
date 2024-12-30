package storage

import "context"

type Auth interface {
	GetToken(ctx context.Context, login, password string) (string, int64, error)
	ValidateToken(token string) (string, error)
}

type PasswordValidator interface {
	Validate(hashedPassword, plainPassword string) error
}

type Db interface {
	GetUserPwdHashByLogin(ctx context.Context, login string) (string, error)
	GetDataByAssetName(ctx context.Context, id, login string) ([]byte, string, error)
	SetDataByAssetName(ctx context.Context, assetName, login, contentType string, data []byte) error
}
