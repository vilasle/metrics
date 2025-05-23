package middleware

import "errors"

var ErrInvalidKeyType = errors.New("invalid hash key type")
var ErrInvalidHashSum = errors.New("invalid hash sum")
