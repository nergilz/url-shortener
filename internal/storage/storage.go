package storage

import "errors"

// TODO: перенести в пакет common
var (
	ErrUrlNotFound = errors.New("url not found")
	ErrUrlExists   = errors.New("url exists")
)
