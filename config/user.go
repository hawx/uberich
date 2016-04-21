package config

type User struct {
	Email string `toml:"email"`
	Hash  string `toml:"hash"`
}
