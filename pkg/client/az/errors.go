package az

import "github.com/pkg/errors"

var (
	ErrNilResponseFromAPI = errors.New("nil response from API")
	ErrTooManyDocuments   = errors.New("too many documents")
)
