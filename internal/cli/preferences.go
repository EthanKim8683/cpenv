package cli

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var (
	stateBucketKey     = []byte("state")
	defaultTemplateKey = []byte("default_template")
)

type Preferences interface {
	DefaultTemplate() (string, error)
	SetDefaultTemplate(path string) error
}

type DBPreferences struct {
	DB *bolt.DB
}

func (s *DBPreferences) get(k []byte) ([]byte, error) {
	var v []byte
	if err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(stateBucketKey)
		if b == nil {
			return nil
		}
		v = b.Get(k)
		return nil
	}); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *DBPreferences) set(k []byte, v []byte) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(stateBucketKey)
		if err != nil {
			return err
		}
		return b.Put(k, v)
	})
}

func (s *DBPreferences) DefaultTemplate() (string, error) {
	v, err := s.get(defaultTemplateKey)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (s *DBPreferences) SetDefaultTemplate(path string) error {
	if err := s.set(defaultTemplateKey, []byte(path)); err != nil {
		return fmt.Errorf("set default template: %w", err)
	}
	return nil
}

var _ Preferences = (*DBPreferences)(nil)
