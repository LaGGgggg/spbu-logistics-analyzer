package logistics_service

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrAreaUnavailable = errors.New("no data source covers this area")
	ErrNotFound        = errors.New("not found")
	ErrGraphTooLarge   = errors.New("graph payload exceeds storage limit")
)
