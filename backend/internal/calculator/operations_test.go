package calculator_test

import (
	"errors"
	"testing"

	"github.com/Milan08S/Sezzle-Calculator-Technical-Test/backend/internal/calculator"
)

func TestOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		operands []float64
		want     float64
		wantErr  error
	}{
		{"add positives", "add", []float64{2, 3}, 5, nil},
		{"add negatives", "add", []float64{-2, -3}, -5, nil},
		{"add decimals", "add", []float64{0.1, 0.2}, 0.30000000000000004, nil},
		{"add zero", "add", []float64{7, 0}, 7, nil},
		{"add large", "add", []float64{9e15, 1e15}, 1e16, nil},

		{"subtract positives", "subtract", []float64{10, 4}, 6, nil},
		{"subtract to negative", "subtract", []float64{4, 10}, -6, nil},
		{"subtract negatives", "subtract", []float64{-4, -10}, 6, nil},
		{"subtract decimals", "subtract", []float64{1.5, 0.25}, 1.25, nil},
		{"subtract zero", "subtract", []float64{0, 0}, 0, nil},

		{"multiply positives", "multiply", []float64{6, 7}, 42, nil},
		{"multiply mixed signs", "multiply", []float64{-6, 7}, -42, nil},
		{"multiply decimals", "multiply", []float64{0.5, 0.5}, 0.25, nil},
		{"multiply by zero", "multiply", []float64{123, 0}, 0, nil},
		{"multiply large", "multiply", []float64{1 << 40, 1 << 40}, 1 << 80, nil},

		{"divide positives", "divide", []float64{10, 2}, 5, nil},
		{"divide negatives", "divide", []float64{-10, -4}, 2.5, nil},
		{"divide decimals", "divide", []float64{1, 3}, 1.0 / 3, nil},
		{"divide zero numerator", "divide", []float64{0, 5}, 0, nil},
		{"divide by zero", "divide", []float64{1, 0}, 0, calculator.ErrDivisionByZero},
		{"divide zero by zero", "divide", []float64{0, 0}, 0, calculator.ErrDivisionByZero},

		{"power positive", "power", []float64{2, 10}, 1024, nil},
		{"power negative base", "power", []float64{-2, 3}, -8, nil},
		{"power negative exponent", "power", []float64{2, -2}, 0.25, nil},
		{"power fractional exponent", "power", []float64{9, 0.5}, 3, nil},
		{"power zero exponent", "power", []float64{5, 0}, 1, nil},

		{"sqrt perfect square", "sqrt", []float64{16}, 4, nil},
		{"sqrt decimal", "sqrt", []float64{2.25}, 1.5, nil},
		{"sqrt zero", "sqrt", []float64{0}, 0, nil},
		{"sqrt large", "sqrt", []float64{1e300}, 1e150, nil},
		{"sqrt negative", "sqrt", []float64{-4}, 0, calculator.ErrNegativeSqrt},

		{"percent of positive", "percent", []float64{200, 15}, 30, nil},
		{"percent negative", "percent", []float64{200, -10}, -20, nil},
		{"percent decimal", "percent", []float64{50, 12.5}, 6.25, nil},
		{"percent of zero", "percent", []float64{0, 50}, 0, nil},
	}

	r := calculator.NewRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Execute(tt.op, tt.operands)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute(%q, %v) error = %v, want %v", tt.op, tt.operands, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Execute(%q, %v) = %v, want %v", tt.op, tt.operands, got, tt.want)
			}
		})
	}
}
