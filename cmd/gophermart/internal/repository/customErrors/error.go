package customerrors

import "errors"

var ErrOrderUploadedByAnotherUser = errors.New("order uploaded by another user")
var ErrOrderAlreadyUploadedBySameUser = errors.New("order uploaded by same user")
var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrInvalidOrderNumber = errors.New("invalid order number")
