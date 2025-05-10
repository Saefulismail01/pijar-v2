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

// PaginatedResponse adalah model untuk response API dengan pagination
type PaginatedResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int64       `json:"total"`
}
