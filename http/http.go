package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type wrapError struct {
	Error any `json:"error"`
}

func Error(w http.ResponseWriter, r *http.Request, message error, code int) {
	err := HTTPError{
		Message: message.Error(),
		Code:    code,
	}

	if code == http.StatusInternalServerError {
		logError(r, message)
		err.Message = "Internal server error"
	}

	if message == ErrMaxBytes {
		err.Code = http.StatusRequestEntityTooLarge
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&wrapError{err})
}

func logError(r *http.Request, err error) {
	log.Printf("[ERROR]: %s %s: %s", r.Method, r.URL.Path, err)
}
