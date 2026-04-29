package runclient

import (
	"errors"
	"github.com/yazanabuashour/openclerk/internal/domain"
)

// Error is the internal error shape returned by the internal runner client.
type Error struct {
	Code      string
	Message   string
	Status    int
	Retryable bool
	Details   map[string]any
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func wrapError(err error) error {
	if err == nil {
		return nil
	}
	var localErr *Error
	if errors.As(err, &localErr) {
		return localErr
	}
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return &Error{
			Code:      domainErr.Code,
			Message:   domainErr.Message,
			Status:    domainErr.Status,
			Retryable: domainErr.Retryable,
			Details:   cloneDetails(domainErr.Details),
		}
	}
	return &Error{
		Code:      "internal_error",
		Message:   err.Error(),
		Status:    500,
		Retryable: true,
	}
}

func cloneDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(details))
	for key, value := range details {
		cloned[key] = value
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
