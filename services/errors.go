package services

import "errors"

var (
	ErrMissingInfoForServiceInit = errors.New("cannot setup service, missing config param")
	ErrRegistrarNotFound         = errors.New("registrar not found")
	ErrBuildingRequest           = errors.New("error building request")
	ErrParsingResponse           = errors.New("error parsing api response")
	ErrRegistrarRejectedRequest  = errors.New("registrar rejected request")
	ErrExecutingRequest          = errors.New("error executing request")
)
