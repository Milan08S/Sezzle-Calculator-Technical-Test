package calculator

import "errors"

var (
	ErrUnknownOperation = errors.New("unknown operation")
	ErrInvalidArity     = errors.New("invalid number of operands")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrNegativeSqrt     = errors.New("square root of a negative number")
	ErrNonFiniteResult  = errors.New("result is not a finite number")
)
