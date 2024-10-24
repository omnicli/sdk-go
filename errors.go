package omnicli

import "fmt"

// ArgListMissingError is returned when the OMNI_ARG_LIST environment variable is missing.
type ArgListMissingError struct{}

func (e *ArgListMissingError) Error() string {
	return "OMNI_ARG_LIST environment variable is not set. " +
		"Are you sure \"argparser: true\" is set for this command?"
}

// TypeMismatchError is returned when an argument's type doesn't match the struct field
// type. This can happen when the declared type in environment variables doesn't match
// the Go struct field type.
type TypeMismatchError struct {
	fieldName    string
	expectedType string
	receivedType string
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch for field %q: expected %s, got %s",
		e.fieldName, e.expectedType, e.receivedType)
}

// InvalidBooleanValueError is returned when a boolean value cannot be parsed.
// Only "true" and "false" (case insensitive) are valid boolean values.
type InvalidBooleanValueError struct {
	message string
}

func (e *InvalidBooleanValueError) Error() string {
	return e.message
}

// InvalidIntegerValueError is returned when an integer value cannot be parsed.
type InvalidIntegerValueError struct {
	message string
}

func (e *InvalidIntegerValueError) Error() string {
	return e.message
}

// InvalidFloatValueError is returned when a float value cannot be parsed.
type InvalidFloatValueError struct {
	message string
}

func (e *InvalidFloatValueError) Error() string {
	return e.message
}
