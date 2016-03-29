package data

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

type Record struct {
	Email    string
	Hash     string
	Token    string
	Expires  time.Time
	Verified bool
}

type Database interface {
	Get(email string) (record Record, ok bool)
	Set(record Record) (ok bool)
	Close() error
}

type database struct {
	db *bolt.DB
}

var bucketName = []byte("users")

func Open(path string) (Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})

	return &database{db}, err
}

func (d *database) Get(email string) (record Record, ok bool) {
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		r := b.Get([]byte(email))

		if r == nil {
			return errors.New("no such email")
		}

		json.Unmarshal(r, &record)
		ok = true
		return nil
	})

	return record, ok
}

func (d *database) Set(record Record) (ok bool) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		value, err := json.Marshal(record)
		if err != nil {
			return err
		}
		return b.Put([]byte(record.Email), value)
	})

	return err == nil
}

func (d *database) Close() error {
	return d.db.Close()
}
