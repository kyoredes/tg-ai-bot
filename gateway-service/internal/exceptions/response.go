package exceptions

import "errors"

var (
	ErrResponseExternalService = errors.New("Error response from external service")
)
