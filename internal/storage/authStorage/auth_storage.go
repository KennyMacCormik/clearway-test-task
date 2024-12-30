package authStorage

import (
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
	db storage.Db
	pv storage.PasswordValidator
	// sync.Map will be better at large request amount
	cache map[string]token
	// RW mutex because each request checks token
	cacheMtx    sync.RWMutex
	cacheTicker *time.Ticker
	closer      chan struct{}
	once        sync.Once
	tokenTTL    time.Duration
	hmacSecret  string
}

type token struct {
	Token    string
	ExpireAt int64
}

func NewAuthStorage(db storage.Db, validator storage.PasswordValidator, cacheCleanupInterval time.Duration, tokenTTL time.Duration, hmacSecret string) *AuthStorage {
	st := &AuthStorage{
		db:          db,
		pv:          validator,
		cache:       make(map[string]token),
		cacheTicker: time.NewTicker(cacheCleanupInterval),
		closer:      make(chan struct{}),
		tokenTTL:    tokenTTL,
		hmacSecret:  hmacSecret,
	}
	go st.cacheCleaner()
	return st
}

func (a *AuthStorage) auth(ctx context.Context, login, password string) error {
	hash, err := a.db.GetUserPwdHashByLogin(ctx, login)
	if err != nil {
		return err
	}

	if err = a.pv.Validate(hash, password); err != nil {
		return fmt.Errorf("%w: invalid credentials: %w", pkg.ErrNotFound, err)
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
	base64Token := pkg.Base64Encode(signedToken)

	a.cacheMtx.Lock()
	a.cache[login] = token{Token: base64Token, ExpireAt: exp}
	a.cacheMtx.Unlock()

	return base64Token, exp, nil
}

func (a *AuthStorage) ValidateToken(base64EncToken string) (string, error) {
	// decode token
	decToken, err := decodeToken(base64EncToken)
	if err != nil {
		return "", err
	}
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

func decodeToken(base64EncToken string) (string, error) {
	decToken := pkg.Base64Decode(base64EncToken)
	if decToken == "" || decToken == " " {
		return "", errors.New("token is empty")
	}
	return decToken, nil
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

func (a *AuthStorage) getCachedTokenWithRLock(login string) (token, bool) {
	a.cacheMtx.RLock()
	defer a.cacheMtx.RUnlock()
	cachedToken, ok := a.cache[login]
	return cachedToken, ok
}

func (a *AuthStorage) Close(lg *slog.Logger) error {
	a.once.Do(func() {
		close(a.closer)
		a.cacheTicker.Stop()
	})
	lg.Debug("auth cache closed")
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
