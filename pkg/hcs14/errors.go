package hcs14

import "errors"

var (
	ErrInvalidBase58Character = errors.New("invalid base58 character")
	ErrInvalidMultibase       = errors.New("invalid multibase base58btc")
)
