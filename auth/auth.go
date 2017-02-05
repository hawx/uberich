package auth

import (
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"hawx.me/code/uberich/config"
)

type Checker struct {
	conf   *config.Config
	logger *log.Logger

	mu       sync.Mutex
	bucketFn func() *rate.Limiter
	buckets  map[string]*rate.Limiter
}

func NewChecker(conf *config.Config, logger *log.Logger) *Checker {
	return &Checker{
		conf:     conf,
		logger:   logger,
		bucketFn: func() *rate.Limiter { return rate.NewLimiter(rate.Every(time.Second*30), 3) },
		buckets:  map[string]*rate.Limiter{},
	}
}

func (c *Checker) IsAuthorised(email, password string) bool {
	c.mu.Lock()
	bucket, ok := c.buckets[email]
	if !ok {
		bucket = c.bucketFn()
		c.buckets[email] = bucket
	}
	c.mu.Unlock()

	if !bucket.Allow() {
		c.logger.Println("checker: rate limit exceeded for", email)
		return false
	}

	user := c.conf.GetUser(email)
	if user == nil {
		c.logger.Println("checker: no such user", email)
		return false
	}

	if !user.IsPassword(password) {
		c.logger.Println("checker: password incorrect", email)
		return false
	}

	return true
}
