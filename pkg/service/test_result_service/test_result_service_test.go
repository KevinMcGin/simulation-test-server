package test_result_service

import (
    "testing"
    "regexp"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestGetFolderName(t *testing.T) {
    folderName := getFolderName()
	match, _ := regexp.MatchString("^[0-9a-f]+$", folderName)
	if !match {
		t.Fatalf(`getFolderName() = %v; want a string of 0-9a-f`, folderName)
	}
}