package authS

import (
	"clearway-test-task/internal/helpers"
	"clearway-test-task/internal/storage"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sync"
	"time"
)

const tokenTTL = 24 * time.Hour
const cacheCleanupInterval = 1 * time.Hour

// it is unsafe to keep secret hardcoded
const hmacSecret = "secret"

type AuthStorage struct {
	db storage.Db
	// sync.Map will be better at large request amount
	cache map[string]token
	// RW mutex because each request checks token
	cacheMtx    sync.RWMutex
	cacheTicker *time.Ticker
	closer      chan struct{}
	once        sync.Once
}

type token struct {
	Token    string
	ExpireAt int64
}

func NewAuthStorage(db storage.Db) *AuthStorage {
	st := &AuthStorage{
		db:          db,
		cache:       make(map[string]token),
		cacheTicker: time.NewTicker(cacheCleanupInterval),
		closer:      make(chan struct{}),
	}
	go st.cacheCleaner()
	return st
}

func (a *AuthStorage) auth(login, password string) error {
	hash, err := a.db.GetUserPwdHashByLogin(login)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(helpers.ConvertStrToBytes(hash), helpers.ConvertStrToBytes(password))
}

func (a *AuthStorage) GetToken(login, password string) (string, int64, error) {
	if err := a.auth(login, password); err != nil {
		return "", 0, err
	}

	iat := time.Now().Unix()
	exp := time.Unix(iat, 0).Add(tokenTTL).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"iat":   iat,
		"exp":   exp,
	})

	signedToken, err := t.SignedString(helpers.ConvertStrToBytes(hmacSecret))
	if err != nil {
		return "", 0, err
	}
	base64Token := helpers.Base64Encode(signedToken)

	a.cacheMtx.Lock()
	a.cache[login] = token{Token: base64Token, ExpireAt: exp}
	a.cacheMtx.Unlock()

	return base64Token, exp, nil
}

func (a *AuthStorage) ValidateToken(token string) bool {
	// validate token
	t := validateJwt(token)
	if t == nil {
		return false
	}
	// validate necessary claims exist
	claims := t.Claims.(jwt.MapClaims)
	if !validateClaims(claims) {
		return false
	}
	// validate ttl
	ttl := claims["exp"].(float64)
	if int64(ttl) < time.Now().Unix() {
		return false
	}
	// validate this token is in effect
	cachedToken, ok := a.getCachedTokenWithRLock(claims["login"].(string))
	if !ok || cachedToken.Token != token {
		return false
	}

	return true
}

// validateJwt checks if token contains valid JWT and claims are accessible. Returns nil in case of an error.
func validateJwt(token string) *jwt.Token {
	// validate token can be decoded
	decodedToken := helpers.Base64Decode(token)
	if token == "" {
		return nil
	}
	// validate signature
	t, err := jwt.Parse(decodedToken, func(tkn *jwt.Token) (interface{}, error) {
		if _, ok := tkn.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tkn.Header["alg"])
		}
		return helpers.ConvertStrToBytes(hmacSecret), nil
	})
	if err != nil {
		return nil
	}
	// validate claims are accessible
	_, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	return t
}

func validateClaims(claims jwt.MapClaims) bool {
	// validate login exists and contains value
	s, ok := claims["login"].(string)
	if !ok {
		return false
	}
	if s == "" {
		return false
	}
	// validate iat exists and contains value
	iat, ok := claims["iat"].(float64)
	if !ok {
		return false
	}
	if iat < 1 {
		return false
	}
	// validate exp exists and contains value
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}
	if exp < 1 {
		return false
	}
	// validate sanity of exp and iat values
	if iat >= exp || int64(exp)-int64(iat) == int64(tokenTTL) {
		return false
	}
	return true
}

func (a *AuthStorage) getCachedTokenWithRLock(login string) (token, bool) {
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
