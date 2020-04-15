package utility

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertEquals fails if want is not equal to got
func AssertEquals(t *testing.T, want interface{}, got interface{}, message string) {
	if reflect.DeepEqual(got, want) {
		return
	}

	if len(message) == 0 {
		message = fmt.Sprintf("Expected '%v' but got '%v'", want, got)
	} else {
		message = fmt.Sprintf("%s: Expected '%v' but got '%v'", message, want, got)
	}
	t.Fatal(message)
}

// AssertNotEquals fails if want is equal to got
func AssertNotEquals(t *testing.T, want interface{}, got interface{}, message string) {
	if !reflect.DeepEqual(got, want) {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("Didn't expect '%v'", want)
	} else {
		message = fmt.Sprintf("%s: Expected '%v' but got '%v'", message, want, got)
	}
	t.Fatal(message)
}

// AssertGte fails if want is less than than got
func AssertGte(t *testing.T, want int, got int, message string) {
	if want <= got {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("Expected '%v' <= '%v'", want, got)
	} else {
		message = fmt.Sprintf("%s: Expected '%v' <= '%v'", message, want, got)
	}
	t.Fatal(message)
}

// AssertTrue fails if status is not true
func AssertTrue(t *testing.T, status bool, message string) {
	if status {
		return
	}
	t.Fatal(message)
}
