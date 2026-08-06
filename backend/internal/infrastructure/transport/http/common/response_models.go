package common

type ValidationError struct {
	Code  string `json:"code"`            // Код ошибки, например FIELD_REQUIRED
	Param string `json:"param,omitempty"` // Параметр (для min/max), например "8"
}
type ValidationErrors = map[string][]ValidationError // { "field_name": ["error1", "error2"] }

type ErrorResponse struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
	Errors    ValidationErrors `json:"errors,omitempty"`
}

type SuccessResponse struct {
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// type PaginatedMeta struct {
// 	Page       int `json:"page"`
// 	PerPage    int `json:"per_page"`
// 	Total      int `json:"total"`
// 	TotalPages int `json:"total_pages"`
// }

// type Paginated struct {
// 	Data      any           `json:"data,omitempty"`
// 	Meta      PaginatedMeta `json:"meta"`
// 	RequestID string        `json:"request_id,omitempty"`
// }
