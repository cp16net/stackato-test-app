package common

import (
	"errors"

	"github.com/hpcloud/hsm/generated/hsm/models"
)

// CreateAndLogServiceError creates an error for use with generate service code and logs it
func CreateAndLogServiceError(statusCode int, message string) *models.Error {
	e := CreateServiceError(statusCode, message)
	Logger.Errorf("Returning error: %d %s", statusCode, message)
	return e
}

// CreateServiceError creates an error for use with generate service code
func CreateServiceError(statusCode int, message string) *models.Error {
	statusCodeInt64 := int64(statusCode)
	return &models.Error{
		Code:    &statusCodeInt64,
		Message: message,
	}
}

// CreateAndLogServiceErrorWithDetail creates an error for use with generate service code and logs it with details
func CreateAndLogServiceErrorWithDetail(statusCode int, message, details string) *models.Error {
	err := CreateServiceErrorWithDetail(statusCode, message, details)
	if details != "" {
		details = "no details available"
	}
	Logger.Errorf("Returning error - code: %d message: %s detail: %s", statusCode, message, details)
	return err
}

// CreateServiceErrorWithDetail creates an error for use with generate service code with details
func CreateServiceErrorWithDetail(statusCode int, message, details string) *models.Error {
	statusCodeInt64 := int64(statusCode)
	err := models.Error{
		Code:    &statusCodeInt64,
		Message: message,
	}
	if details != "" {
		err.Details = &details
	}
	return &err
}

// ConsolidateErrors consolidates all errors to one error.
func ConsolidateErrors(allErrors ...error) error {
	if allErrors == nil || len(allErrors) == 0 {
		return nil
	}

	consolidatedError := ""
	for _, innerError := range allErrors {
		if consolidatedError == "" {
			consolidatedError = innerError.Error()
		} else {
			consolidatedError = consolidatedError + ": " + innerError.Error()
		}
	}

	if len(allErrors) > 0 {
		return errors.New(consolidatedError)
	}

	return nil
}
