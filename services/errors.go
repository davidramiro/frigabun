package services

import "errors"

var (
	ErrMissingInfoForServiceInit = errors.New("cannot setup service, missing config param")
	ErrRegistrarNotFound         = errors.New("registrar not found")
)
