package httpapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

func readJSON[T any](r *http.Request) (T, *RequestError) {
	var zero T
	var input T

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return zero, ErrInvalidContentType
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		return zero, ErrInvalidBody
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return zero, ErrMultipleJSONValues
		}

		return zero, ErrInvalidBody
	}

	return input, nil
}
