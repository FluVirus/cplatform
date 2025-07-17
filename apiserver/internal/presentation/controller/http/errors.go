package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return e.Message
}

var (
	ErrInvalidJsonSchema = &ApiError{Code: 000, Message: "invalid Json Schema"}
	ErrInvalidEmail      = &ApiError{Code: 001, Message: "invalid email"}
	ErrDuplicateEmail    = &ApiError{Code: 002, Message: "duplicate email"}
	ErrInvalidPassword   = &ApiError{Code: 003, Message: "invalid password"}
	ErrInvalidName       = &ApiError{Code: 004, Message: "invalid name"}
	ErrCancelled         = &ApiError{Code: 900, Message: "cancelled"}
	ErrDeadlineExceeded  = &ApiError{Code: 901, Message: "deadline exceeded"}
	ErrUnknown           = &ApiError{Code: 999, Message: "unknown error"}
)

type ErrorDescription struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type ErrorResponse struct {
	Errors []ErrorDescription `json:"errors"`
}

func WriteErrors(w http.ResponseWriter, status int, errs ...error) error {
	var res ErrorResponse
	var useInternalServerError bool

	for _, err := range errs {
		var apiErr *ApiError
		var description ErrorDescription

		if errors.As(err, &apiErr) {
			description = ErrorDescription{
				Code:    apiErr.Code,
				Message: err.Error(),
			}
		} else {
			description = ErrorDescription{
				Code:    ErrUnknown.Code,
				Message: fmt.Sprintf("%s: %s", ErrUnknown.Message, err.Error()),
			}

			useInternalServerError = true
		}

		res.Errors = append(res.Errors, description)
	}

	jsonBytes, jsonErr := json.Marshal(res)
	if jsonErr != nil {
		return fmt.Errorf("error marshalling error response: %v", jsonErr)
	}

	if useInternalServerError {
		status = http.StatusInternalServerError
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("error writing error response: %v", err)
	}

	return nil
}
