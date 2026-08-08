package common

type ValidationError struct {
	Code  string `json:"code"`            // Код ошибки, например FIELD_REQUIRED
	Param string `json:"param,omitempty"` // Параметр (для min/max), например "8"
}
type ValidationErrors = map[string][]ValidationError // { "field_name": ["error1", "error2"] }

type ErrorResponse struct {
	Message   string           `json:"message"`
	Code      string           `json:"code"`
	RequestID string           `json:"request_id,omitempty"`
	Errors    ValidationErrors `json:"errors,omitempty"`
}

type SuccessResponse struct {
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type PaginatedResponseMeta struct {
	Page       int64 `json:"page"`
	PerPage    int64 `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type PaginatedResponse struct {
	Data      any                   `json:"data,omitempty"`
	Meta      PaginatedResponseMeta `json:"meta"`
	RequestID string                `json:"request_id,omitempty"`
}
