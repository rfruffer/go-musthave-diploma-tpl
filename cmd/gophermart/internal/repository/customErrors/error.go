package customErrors

import "errors"

var ErrAlreadyExists = errors.New("short url already exists")
var ErrGone = errors.New("URL is deleted")
var ErrOrderUploadedByAnotherUser = errors.New("UOrder Uploaded By Another User")
var ErrOrderAlreadyUploadedBySameUser = errors.New("UOrder Uploaded By Same User")
