package domain

import "fmt"

type ErrorCode string

const (
	CodeValidation ErrorCode = "validation"
	CodeNotFound   ErrorCode = "not_found"
	CodeConflict   ErrorCode = "conflict"
	CodeStorage    ErrorCode = "storage"
	CodePermission ErrorCode = "permission"
)

type AppError struct {
	Code ErrorCode
	Op   string
	Err  error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Op)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Op, e.Err)
}

func (e *AppError) Unwrap() error { return e.Err }

func Wrap(code ErrorCode, op string, err error) error {
	if err == nil {
		return nil
	}
	return &AppError{Code: code, Op: op, Err: err}
}

func NewError(code ErrorCode, op, message string) error {
	return &AppError{Code: code, Op: op, Err: fmt.Errorf("%s", message)}
}

func IsCode(err error, code ErrorCode) bool {
	for current := err; current != nil; {
		if app, ok := current.(*AppError); ok {
			return app.Code == code
		}
		u, ok := current.(interface{ Unwrap() error })
		if !ok {
			break
		}
		current = u.Unwrap()
	}
	return false
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if app, ok := err.(*AppError); ok && app.Err != nil {
		return app.Err.Error()
	}
	return err.Error()
}
