package controller

import (
	"context"
	"cplatform/internal/application/contracts/application"
	"cplatform/internal/di/middleware"
	"cplatform/internal/presentation"
	presentation_http "cplatform/internal/presentation/http"
	"cplatform/pkg/slogext"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"unicode/utf8"
)

const minNameLength = 1
const maxNameLength = 85
const minPasswordLength = 4
const maxPasswordLength = 32

var nameAlphabet map[rune]struct{}

func init() {
	nameAlphabet = make(map[rune]struct{})
	intervals := [][2]rune{
		{'a', 'z'},
		{'A', 'Z'},
		{'0', '9'},
	}

	for _, interval := range intervals {
		for i := interval[0]; i <= interval[1]; i++ {
			nameAlphabet[i] = struct{}{}
		}
	}
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (c *Controller) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr := fmt.Errorf("%w: %s", presentation.ErrInvalidJsonSchema, slogext.Cause(err))

		err = presentation_http.WriteErrors(w, http.StatusBadRequest, writeErr)
		if err != nil {
			c.logger.Error("fail write json schema error", slogext.Cause(err))
		}

		return
	}

	if errs := validateRequest(&req); len(errs) > 0 {
		c.logger.Error("fail on validate data", "causes", errs)

		err := presentation_http.WriteErrors(w, http.StatusBadRequest, errs...)
		if err != nil {
			c.logger.Error("fail write validation error", slogext.Cause(err))
		}

		return
	}

	scope := middleware.GetScope(r.Context())
	if scope == nil {
		// TODO: migrate to coded api errors
		c.logger.Error("fail to get scope during user registration")
		return
	}

	err := scope.UserService(r.Context()).RegisterUser(r.Context(), req.Name, req.Email, req.Password)

	if err != nil {
		var status int
		var sendError error

		if errors.Is(err, application.ErrDuplicateEmail) {
			status = http.StatusConflict
			sendError = presentation.ErrDuplicateEmail

			c.logger.Error("fail register new user", slogext.Cause(err))
		} else if errors.Is(err, context.Canceled) {
			status = http.StatusRequestTimeout
			sendError = presentation.ErrCancelled

			c.logger.Error("fail register new user due cancellation", slogext.Cause(err))
		} else if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusRequestTimeout
			sendError = presentation.ErrDeadlineExceeded

			c.logger.Error("fail register new user due deadline", slogext.Cause(err))
		} else {
			status = http.StatusBadRequest
			sendError = err

			c.logger.Error("fail register new user with unexpected error", slogext.Cause(err))
		}

		if writeErr := presentation_http.WriteErrors(w, status, sendError); writeErr != nil {
			c.logger.Error("fail write conflict error", slogext.Cause(writeErr))
		}

		return
	}

	err = scope.UnitOfWork(r.Context()).SaveChanges(r.Context())
	if err != nil {
		c.logger.Error("fail save changes", slogext.Cause(err))
		writeErr := presentation_http.WriteErrors(w, http.StatusInternalServerError, err)
		if writeErr != nil {
			c.logger.Error("fail write conflict error", slogext.Cause(writeErr))
		}

		return
	}

	c.logger.Info("user created", "email", req.Email)

	w.WriteHeader(http.StatusCreated)
}

func validateRequest(req *RegisterUserRequest) []error {
	var errs []error

	errs = append(errs, validateEmail(req.Email)...)
	errs = append(errs, validateName(req.Name)...)
	errs = append(errs, validatePassword(req.Password)...)

	return errs
}

func validateEmail(email string) []error {
	var errs []error

	_, err := mail.ParseAddress(email)
	if err != nil {
		errs = append(errs, fmt.Errorf("%w: %s", presentation.ErrInvalidEmail, err))
	}

	return errs
}

func validateName(name string) []error {
	var errs []error

	length := utf8.RuneCountInString(name)

	if length < minNameLength {
		errs = append(errs, fmt.Errorf("%w: name too short", presentation.ErrInvalidName))
	} else if length > maxNameLength {
		errs = append(errs, fmt.Errorf("%w: name too long", presentation.ErrInvalidName))
	}

	if !allRunesInAlphabet(name, nameAlphabet) {
		errs = append(errs, fmt.Errorf("%w: name contains unexpected chars", presentation.ErrInvalidName))
	}

	return errs
}

func allRunesInAlphabet(s string, alphabet map[rune]struct{}) bool {
	for _, r := range s {
		_, ok := alphabet[r]
		if !ok {
			return false
		}
	}

	return true
}

func validatePassword(password string) []error {
	var errs []error

	if len(password) < minPasswordLength {
		errs = append(errs, fmt.Errorf("%w: password too short", presentation.ErrInvalidPassword))
	}

	if len(password) > maxPasswordLength {
		errs = append(errs, fmt.Errorf("%w: password too long", presentation.ErrInvalidPassword))
	}

	return errs
}
