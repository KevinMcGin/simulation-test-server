package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
	"errors"
	"strings"

	"github.com/joho/godotenv"
)

var isRunningOnGpu bool 
var testToken string
var port string
var testAreaDirectory string

var resultsMap map[string]TestResult = make(map[string]TestResult)

func main() {
	godotenv.Load()
	testToken = os.Getenv("TEST_TOKEN")
	isRunningOnGpu = os.Getenv("IS_RUNNING_GPU") == "true"
	port = os.Getenv("PORT")
	testAreaDirectory = os.Getenv("TEST_AREA")

	fmt.Println("Test server starting on http://127.0.0.1:" + port)
	deleteFolderInTestArea("")
	err := os.Mkdir(testAreaDirectory, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating test area directory: ", err.Error())
	}

	// Define routes
	http.HandleFunc("/api/sim", homeFunc)
	http.HandleFunc("/api/sim/test/{commitId}/commit", testFunc)
	http.HandleFunc("/api/sim/test/{testResultId}/result", getTestResultFunc)

	// Start the server
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}

func homeFunc(w http.ResponseWriter, r *http.Request) {
	welcomeMessage := "Welcome to the sim test server"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(welcomeMessage))
	fmt.Println(welcomeMessage)
}

func testFunc(w http.ResponseWriter, r *http.Request) {
	removeExpiredResults()
	
	commitId := r.PathValue("commitId")
	testResult := TestResult{
		"RUNNING",
		"Test running",
		getExpiryEpochSeconds(),
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
			testResult = TestResult{
				"ERRORED",
				"Invalid commit id",
				getExpiryEpochSeconds(),
			}
		} else {
			testResult, err = runTestsAndGetResult(folderName)
		}
		if err != nil {
			testResult = TestResult{
				"ERRORED",
				err.Error(),
				getExpiryEpochSeconds(),
			}
		}
		resultsMap[testResultId] = testResult
		deleteFolderInTestArea(folderName)
		fmt.Println("Result generated, isSuccess: ", testResult.TestStatus)
	}()
}

func getTestResultFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !validateCanTest(r) {
		w.WriteHeader(http.StatusUnauthorized)
		testResult := TestResult{
			"ERRORED",
			"Invalid token",
			0,
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
		testResult = TestResult{
			"ERRORED",
			"Test result not found",
			0,
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
		if testResult.TestStatus != "RUNNING" {
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
	return testerToken == testToken
}

func runTestsAndGetResult(folderName string) (TestResult, error) {
	if isRunningOnGpu {
		out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation && rm config/project.config.example && mv config/gpu_project.config.example config/project.config.example").Output()
		if err != nil {
			fmt.Println("Error renaming project.config: ", string(out), err.Error())
			return TestResult{
				"ERRORED",
				"Error renaming project.config",
				getExpiryEpochSeconds(),
			}, errors.New("error renaming project.config: " + err.Error())
		}
	}

	testResult, err := runTests(folderName)
	return testResult, err
}

func validateCommmitId(commitId string, folderName string) bool {
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
	// Todo: get the timestamp
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

func runTests(folderName string) (TestResult, error) {
	out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation/scripts && ./test.sh").Output()
	var testStatus TestStatus = "SUCCESS" 
	if err != nil {
		fmt.Println("Tests failed:\n ", err, out)
		testStatus = "FAILURE"
	}
	testMessage := string(out)
	return TestResult{
		testStatus,
		testMessage,
		getExpiryEpochSeconds(),
	}, nil
}

func validateDeleteFolderPath(folderPath string) bool {
	res := strings.Split(folderPath, "/")
	return len(res) >= 2 && 
		res[0] == "." &&
		!strings.Contains(folderPath, "..")
}

func deleteFolderInTestArea(folderName string) {
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
		if value.expiryEpochSeconds < time.Now().Unix() {
			delete(resultsMap, key)
			fmt.Println("Deleted expired test result: ", key)
		}
	}
}

type TestStatus string

const (
    running TestStatus = "RUNNING"
    success TestStatus = "SUCCESS"
    failure TestStatus = "FAILURE"
    errored TestStatus = "ERRORED"
)

type TestResult struct {
	TestStatus 			TestStatus 	`json:"testStatus"`
	Message 			string 		`json:"message"`
	expiryEpochSeconds 	int64
}
