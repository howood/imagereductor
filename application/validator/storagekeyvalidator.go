package validator

import (
	"errors"
	"path"
	"strings"
)

// ErrInvalidStorageKey is returned when a storage key contains path traversal.
var ErrInvalidStorageKey = errors.New("invalid storage key")

// StorageKeyValidator validates storage keys for path traversal attacks.
type StorageKeyValidator struct{}

// NewStorageKeyValidator creates a new StorageKeyValidator.
func NewStorageKeyValidator() *StorageKeyValidator {
	return &StorageKeyValidator{}
}

// Validate checks that the given key does not contain path traversal sequences.
func (v *StorageKeyValidator) Validate(key string) error {
	cleaned := path.Clean(key)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") {
		return ErrInvalidStorageKey
	}
	return nil
}
