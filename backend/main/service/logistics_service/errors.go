package logistics_service

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrAreaUnavailable   = errors.New("no data source covers this area")
	ErrNotFound          = errors.New("not found")
	ErrJobNotCancellable = errors.New("job cannot be cancelled")
	ErrJobCancelled      = errors.New("job cancelled")
)
