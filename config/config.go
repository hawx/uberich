package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

func Read(path string) (*Config, error) {
	conf := &Config{path: path}
	_, err := toml.DecodeFile(path, conf)
	return conf, err
}

type Config struct {
	path  string
	Apps  []*App  `toml:"apps"`
	Users []*User `toml:"users"`
}

func (c *Config) GetApp(name string) *App {
	for _, app := range c.Apps {
		if app.Name == name {
			return app
		}
	}
	return nil
}

func (c *Config) SetApp(app *App) {
	if existing := c.GetApp(app.Name); existing != nil {
		existing.URI = app.URI
		existing.Secret = app.Secret
	} else {
		c.Apps = append(c.Apps, app)
	}
}

func (c *Config) RemoveApp(name string) {
	idx := -1
	for i, app := range c.Apps {
		if app.Name == name {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	c.Apps = append(c.Apps[:idx], c.Apps[idx+1:]...)
}

func (c *Config) GetUser(email string) *User {
	for _, user := range c.Users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

func (c *Config) SetUser(user *User) {
	if existing := c.GetUser(user.Email); existing != nil {
		existing.Hash = user.Hash
	} else {
		c.Users = append(c.Users, user)
	}
}

func (c *Config) RemoveUser(email string) {
	idx := -1
	for i, user := range c.Users {
		if user.Email == email {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	c.Users = append(c.Users[:idx], c.Users[idx+1:]...)
}

func (c *Config) Save() error {
	file, err := os.OpenFile(c.path, os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(c)
}
