package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// revive:disable
type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type wrapError struct {
	Error HTTPError `json:"error"`
}

// Error prints errors to users as json and log server errors to stdout.
func Error(w http.ResponseWriter, r *http.Request, message error, code int) {
	err := HTTPError{
		Message: message.Error(),
		Code:    code,
	}

	if code == http.StatusInternalServerError {
		logError(r, message)
		err.Message = "Internal server error"
	}

	// this does not seem right but simplifies error handling
	if message == ErrMaxBytes {
		err.Code = http.StatusRequestEntityTooLarge
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&wrapError{err})
}

func logError(r *http.Request, err error) {
	log.Printf("[ERROR]: %s %s: %s", r.Method, r.URL.Path, err)
}
