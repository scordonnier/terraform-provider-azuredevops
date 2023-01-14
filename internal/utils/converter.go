package utils

import "github.com/google/uuid"

var EmptyString = String("")

// Bool Get a pointer to a boolean value
func Bool(value bool) *bool {
	return &value
}

// Int Get a pointer to an integer value
func Int(value int) *int {
	return &value
}

// String Get a pointer to a string
func String(value string) *string {
	return &value
}

// StringFromInterface get a string pointer from an interface
func StringFromInterface(value interface{}) *string {
	return String(value.(string))
}

// UUID Get a pointer to a UUID
func UUID(value string) *uuid.UUID {
	uuid := uuid.MustParse(value)
	return &uuid
}
