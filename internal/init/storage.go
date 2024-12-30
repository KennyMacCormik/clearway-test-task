package myinit

import (
	"clearway-test-task/internal/config"
	"clearway-test-task/internal/storage/authStorage"
	"clearway-test-task/internal/storage/db"
)

func Storage(cfg config.Config) (*db.Db, *authStorage.AuthStorage, error) {
	database, err := db.NewDb(cfg.Db.Dsn, cfg.Db.ConnMax, cfg.Db.ConnMaxIdle, cfg.Db.ConnMaxReuse)
	if err != nil {
		return &db.Db{}, &authStorage.AuthStorage{}, err
	}

	return database,
		authStorage.NewAuthStorage(database,
			authStorage.BcryptPasswordValidator{},
			cfg.Auth.CacheCleanupInterval,
			cfg.Auth.TokenTTL,
			cfg.Auth.HmacSecret,
		),
		nil
}
