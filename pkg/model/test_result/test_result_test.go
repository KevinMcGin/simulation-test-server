package test_result

import (
    "testing"
	"simulation-test-server/pkg/model/test_result/test_status"
)

func TestTestResultString(t *testing.T) {
	testResult := TestResult{
		TestStatus: test_status.Running,
		Message: "Tests are runing",
		ExpiryEpochSeconds: 1234567890,
	}
	expected := "TestStatus: RUNNING\nMessage: Tests are runing\nExpiryEpochSeconds: 1234567890\n"
	if testResult.String() != expected {
		t.Fatalf(`testResult.String() = %v; want %v`, testResult.String(), expected)
	}
}