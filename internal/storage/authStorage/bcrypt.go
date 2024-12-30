package authStorage

import (
	"clearway-test-task/pkg"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordValidator struct {
}

func (v BcryptPasswordValidator) Validate(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword(pkg.ConvertStrToBytes(hashedPassword), pkg.ConvertStrToBytes(plainPassword))
}
