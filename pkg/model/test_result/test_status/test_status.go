package test_status

type TestStatus string

const (
    Running TestStatus = "RUNNING"
    Success TestStatus = "SUCCESS"
    Failure TestStatus = "FAILURE"
    Errored TestStatus = "ERRORED"
)