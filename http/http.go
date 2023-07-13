package http

import (
	"fmt"
	"log"
	"net/http"
)

// errorTemplate error messages structure shown to users.
type errorTemplate struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// WrapError wraps error for user representation.
type WrapError struct {
	Error errorTemplate `json:"error"`
}

// Errorf prints errors to users as json and log server errors to stdout.
func Errorf(w http.ResponseWriter, r *http.Request, code int, format string, a ...any) {
	message := fmt.Sprintf(format, a...)

	if code == http.StatusInternalServerError {
		logError(r, message)
	}

	_ = ToJSON(w, WrapError{errorTemplate{Message: message, Code: code}}, code)
}

func logError(r *http.Request, err string) {
	log.Printf("[ERROR]: %s %s: %s", r.Method, r.URL.Path, err)
}

// ErrorInvalidQuery uese Error to reports invalid url quries.
func ErrorInvalidQuery(w http.ResponseWriter, r *http.Request) {
	Errorf(w, r, http.StatusBadRequest, "invalid url query")
}
