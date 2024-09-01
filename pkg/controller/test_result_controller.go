package controller


import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
	"errors"
	"strings"

	"simulation-test-server/pkg/model/test_result"
)

var resultsMap map[string]test_result.TestResult = make(map[string]test_result.TestResult)

func HomeFunc(w http.ResponseWriter, r *http.Request) {
	welcomeMessage := "Welcome to the sim test server"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(welcomeMessage))
	fmt.Println(welcomeMessage)
}

func TestFunc(w http.ResponseWriter, r *http.Request) {
	removeExpiredResults()
	
	commitId := r.PathValue("commitId")
	testResult := test_result.TestResult{
		TestStatus: test_result.Running,
		Message: "Test running",
		ExpiryEpochSeconds: getExpiryEpochSeconds(),
	}
	if !validateCanTest(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("error: invalid token"))
		return
	}
	testResultId := strconv.FormatInt(time.Now().UnixMilli(), 16)
	resultsMap[testResultId] = testResult
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(testResultId))
	go func() {
		fmt.Println("Running tests for commit id:", commitId)

		folderName, err := pullDownCodeAndGetFolderName()
		if err != nil {
			// covered by error check below
		} else if !validateCommmitId(commitId, folderName) {
			testResult = test_result.TestResult{
				TestStatus: test_result.Errored,
				Message: "Invalid commit id",
				ExpiryEpochSeconds: getExpiryEpochSeconds(),
			}
		} else {
			testResult, err = runTestsAndGetResult(folderName)
		}
		if err != nil {
			testResult = test_result.TestResult{
				TestStatus: test_result.Errored,
				Message: err.Error(),
				ExpiryEpochSeconds: getExpiryEpochSeconds(),
			}
		}
		resultsMap[testResultId] = testResult
		DeleteFolderInTestArea(folderName)
		fmt.Println("Result generated, isSuccess: ", testResult.TestStatus)
	}()
}

func GetTestResultFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !validateCanTest(r) {
		w.WriteHeader(http.StatusUnauthorized)
		testResult := test_result.TestResult{
			TestStatus: test_result.Errored,
			Message: "Invalid token",
			ExpiryEpochSeconds: 0,
		}
		jsonBytes, err := json.Marshal(testResult)
		if err != nil {
			fmt.Println("error marshalling json:", err.Error())
			w.Write([]byte("error marshalling json"))
			return
		}
		w.Write(jsonBytes)
		return
	}
	testResultId := r.PathValue("testResultId")
	testResult, ok := resultsMap[testResultId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		testResult = test_result.TestResult{
			TestStatus: test_result.Errored,
			Message: "Test result not found",
			ExpiryEpochSeconds: 0,
		}
		jsonBytes, err := json.Marshal(testResult)
		if err != nil {
			fmt.Println("Error marshalling json:", err.Error())
			w.Write([]byte("error marshalling json"))
			return
		}
		w.Write(jsonBytes)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		if testResult.TestStatus != test_result.Running {
			delete(resultsMap, testResultId)
			fmt.Println("Deleted retrieved test result: ", testResultId)
		}
	}
	jsonBytes, err := json.Marshal(testResult)
	if err != nil {
		fmt.Println("Error marshalling json:", err.Error())
		w.Write([]byte("error marshalling json"))
		return
	}
	w.Write(jsonBytes)
}

func validateCanTest(r *http.Request) bool {
	testerToken := r.Header.Get("Tester-Token")
	testToken := os.Getenv("TEST_TOKEN")
	return testerToken == testToken
}

func runTestsAndGetResult(folderName string) (test_result.TestResult, error) {
	isRunningOnGpu := os.Getenv("IS_RUNNING_GPU") == "true"
	testAreaDirectory := os.Getenv("TEST_AREA")
	if isRunningOnGpu {
		out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation && rm config/project.config.example && mv config/gpu_project.config.example config/project.config.example").Output()
		if err != nil {
			fmt.Println("Error renaming project.config: ", string(out), err.Error())
			return test_result.TestResult{
				TestStatus: test_result.Errored,
				Message: "Error renaming project.config",
				ExpiryEpochSeconds: getExpiryEpochSeconds(),
			}, errors.New("error renaming project.config: " + err.Error())
		}
	}

	testResult, err := runTests(folderName)
	return testResult, err
}

func validateCommmitId(commitId string, folderName string) bool {
	testAreaDirectory := os.Getenv("TEST_AREA")
	out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation && git cat-file commit " + commitId).Output()
	if err != nil {
		fmt.Println("Error validating commit: ", commitId, " ", string(out), err.Error())
	} else {
		out, err = exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation && git checkout " + commitId).Output()
		if err != nil {
			fmt.Println("Error checking out commit: ", string(out), err.Error())
		}
	}
	return err == nil
}

func pullDownCodeAndGetFolderName() (string, error) {
	testAreaDirectory := os.Getenv("TEST_AREA")
	folderName := getFolderName() 
	err := os.Mkdir(testAreaDirectory + "/" + folderName, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder: ", err.Error())
		return "", errors.New("error creating folder: " + err.Error())
	}
	out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + " && git clone https://github.com/KevinMcGin/Simulation.git && cd Simulation && git fetch").Output()
	if err != nil {
		fmt.Println("Error cloning repo: ", string(out), err.Error())
		return "", errors.New("error cloning repo: " + err.Error())
	}
	return folderName, nil
}

func getFolderName() string {
	return strconv.FormatInt(time.Now().UnixMicro(), 16)
}

func runTests(folderName string) (test_result.TestResult, error) {
	testAreaDirectory := os.Getenv("TEST_AREA")
	out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation/scripts && ./test.sh").Output()
	var testStatus test_result.TestStatus = test_result.Success 
	if err != nil {
		fmt.Println("Tests failed:\n ", err, out)
		testStatus = test_result.Failure
	}
	testMessage := string(out)
	return test_result.TestResult{
		TestStatus: testStatus,
		Message: testMessage,
		ExpiryEpochSeconds: getExpiryEpochSeconds(),
	}, nil
}

func validateDeleteFolderPath(folderPath string) bool {
	res := strings.Split(folderPath, "/")
	return len(res) >= 2 && 
		res[0] == "." &&
		!strings.Contains(folderPath, "..")
}

func DeleteFolderInTestArea(folderName string) {
	testAreaDirectory := os.Getenv("TEST_AREA")
	folderPath := testAreaDirectory + "/" + folderName
	if !validateDeleteFolderPath(folderPath) {
		fmt.Println("Invalid folder path: ", folderPath)
		return
	}
	err := os.RemoveAll(folderPath)
	if err != nil {
		fmt.Println("Error deleting folder(s): " + folderPath, err.Error())
	} else {
		fmt.Println("Deleted folder(s): " + folderPath)
	}
}

func getExpiryEpochSeconds() int64 {
	return time.Now().Add(2 * time.Hour).Unix()
}

func removeExpiredResults() {
	for key, value := range resultsMap {
		if value.ExpiryEpochSeconds < time.Now().Unix() {
			delete(resultsMap, key)
			fmt.Println("Deleted expired test result: ", key)
		}
	}
}