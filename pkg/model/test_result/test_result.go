package test_result

import (
	"fmt"
	"simulation-test-server/pkg/model/test_result/test_status"	
)

type TestResult struct {
	TestStatus 			test_status.TestStatus 	`json:"testStatus"`
	Message 			string 		`json:"message"`
	ExpiryEpochSeconds 	int64
}

func (t TestResult) String() string {
	return fmt.Sprintf(
		"TestStatus: %v\n" +
		"Message: %v\n" +
		"ExpiryEpochSeconds: %v\n", 
		t.TestStatus, t.Message, t.ExpiryEpochSeconds,
	)
}