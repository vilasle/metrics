package rest

import "errors"

var ErrForbiddenResource = errors.New("forbidden resource")
var ErrUnknownContentType = errors.New("unknown content type")
var ErrReadingRequestBody = errors.New("can not read request body")
var ErrEmptyRequestBody = errors.New("empty request body")
var ErrEmptyRequiredFields = errors.New("empty required fields")
var ErrInvalidKeyType = errors.New("invalid hash key type")
var ErrInvalidHashSum = errors.New("invalid hash sum")