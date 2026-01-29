// Package testutil provides shared test utilities for the Yutani project.
package testutil

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(b bool) *bool { return &b }

// StrPtr returns a pointer to the given string value.
func StrPtr(s string) *string { return &s }

// Int32Ptr returns a pointer to the given int32 value.
func Int32Ptr(i int32) *int32 { return &i }
