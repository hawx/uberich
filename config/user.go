package config

import "golang.org/x/crypto/bcrypt"

type User struct {
	Email string `toml:"email"`
	Hash  string `toml:"hash"`
}

func (u User) IsPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Hash), []byte(password)) == nil
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), -1)
	if err == nil {
		u.Hash = string(hash)
	}
	return err
}
