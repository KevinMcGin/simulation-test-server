package test_result_service

import (
	"net/http"
	"os"
	"regexp"
	"testing"
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

func TestValidateDeleteFolderPathInvalidIfDoubleDots(t *testing.T) {
	testAreaDirectory := "../tmp/test_area"
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

func TestValidateCanTestNotIfWrongToken(t *testing.T) {
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

func TestInValidCommitIfIsASpace(t *testing.T) {
	valid := ValidateCommmitId(" ", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsSpace(t *testing.T) {
	valid := ValidateCommmitId("a b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsSemiColon(t *testing.T) {
	valid := ValidateCommmitId("a;b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsTilda(t *testing.T) {
	valid := ValidateCommmitId("a~b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsBackSlash(t *testing.T) {
	valid := ValidateCommmitId("a\\b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsDollarSign(t *testing.T) {
	valid := ValidateCommmitId("a$b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsQuestionMark(t *testing.T) {
	valid := ValidateCommmitId("a?b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsSingleQuote(t *testing.T) {
	valid := ValidateCommmitId("a'b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsDoubleQuotes(t *testing.T) {
	valid := ValidateCommmitId("a\"b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsMinus(t *testing.T) {
	valid := ValidateCommmitId("a-b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}

func TestInValidCommitIfContainsStar(t *testing.T) {
	valid := ValidateCommmitId("a*b", "folderName")
	if valid {
		t.Fatalf(`ValidateCommmitId() = %v; want false`, valid)
	}
}
