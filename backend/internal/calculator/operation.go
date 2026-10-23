// Package calculator holds the arithmetic domain. It knows nothing about HTTP.
package calculator

// Operation is a named arithmetic function with a fixed number of operands.
// Apply may assume len(operands) == Arity; Registry.Execute enforces it.
type Operation struct {
	Name  string
	Arity int
	Apply func(operands []float64) (float64, error)
}

// OperationInfo describes an operation without exposing its implementation.
type OperationInfo struct {
	Name  string
	Arity int
}
