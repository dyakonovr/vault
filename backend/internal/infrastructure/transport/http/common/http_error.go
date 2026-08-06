package common

type HttpError struct {
	ClientMessage string
	Code          string
	StatusCode    int
}

func (e *HttpError) Error() string {
	return e.ClientMessage
}