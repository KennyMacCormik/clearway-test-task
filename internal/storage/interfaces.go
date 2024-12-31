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
	DeleteDataByAssetName(ctx context.Context, assetName, login string) error
	UpdateSession(ctx context.Context, login, token string, iat, exp int64) error
	DeleteSessionByLogin(ctx context.Context, login string) error
	GetActiveSessions(ctx context.Context) (map[string]Token, error)
}

type Token struct {
	Token    string
	ExpireAt int64
}
