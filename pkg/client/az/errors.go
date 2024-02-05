package az

import "github.com/pkg/errors"

var (
	ErrNilResponseFromAPI               error = errors.New("nil response from API")
	ErrTooFewDocumentsForEntityRequest  error = errors.New("cannot send request with no documents")
	ErrTooManyDocumentsForEntityRequest error = errors.New("too many documents for one request")
)
