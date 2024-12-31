package authStorage

import (
	myerrors "clearway-test-task/internal/errors"
	"clearway-test-task/internal/storage"
	"clearway-test-task/pkg"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"sync"
	"time"
)

type AuthStorage struct {
	db                   storage.Db
	DeleteSessionTimeout time.Duration
	pv                   storage.PasswordValidator
	// sync.Map will be better at large request amount
	cache map[string]storage.Token
	// RW mutex because each request checks token
	cacheMtx    sync.RWMutex
	cacheTicker *time.Ticker
	closer      chan struct{}
	once        sync.Once
	tokenTTL    time.Duration
	hmacSecret  string
	lg          *slog.Logger
}

func NewAuthStorage(db storage.Db, DeleteSessionTimeout time.Duration, validator storage.PasswordValidator, cacheCleanupInterval time.Duration, tokenTTL time.Duration, hmacSecret string, lg *slog.Logger) *AuthStorage {
	st := &AuthStorage{
		db:                   db,
		DeleteSessionTimeout: DeleteSessionTimeout,
		pv:                   validator,
		cache:                make(map[string]storage.Token),
		cacheTicker:          time.NewTicker(cacheCleanupInterval),
		closer:               make(chan struct{}),
		tokenTTL:             tokenTTL,
		hmacSecret:           hmacSecret,
		lg:                   lg,
	}
	cache, err := loadCache(db, DeleteSessionTimeout)
	if err != nil {
		st.lg.Error("failed to load the cache", "error", err)
	} else {
		st.cache = cache
	}
	go st.cacheCleaner()
	return st
}

func loadCache(db storage.Db, DeleteSessionTimeout time.Duration) (map[string]storage.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DeleteSessionTimeout)
	defer cancel()
	return db.GetActiveSessions(ctx)
}

func (a *AuthStorage) auth(ctx context.Context, login, password string) error {
	hash, err := a.db.GetUserPwdHashByLogin(ctx, login)
	if err != nil {
		return err
	}

	if err = a.pv.Validate(hash, password); err != nil {
		return fmt.Errorf("%w: invalid credentials: %w", myerrors.ErrNotFound, err)
	}

	return nil
}

func (a *AuthStorage) GetToken(ctx context.Context, login, password string) (string, int64, error) {
	if err := a.auth(ctx, login, password); err != nil {
		return "", 0, err
	}

	iat := time.Now().Unix()
	exp := time.Unix(iat, 0).Add(a.tokenTTL).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"iat":   iat,
		"exp":   exp,
	})

	signedToken, err := t.SignedString(pkg.ConvertStrToBytes(a.hmacSecret))
	if err != nil {
		return "", 0, err
	}

	if err = a.db.UpdateSession(ctx, login, signedToken, iat, exp); err != nil {
		return "", 0, err
	}
	a.setTokenWithLock(login, signedToken, exp)

	return signedToken, exp, nil
}

func (a *AuthStorage) setTokenWithLock(login, signedToken string, exp int64) {
	a.cacheMtx.Lock()
	defer a.cacheMtx.Unlock()
	a.cache[login] = storage.Token{Token: signedToken, ExpireAt: exp}
}

func (a *AuthStorage) ValidateToken(decToken string) (string, error) {
	// validate token
	tkn, err := a.validateToken(decToken)
	if err != nil {
		return "", err
	}
	// validate claims
	claims, ok := tkn.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("token claims are not accessible")
	}
	if err = a.validateClaims(claims); err != nil {
		return "", err
	}
	// validate cache
	login := claims["login"].(string)
	if err = a.validateCache(login, decToken); err != nil {
		return "", err
	}

	return login, nil
}

func (a *AuthStorage) validateCache(login, decToken string) error {
	cachedToken, ok := a.getCachedTokenWithRLock(login)
	if !ok || cachedToken.Token != decToken {
		return errors.New("token not registered")
	}
	return nil
}

func (a *AuthStorage) validateToken(decToken string) (*jwt.Token, error) {
	tkn, err := jwt.Parse(decToken, func(tkn *jwt.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tkn.Header["alg"])
		}
		return pkg.ConvertStrToBytes(a.hmacSecret), nil
	})
	if err != nil || tkn.Valid != true {
		return nil, err
	}
	return tkn, nil
}
func (a *AuthStorage) validateClaims(claims jwt.MapClaims) error {
	validateTimestamp := func(val float64, ok bool) error {
		if !ok {
			return errors.New("iat claim does not exist")
		}
		if val < 1 {
			return errors.New("invalid timestamp")
		}
		return nil
	}
	// validate login exists and contains value
	s, ok := claims["login"].(string)
	if !ok {
		return errors.New("login claim does not exist")
	}
	if s == "" {
		return errors.New("login claim is empty")
	}
	// validate iat exists and contains value
	iat, ok := claims["iat"].(float64)
	if err := validateTimestamp(iat, ok); err != nil {
		return err
	}
	// validate exp exists and contains value
	exp, ok := claims["exp"].(float64)
	if err := validateTimestamp(exp, ok); err != nil {
		return err
	}
	// validate sanity of exp and iat values
	if iat >= exp || int64(exp)-int64(iat) == int64(a.tokenTTL) || int64(exp) < time.Now().Unix() {
		return errors.New("expired or invalid token")
	}
	return nil
}

func (a *AuthStorage) getCachedTokenWithRLock(login string) (storage.Token, bool) {
	a.cacheMtx.RLock()
	defer a.cacheMtx.RUnlock()
	cachedToken, ok := a.cache[login]
	return cachedToken, ok
}

func (a *AuthStorage) Close() error {
	a.once.Do(func() {
		close(a.closer)
		a.cacheTicker.Stop()
	})
	a.lg.Debug("auth cache closed")
	return nil
}

func (a *AuthStorage) cacheCleaner() {
	for {
		select {
		case <-a.cacheTicker.C:
			a.cacheMtx.RLock()
			for login, tkn := range a.cache {
				if time.Now().Unix() > tkn.ExpireAt {
					a.cacheMtx.RUnlock()
					err := a.deleteExpiredSessionFromDb(login)
					if err != nil {
						a.lg.Error("failed to delete expired session from db", "error", err)
						continue
					}
					a.cacheMtx.Lock()
					delete(a.cache, login)
					a.cacheMtx.Unlock()
					a.cacheMtx.RLock()
				}
			}
			a.cacheMtx.RUnlock()
		case <-a.closer:
			return
		}
	}
}

func (a *AuthStorage) deleteExpiredSessionFromDb(login string) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.DeleteSessionTimeout)
	defer cancel()
	return a.db.DeleteSessionByLogin(ctx, login)
}
