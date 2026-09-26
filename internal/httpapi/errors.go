package httpapi

import "net/http"

type RequestError struct {
	status  int
	message string
}

func (err *RequestError) Error() string {
	return err.message
}

func (err *RequestError) Status() int {
	return err.status
}

var (
	ErrInvalidContentType = &RequestError{
		status:  http.StatusUnsupportedMediaType,
		message: "Content-Type must be application/json",
	}

	ErrInvalidBody = &RequestError{
		status:  http.StatusBadRequest,
		message: "invalid JSON request body",
	}

	ErrMultipleJSONValues = &RequestError{
		status:  http.StatusBadRequest,
		message: "request body must contain a single JSON value",
	}
)
