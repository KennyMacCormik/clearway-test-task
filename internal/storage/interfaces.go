package storage

type Auth interface {
	GetToken(login, password string) (string, int64, error)
	ValidateToken(token string) bool
}

type Db interface {
	GetUserPwdHashByLogin(login string) (string, error)
	GetDataByAssetName(id, login string) ([]byte, string, error)
	SetDataByAssetName(id, login, contextType string, data []byte) error
}
