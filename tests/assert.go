package tests

import "testing"

func AssertPanic(t *testing.T, f func(), expectedError string) {
	t.Helper()

	defer func() {
		err := recover()
		if err != expectedError {
			t.Errorf("The code did not panic with expected error. Got: %s", err)
		}
	}()

	f()
}
