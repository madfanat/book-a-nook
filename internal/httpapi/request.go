package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

const maxBodyBytes = 16 << 10

func readJSON[T any](w http.ResponseWriter, r *http.Request) (T, *RequestError) {
	var zero T
	var input T

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return zero, ErrInvalidContentType
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return zero, bodyError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return zero, ErrMultipleJSONValues
		}

		return zero, bodyError(err)
	}

	return input, nil
}

func bodyError(err error) *RequestError {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return ErrBodyTooLarge
	}
	return ErrInvalidBody
}
