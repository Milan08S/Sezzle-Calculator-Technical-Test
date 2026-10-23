package calculator

import (
	"cmp"
	"fmt"
	"math"
	"slices"
)

// Registry looks up operations by name and applies the rules shared by all of them.
// It is read-only after construction, so it is safe for concurrent use.
type Registry struct {
	ops map[string]Operation
}

// NewRegistry returns a Registry with all built-in operations.
func NewRegistry() *Registry {
	r := &Registry{ops: make(map[string]Operation)}
	for _, op := range defaultOperations() {
		r.ops[op.Name] = op
	}
	return r
}

// Execute runs the named operation on operands.
func (r *Registry) Execute(name string, operands []float64) (float64, error) {
	op, ok := r.ops[name]
	if !ok {
		return 0, fmt.Errorf("%q: %w", name, ErrUnknownOperation)
	}
	if len(operands) != op.Arity {
		return 0, fmt.Errorf("%s expects %d operand(s), got %d: %w", name, op.Arity, len(operands), ErrInvalidArity)
	}
	result, err := op.Apply(operands)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, fmt.Errorf("%s: %w", name, ErrNonFiniteResult)
	}
	return result, nil
}

// Operations lists every operation's name and arity, sorted by name.
func (r *Registry) Operations() []OperationInfo {
	infos := make([]OperationInfo, 0, len(r.ops))
	for _, op := range r.ops {
		infos = append(infos, OperationInfo{Name: op.Name, Arity: op.Arity})
	}
	slices.SortFunc(infos, func(a, b OperationInfo) int { return cmp.Compare(a.Name, b.Name) })
	return infos
}
