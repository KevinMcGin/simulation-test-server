package test_result_service

import (
    "testing"
    "regexp"
	"os"
	"net/http"
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

func TestValidateCanTest(t *testing.T) {
	os.Setenv("TEST_TOKEN", "test")
	r := &http.Request{
		Header: http.Header{
			"Tester-Token": []string{"test"},
		},
	}
	valid := ValidateCanTest(r)
	if !valid {
		t.Fatalf(`ValidateCanTest() = %v; want true`, valid)
	}
}

func TestValidateCanTestNotIfWrontToken(t *testing.T) {
	os.Setenv("TEST_TOKEN", "not-test")
	r := &http.Request{
		Header: http.Header{
			"Tester-Token": []string{"test"},
		},
	}
	valid := ValidateCanTest(r)
	if valid {
		t.Fatalf(`ValidateCanTest() = %v; want false`, valid)
	}
}

