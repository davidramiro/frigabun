package api

import "errors"

var (
	ErrCannotParseRequest = errors.New("cannot parse request")
	ErrMissingParameter   = errors.New("missing parameter")
	ErrInvalidIP          = errors.New("missing or invalid IP address, only IPv4 allowed")
	ErrInvalidDomain      = errors.New("missing or invalid domain name")
)
