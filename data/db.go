package data

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type User struct {
	Email    string
	Hash     string
	Token    string
	Expires  time.Time
	Verified bool
}

type Application struct {
	Name    string
	RootURI string
	Secret  string
}

func (a Application) CanRedirectTo(uri string) bool {
	return strings.HasPrefix(uri, a.RootURI)
}

func (a Application) HashWithSecret(data []byte) []byte {
	mac := hmac.New(sha256.New, []byte(a.Secret))
	mac.Write(data)
	return mac.Sum(nil)
}

type Database interface {
	GetUser(email string) (User, error)
	SetUser(user User) error

	GetApplication(name string) (Application, error)
	ListApplications() ([]Application, error)
	SetApplication(app Application) error
	RemoveApplication(name string) error

	Close() error
}

type database struct {
	db *bolt.DB
}

var (
	usersBucket        = []byte("users")
	applicationsBucket = []byte("applications")
)

func Open(path string) (Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(usersBucket)
		if err == nil {
			_, err = tx.CreateBucketIfNotExists(applicationsBucket)
		}
		return err
	})

	return &database{db}, err
}

func (d *database) GetUser(email string) (user User, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		v := b.Get([]byte(email))

		if v == nil {
			return errors.New("no such email")
		}

		return json.Unmarshal(v, &user)
	})

	return user, err
}

func (d *database) SetUser(user User) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		v, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return b.Put([]byte(user.Email), v)
	})
}

func (d *database) GetApplication(name string) (app Application, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(applicationsBucket)
		v := b.Get([]byte(name))

		if v == nil {
			return errors.New("no such application")
		}

		return json.Unmarshal(v, &app)
	})

	return app, err
}

func (d *database) ListApplications() (apps []Application, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(applicationsBucket)

		return b.ForEach(func(k []byte, v []byte) error {
			var app Application
			if e := json.Unmarshal(v, &app); e != nil {
				return e
			}
			apps = append(apps, app)
			return nil
		})
	})

	return apps, err
}

func (d *database) SetApplication(app Application) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(applicationsBucket)
		v, err := json.Marshal(app)
		if err != nil {
			return err
		}
		return b.Put([]byte(app.Name), v)
	})
}

func (d *database) RemoveApplication(name string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(applicationsBucket)
		return b.Delete([]byte(name))
	})
}

func (d *database) Close() error {
	return d.db.Close()
}
