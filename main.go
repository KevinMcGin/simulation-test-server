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

	"github.com/joho/godotenv"
)

var isRunningOnGpu bool = false
var testToken string = "test_token"
var port string = "9000"

var resultsMap map[string]TestResult = make(map[string]TestResult)

func main() {
	fmt.Println("Test server starting on http://127.0.0.1:" + port)

	godotenv.Load()
	testToken = os.Getenv("TEST_TOKEN")
	isRunningOnGpu = os.Getenv("IS_RUNNING_GPU") == "true"
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
		w.Write([]byte("Invalid token"))
		return
	}
	testResultId := strconv.FormatInt(time.Now().UnixMilli(), 16)
	resultsMap[testResultId] = testResult
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(testResultId))
	go func() {
		fmt.Println("Running tests for commit id:", commitId)

		folderName := pullDownCode()
		if !validateCommmitId(commitId, folderName) {
			testResult = TestResult{
				true,
				false,
				"Invalid commit id",
				getExpiryEpochSeconds(),
			}
		} else {
			testResult = runTestsAndGetResult(folderName)
		}
		resultsMap[testResultId] = testResult
		fmt.Println(testResult)
	}()
}

func getTestResultFunc(w http.ResponseWriter, r *http.Request) {
	testResultId := r.PathValue("testResultId")
	testResult, ok := resultsMap[testResultId]
	// if test result is nil
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Test result not found"))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonBytes, err := json.Marshal(testResult)
		if err != nil {
			fmt.Println("Error marshalling json:", err)
		}
		w.Write(jsonBytes)
		if testResult.IsReady {
			delete(resultsMap, testResultId)
			fmt.Println("Deleted retrieved test result: ", testResultId)
		}
	}
}

func validateCanTest(r *http.Request) bool {
	testerToken := r.Header.Get("Tester-Token")
	return testerToken == testToken
}

func runTestsAndGetResult(folderName string) TestResult {
	if isRunningOnGpu {
		out, err := exec.Command("bash", "-c", "cd test_area/"+folderName+"/Simulation && rm config/project.config.example && mv config/gpu_project.config.example config/project.config.example").Output()
		if err != nil {
			fmt.Println("Error renaming project.config: ", out, err)
		}
	}

	testResult := runTests(folderName)
	deleteFolder(folderName)
	return testResult
}

func validateCommmitId(commitId string, folderName string) bool {
	out, err := exec.Command("bash", "-c", "cd test_area/"+folderName+"/Simulation && git cat-file commit "+commitId).Output()
	if err != nil {
		fmt.Println("Error validating commit: ", commitId, " ", out, err)
		fmt.Println(err)
	}
	if err != nil {
		out, err = exec.Command("bash", "-c", "cd test_area/"+folderName+"/Simulation && git checkout "+commitId).Output()
		if err != nil {
			fmt.Println("Error checking out commit: ", out, err)
		}
	}
	return err == nil
}

func pullDownCode() string {
	// Todo: get the timestamp
	folderName := strconv.FormatInt(time.Now().Unix(), 10)
	err := os.Mkdir("test_area/"+folderName, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder: ", err)
		fmt.Println(err)
	}
	out, err := exec.Command("bash", "-c", "cd test_area/"+folderName+" && git clone https://github.com/KevinMcGin/Simulation.git").Output()
	if err != nil {
		fmt.Println("Error cloning repo: ", out, err)
	}
	return folderName
}

func runTests(folderName string) TestResult {
	out, err := exec.Command("bash", "-c", "cd test_area/"+folderName+"/Simulation/scripts && ./test.sh").Output()
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
	}
}

func deleteFolder(folderName string) {
	out, err := exec.Command("bash", "-c", "rm -rf test_area/"+folderName).Output()
	if err != nil {
		fmt.Println("Error deleting folder: ", out, err)
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
