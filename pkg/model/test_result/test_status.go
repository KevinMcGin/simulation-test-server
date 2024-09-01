package test_result

type TestResult struct {
	TestStatus 			TestStatus 	`json:"testStatus"`
	Message 			string 		`json:"message"`
	ExpiryEpochSeconds 	int64
}