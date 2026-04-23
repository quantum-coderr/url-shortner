package domain

import "errors"

var (
	ErrNotFound     = errors.New("short url not found")
	ErrDuplicateKey = errors.New("short key already exists")
	ErrInvalidURL   = errors.New("invalid url")
)
