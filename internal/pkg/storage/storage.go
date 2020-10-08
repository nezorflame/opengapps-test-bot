package storage

import (
	"errors"
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

// Package errors
var (
	ErrNotFound = errors.New("key not found")
	ErrNilValue = errors.New("value is nil")
)

var bucketName = []byte("global")

type storage struct {
	db      *bbolt.DB
	timeout time.Duration
}

// New creates new instance of storage
func New(path string, timeout time.Duration) (*storage, error) {
	// open connection to the DB
	log.WithField("path", path).WithField("timeout", timeout).Debug("Creating DB connection")
	opts := bbolt.DefaultOptions
	if timeout > 0 {
		opts.Timeout = timeout
	}
	b, err := bbolt.Open(path, 0755, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB: %w", err)
	}

	// create global bucket if it doesn't exist yet
	log.WithField("bucket", string(bucketName)).Debug("Setting the default bucket")
	err = b.Update(func(tx *bbolt.Tx) error {
		_, bErr := tx.CreateBucketIfNotExists(bucketName)
		return bErr
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create global bucket: %w", err)
	}

	// return the DB
	db := &storage{db: b, timeout: timeout}
	log.Debug("DB initiated")
	return db, nil
}

// Keys returns a list of available keys in the global bucket, sorted alphabetically
func (s *storage) Keys() ([]string, error) {
	var keys []string
	log.Debug("Getting the list of DB current keys")
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.ForEach(func(k, v []byte) error {
			if v != nil {
				keys = append(keys, string(k))
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of keys from DB: %w", err)
	}
	sort.Strings(keys)
	return keys, nil
}

// Get acquires value from storage by provided key
func (s *storage) Get(key string) ([]byte, error) {
	var value []byte
	log.WithField("key", key).Debug("Getting value from DB")
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		k, v := b.Cursor().Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrNotFound
		} else if v == nil {
			return ErrNilValue
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get value for key '%s' from DB: %w", key, err)
	}
	log.WithField("key", key).Debug("Got the value")
	return value, nil
}

// Put sets/updates the value in storage by provided bucket and key
func (s *storage) Put(key string, val []byte) error {
	log.WithField("key", key).Debug("Saving the value to DB")
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), val)
	})
	if err != nil {
		return fmt.Errorf("unable to put value for key '%s' to DB: %w", key, err)
	}
	return nil
}

// Delete removes the value from storage by provided bucket and key
func (s *storage) Delete(key string) error {
	log.WithField("key", key).Debug("Deleting from DB")
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("unable to delete value for key '%s' from DB: %w", key, err)
	}
	return nil
}

// Purge removes the bucket from storage
func (s *storage) Purge() error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		return fmt.Errorf("unable to purge global bucket from DB: %w", err)
	}
	return nil
}

// Close closes the storage
func (s *storage) Close() error {
	log.Debug("Closing the DB")
	done := make(chan error)
	go func() {
		done <- s.db.Close()
		log.Debug("DB closed OK")
		close(done)
	}()
	timer := time.NewTimer(s.timeout)

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("unable to close DB: %w", err)
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("unable to close DB: %w", bbolt.ErrTimeout)
	}
}

// Name returns the storage identifier
func (s *storage) Name() string {
	return "storage"
}
