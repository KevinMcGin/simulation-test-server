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
	deleteFolderInTestArea("*")

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
		false,
		false,
		"",
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
				true,
				false,
				"Invalid commit id",
				getExpiryEpochSeconds(),
			}
		} else {
			testResult, err = runTestsAndGetResult(folderName)
		}
		if err != nil {
			testResult = TestResult{
				true,
				false,
				err.Error(),
				getExpiryEpochSeconds(),
			}
		}
		resultsMap[testResultId] = testResult
		deleteFolderInTestArea(folderName)
		fmt.Println("Result generated, isSuccess: ", testResult.IsSuccess)
	}()
}

func getTestResultFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !validateCanTest(r) {
		w.WriteHeader(http.StatusUnauthorized)
		testResult := TestResult{
			true,
			false,
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
			true,
			false,
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
		jsonBytes, err := json.Marshal(testResult)
		if err != nil {
			fmt.Println("Error marshalling json:", err.Error())
			w.Write([]byte("error marshalling json"))
			return
		}
		w.Write(jsonBytes)
		if testResult.IsReady {
			delete(resultsMap, testResultId)
			fmt.Println("Deleted retrieved test result: ", testResultId)
		}
	}
	jsonBytes, _ := json.Marshal(testResult)
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
				true,
				false,
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
	}
	if err != nil {
		out, err = exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + "/Simulation && git checkout "+commitId).Output()
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
	out, err := exec.Command("bash", "-c", "cd " + testAreaDirectory + "/" + folderName + " && git clone https://github.com/KevinMcGin/Simulation.git && git fetch").Output()
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
	isSuccess := err == nil
	if !isSuccess {
		fmt.Println("Tests failed:\n ", err, out)
	}
	testMessage := string(out)
	return TestResult{
		true,
		isSuccess,
		testMessage,
		getExpiryEpochSeconds(),
	}, nil
}

func deleteFolderInTestArea(folderName string) {
	out, err := exec.Command("bash", "-c", "rm -rf" + testAreaDirectory + "/" + folderName).Output()
	if err != nil {
		fmt.Println("Error deleting folder(s): ", string(out), err.Error())
	} else {
		fmt.Println("Deleted folder(s): " + testAreaDirectory + "/" + folderName)
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

type TestResult struct {
	IsReady   bool   `json:"isReady"`
	IsSuccess bool   `json:"isSuccess"`
	Message   string `json:"message"`
	expiryEpochSeconds    int64
}
