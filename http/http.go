package http

import (
	"log"
	"net/http"

	"github.com/mortezadadgar/ecommerce-api/domain"
)

type wrapError struct {
	Error domain.Error `json:"error"`
}

// Error prints errors to users as json and log server errors to stdout.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	code, message := domain.ErrorCode(err), domain.ErrorMessage(err)

	if code == domain.EINTERNAL {
		logError(r, err)
	}

	_ = ToJSON(w, wrapError{domain.Error{Message: message, Code: code}}, code)
}

func logError(r *http.Request, err error) {
	log.Printf("[ERROR]: %s %s: %s", r.Method, r.URL.Path, err)
}

// ErrorInvalidQuery uese Error to reports invalid url quries.
func ErrorInvalidQuery(w http.ResponseWriter, r *http.Request) {
	Error(w, r, domain.Errorf(domain.EINVALID, "invalid url query"))
}
