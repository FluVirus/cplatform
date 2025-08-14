package http

import (
	"cplatform/internal/presentation"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		var apiErr *presentation.ApiError
		var description ErrorDescription

		if errors.As(err, &apiErr) {
			description = ErrorDescription{
				Code:    apiErr.Code,
				Message: err.Error(),
			}
		} else {
			description = ErrorDescription{
				Code:    presentation.ErrUnknown.Code,
				Message: fmt.Sprintf("%s: %s", presentation.ErrUnknown.Message, err.Error()),
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
