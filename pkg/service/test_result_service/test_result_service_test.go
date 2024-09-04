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

func TestGetExpiryEpochSeconds(t *testing.T) {
	expiryEpochSeconds := GetExpiryEpochSeconds()
	if expiryEpochSeconds < 0 {
		t.Fatalf(`GetExpiryEpochSeconds() = %v; want a positive integer`, expiryEpochSeconds)
	}
}

func TestValidateDeleteFolderPath(t *testing.T) {
	folderName := getFolderName()
	testAreaDirectory := "./tmp/test_area/"
	folderPath := testAreaDirectory + "/" + folderName
	valid := validateDeleteFolderPath(folderPath)
	if !valid {
		t.Fatalf(`validateDeleteFolderPath() = %v; want true`, valid)
	}
}

func TestValidateDeleteFolderPathInvalidIfNoDot(t *testing.T) {
	folderName := getFolderName()
	testAreaDirectory := "/tmp/test_area/"
	folderPath := testAreaDirectory + "/" + folderName
	valid := validateDeleteFolderPath(folderPath)
	if valid {
		t.Fatalf(`validateDeleteFolderPath() = %v; want false`, valid)
	}
}

func TestValidateDeleteFolderPathInvalidIfNoSlashes(t *testing.T) {
	testAreaDirectory := "./tmp"
	folderPath := testAreaDirectory
	valid := validateDeleteFolderPath(folderPath)
	if valid {
		t.Fatalf(`validateDeleteFolderPath() = %v; want false`, valid)
	}
}

