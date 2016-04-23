package config

import "golang.org/x/crypto/bcrypt"

type User struct {
	Email string `toml:"email"`
	Hash  string `toml:"hash"`
}

func (u User) IsPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password)) == nil
}
