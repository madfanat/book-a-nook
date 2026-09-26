package httpapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

type requestError struct {
	status  int
	message string
}

func (err *requestError) Error() string {
	return err.message
}

func readJSON[T any](r *http.Request) (T, *requestError) {
	var zero T
	var input T

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return zero, &requestError{
			status:  http.StatusUnsupportedMediaType,
			message: "Content-Type must be application/json",
		}
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		return zero, &requestError{
			status:  http.StatusBadRequest,
			message: "invalid JSON request body",
		}
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return zero, &requestError{
				status:  http.StatusBadRequest,
				message: "request body must contain a single JSON value",
			}
		}

		return zero, &requestError{
			status:  http.StatusBadRequest,
			message: "invalid JSON request body",
		}
	}

	return input, nil
}
