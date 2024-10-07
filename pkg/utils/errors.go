package utils

import "errors"

// Custom errors
var (
	ErrInvalidIP           = errors.New("invalid IP address")
	ErrIpNotFound          = errors.New("IP not found in database")
	ErrBuildResponse       = errors.New("error building response")
	ErrDatabaseQuery       = errors.New("error querying database")
	ErrInvalidFields       = errors.New("invalid fields requested")
	ErrJSONMarshal         = errors.New("error marshaling JSON")
	ErrJSONUnmarshal       = errors.New("error unmarshaling JSON")
	ErrInternalServer      = errors.New("internal server error")
	ErrUnsupportedIPFormat = errors.New("unsupported IP format")
	ErrMongoDB             = errors.New("error querying MongoDB")
)
