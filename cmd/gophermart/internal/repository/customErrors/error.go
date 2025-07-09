package customErrors

import "errors"

var ErrOrderUploadedByAnotherUser = errors.New("UOrder Uploaded By Another User")
var ErrOrderAlreadyUploadedBySameUser = errors.New("UOrder Uploaded By Same User")
