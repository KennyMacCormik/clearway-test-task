package errors

import "fmt"

type ErrAssetNotFound struct {
	Login   string
	AssetID string
}

func (e ErrAssetNotFound) Error() string {
	return fmt.Sprintf("resource not found: combination of user %s and asset-id %s not found", e.Login, e.AssetID)
}

func NewErrAssetNotFound(login, assetId string) error {
	return ErrAssetNotFound{
		Login:   login,
		AssetID: assetId,
	}
}

type ErrUserNotFound struct {
	Login string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found: %s", e.Login)
}

func NewErrUserNotFound(login string) error {
	return ErrUserNotFound{
		Login: login,
	}
}
