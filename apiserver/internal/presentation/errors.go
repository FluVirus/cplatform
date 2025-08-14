package presentation

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
