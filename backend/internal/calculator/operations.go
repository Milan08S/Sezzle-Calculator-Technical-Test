package calculator

import "math"

func add(o []float64) (float64, error)      { return o[0] + o[1], nil }
func subtract(o []float64) (float64, error) { return o[0] - o[1], nil }
func multiply(o []float64) (float64, error) { return o[0] * o[1], nil }

func divide(o []float64) (float64, error) {
	if o[1] == 0 {
		return 0, ErrDivisionByZero
	}
	return o[0] / o[1], nil
}

// power returns a^b. Results that are NaN (e.g. (-8)^0.5) or ±Inf are
// rejected later by Registry.Execute.
func power(o []float64) (float64, error) { return math.Pow(o[0], o[1]), nil }

func sqrt(o []float64) (float64, error) {
	if o[0] < 0 {
		return 0, ErrNegativeSqrt
	}
	return math.Sqrt(o[0]), nil
}

// percent returns "b percent of a", i.e. a * b / 100 (percent(200, 15) == 30).
func percent(o []float64) (float64, error) { return o[0] * o[1] / 100, nil }

func defaultOperations() []Operation {
	return []Operation{
		{Name: "add", Arity: 2, Apply: add},
		{Name: "subtract", Arity: 2, Apply: subtract},
		{Name: "multiply", Arity: 2, Apply: multiply},
		{Name: "divide", Arity: 2, Apply: divide},
		{Name: "power", Arity: 2, Apply: power},
		{Name: "sqrt", Arity: 1, Apply: sqrt},
		{Name: "percent", Arity: 2, Apply: percent},
	}
}
