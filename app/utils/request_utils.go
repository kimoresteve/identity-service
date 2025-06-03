// utils/request_utils.go
package utils

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func DecodeAndValidateJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// 1. Check content type
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return NewAPIError("Content-Type must be application/json", http.StatusUnsupportedMediaType, nil)
	}

	// 2. Limit request size
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB

	// 3. Decode JSON
	dec := json.NewDecoder(r.Body)

	//fmt.Printf("body", dec)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return NewAPIError("Invalid request payload", http.StatusBadRequest, err)
	}

	// 4. Validate struct if it has validate tags
	if isValidateable(dst) {
		if err := validate.Struct(dst); err != nil {
			return NewAPIError("Validation failed", http.StatusBadRequest, err)
		}
	}

	return nil
}

func isValidateable(dst interface{}) bool {
	t := reflect.TypeOf(dst)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return false
	}

	// Check if struct has any validate tags
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup("validate"); ok {
			return true
		}
	}
	return false
}

type APIError struct {
	Message string
	Status  int
	Err     error
}

func NewAPIError(message string, status int, err error) *APIError {
	return &APIError{
		Message: message,
		Status:  status,
		Err:     err,
	}
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
