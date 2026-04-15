package domain

import "fmt"

type Error struct {
	Code      string
	Message   string
	Status    int
	Retryable bool
	Details   map[string]any
}

func (e *Error) Error() string {
	return e.Message
}

func ValidationError(message string, details map[string]any) *Error {
	return &Error{
		Code:    "validation_error",
		Message: message,
		Status:  400,
		Details: details,
	}
}

func NotFoundError(resource string, id string) *Error {
	return &Error{
		Code:    "not_found",
		Message: fmt.Sprintf("%s %q was not found", resource, id),
		Status:  404,
		Details: map[string]any{"resource": resource, "id": id},
	}
}

func AlreadyExistsError(resource string, id string) *Error {
	return &Error{
		Code:    "already_exists",
		Message: fmt.Sprintf("%s %q already exists", resource, id),
		Status:  409,
		Details: map[string]any{"resource": resource, "id": id},
	}
}

func UnsupportedError(feature string, backend BackendKind) *Error {
	return &Error{
		Code:    "unsupported",
		Message: fmt.Sprintf("%s is not available for backend %s", feature, backend),
		Status:  404,
		Details: map[string]any{"feature": feature, "backend": string(backend)},
	}
}

func InternalError(message string, err error) *Error {
	details := map[string]any{}
	if err != nil {
		details["cause"] = err.Error()
	}
	return &Error{
		Code:      "internal_error",
		Message:   message,
		Status:    500,
		Retryable: true,
		Details:   details,
	}
}
