package errors

import "errors"

var ErrAlreadyExists = errors.New("short url already exists")
var ErrGone = errors.New("URL is deleted")
