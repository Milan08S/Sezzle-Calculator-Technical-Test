package httpapi

// calculateRequest uses pointers so a missing or null value is distinguishable
// from 0 (a plain float64 would silently decode null as 0).
type calculateRequest struct {
	Operation *string    `json:"operation"`
	Operands  []*float64 `json:"operands"`
}

type calculateResponse struct {
	Operation string    `json:"operation"`
	Operands  []float64 `json:"operands"`
	Result    float64   `json:"result"`
}

type operationsResponse struct {
	Operations []operationDTO `json:"operations"`
}

type operationDTO struct {
	Name  string `json:"name"`
	Arity int    `json:"arity"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
