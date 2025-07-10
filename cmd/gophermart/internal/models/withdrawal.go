package models

import "time"

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type WithdrawalRequest struct {
	Order string  `json:"order" binding:"required"`
	Sum   float64 `json:"sum" binding:"required"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
