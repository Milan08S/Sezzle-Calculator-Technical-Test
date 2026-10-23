package calculator_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
)

func TestRegistryExecuteRules(t *testing.T) {
	tests := []struct {
		name        string
		op          string
		operands    []float64
		wantErr     error
		wantContext string // substring the wrapped message must contain
	}{
		{"unknown operation", "modulo", []float64{1, 2}, calculator.ErrUnknownOperation, `"modulo"`},
		{"empty operation name", "", []float64{1, 2}, calculator.ErrUnknownOperation, `""`},
		{"operation names are case sensitive", "ADD", []float64{1, 2}, calculator.ErrUnknownOperation, `"ADD"`},
		{"too few operands", "add", []float64{1}, calculator.ErrInvalidArity, "add expects 2 operand(s), got 1"},
		{"too many operands", "sqrt", []float64{1, 2}, calculator.ErrInvalidArity, "sqrt expects 1 operand(s), got 2"},
		{"nil operands", "divide", nil, calculator.ErrInvalidArity, "got 0"},
		{"apply error is wrapped", "divide", []float64{1, 0}, calculator.ErrDivisionByZero, "divide: "},
		{"overflow to +Inf", "multiply", []float64{1e308, 10}, calculator.ErrNonFiniteResult, "multiply: "},
		{"overflow to -Inf", "add", []float64{-1e308, -1e308}, calculator.ErrNonFiniteResult, "add: "},
		{"power of zero to negative is Inf", "power", []float64{0, -1}, calculator.ErrNonFiniteResult, "power: "},
		{"power yielding NaN", "power", []float64{-8, 0.5}, calculator.ErrNonFiniteResult, "power: "},
	}

	r := calculator.NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Execute(tt.op, tt.operands)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantContext) {
				t.Errorf("error %q does not contain %q", err, tt.wantContext)
			}
			if got != 0 {
				t.Errorf("result = %v, want 0 on error", got)
			}
		})
	}
}

func TestRegistryOperations(t *testing.T) {
	want := []calculator.OperationInfo{
		{Name: "add", Arity: 2},
		{Name: "divide", Arity: 2},
		{Name: "multiply", Arity: 2},
		{Name: "percent", Arity: 2},
		{Name: "power", Arity: 2},
		{Name: "sqrt", Arity: 1},
		{Name: "subtract", Arity: 2},
	}
	got := calculator.NewRegistry().Operations()
	if !slices.Equal(got, want) {
		t.Errorf("Operations() = %v, want %v", got, want)
	}
}
