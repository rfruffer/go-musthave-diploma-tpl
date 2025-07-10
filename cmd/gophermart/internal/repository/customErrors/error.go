package customerrors

import "errors"

var ErrOrderUploadedByAnotherUser = errors.New("Order Uploaded By Another User")
var ErrOrderAlreadyUploadedBySameUser = errors.New("Order Uploaded By Same User")
var ErrInsufficientBalance = errors.New("Insufficient Balance")
var ErrInvalidOrderNumber = errors.New("Invalid order number")
