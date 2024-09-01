package controller


import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"simulation-test-server/pkg/model/test_result"
	"simulation-test-server/pkg/service/test_result_service"
)

var resultsMap map[string]test_result.TestResult = make(map[string]test_result.TestResult)

func HomeFunc(w http.ResponseWriter, r *http.Request) {
	welcomeMessage := "Welcome to the sim test server"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(welcomeMessage))
	fmt.Println(welcomeMessage)
}

func TestFunc(w http.ResponseWriter, r *http.Request) {
	test_result_service.RemoveExpiredResults(resultsMap)
	
	commitId := r.PathValue("commitId")
	testResult := test_result.TestResult{
		TestStatus: test_result.Running,
		Message: "Test running",
		ExpiryEpochSeconds: test_result_service.GetExpiryEpochSeconds(),
	}
	if !test_result_service.ValidateCanTest(r) {
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

		folderName, err := test_result_service.PullDownCodeAndGetFolderName()
		if err != nil {
			// covered by error check below
		} else if !test_result_service.ValidateCommmitId(commitId, folderName) {
			testResult = test_result.TestResult{
				TestStatus: test_result.Errored,
				Message: "Invalid commit id",
				ExpiryEpochSeconds: test_result_service.GetExpiryEpochSeconds(),
			}
		} else {
			testResult, err = test_result_service.RunTestsAndGetResult(folderName)
		}
		if err != nil {
			testResult = test_result.TestResult{
				TestStatus: test_result.Errored,
				Message: err.Error(),
				ExpiryEpochSeconds: test_result_service.GetExpiryEpochSeconds(),
			}
		}
		resultsMap[testResultId] = testResult
		test_result_service.DeleteFolderInTestArea(folderName)
		fmt.Println("Result generated, isSuccess: ", testResult.TestStatus)
	}()
}

func GetTestResultFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !test_result_service.ValidateCanTest(r) {
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