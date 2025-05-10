package dto

// Response adalah model untuk response API
type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse adalah model untuk response error
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}
