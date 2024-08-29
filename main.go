package main

import( 
	"fmt"
	"os"
	"os/exec"
	"time"
	"strconv"
	"encoding/json"
	"log"
	"net/http"
	"github.com/joho/godotenv"
)


var isRunningOnGpu bool = false
var testToken string = "test_token"
var port string = "9000"


func main() {
	fmt.Println("Test server starting on http://127.0.0.1:" + port)

	godotenv.Load()
	testToken = os.Getenv("TEST_TOKEN")
	isRunningOnGpu = os.Getenv("IS_RUNNING_GPU") == "true"
	// Define routes
    http.HandleFunc("/api/sim", homeFunc)
    http.HandleFunc("/api/sim/test/{commitId}/commit", testFunc)

    // Start the server
    log.Fatal(http.ListenAndServe("127.0.0.1:" + port, nil))
}

func homeFunc(w http.ResponseWriter, r *http.Request) {
	welcomeMessage := "Welcome to the sim test server"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(welcomeMessage))
	fmt.Println(welcomeMessage)
}

func testFunc(w http.ResponseWriter, r *http.Request) {
	commitId := r.PathValue("commitId")
	testResult := TestResult {
		false,
		"",
	}
	if !validateCanTest(r) {
		testResult = TestResult {
			false,
			"Authorization failed",
		}
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		fmt.Println("Running tests for commit id:", commitId)

		folderName := pullDownCode()
		if !validateCommmitId(commitId, folderName) {
			testResult = TestResult {
				false,
				"Invalid commit id",
			}
			w.WriteHeader(http.StatusBadRequest)
		} else {
			testResult = runTestsAndGetResult(folderName)	
			w.WriteHeader(http.StatusOK)	
		}
	}
    w.Header().Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(testResult)
	if err != nil {
		fmt.Println("Error marshalling json:", err)
	}
	fmt.Println(testResult)
	w.Write(jsonBytes)
}

func validateCanTest(r *http.Request) bool {
	testerToken := r.Header.Get("Tester-Token")
	return testerToken == testToken
}

func runTestsAndGetResult(folderName string) TestResult {
	if isRunningOnGpu {
		out, err := exec.Command("bash", "-c", "cd test_area/" + folderName + "/Simulation && rm config/project.config.example && mv config/gpu_project.config.example config/project.config.example").Output()
		if err != nil {
			fmt.Println("Error renaming project.config: ", out, err)
		}
	}

	testResult := runTests(folderName)
	deleteFolder(folderName)
	return testResult
}


func validateCommmitId(commitId string, folderName string) bool {
	out, err := exec.Command("bash", "-c", "cd test_area/" + folderName + "/Simulation && git cat-file commit " + commitId).Output()
	if err != nil {
		fmt.Println("Error validating commit: ", commitId, " ", out, err)
		fmt.Println(err)
	}		
	if err != nil {
		out, err = exec.Command("bash", "-c", "cd test_area/" + folderName + "/Simulation && git checkout " + commitId).Output()
		if err != nil {
			fmt.Println("Error checking out commit: ", out, err)
		}
	}
	return err == nil
}

func pullDownCode() string {
	// Todo: get the timestamp
	folderName := strconv.FormatInt(time.Now().Unix(), 10)
	err := os.Mkdir("test_area/" + folderName, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder: ", err)
		fmt.Println(err)
	}
	out, err := exec.Command("bash", "-c", "cd test_area/" + folderName + " && git clone https://github.com/KevinMcGin/Simulation.git").Output()
	if err != nil {
		fmt.Println("Error cloning repo: ", out, err)
	}
	return folderName
}

func runTests(folderName string) TestResult {
	out, err := exec.Command("bash", "-c", "cd test_area/" + folderName + "/Simulation/scripts && ./test.sh").Output()
	isSuccess := err == nil
	if !isSuccess {
		fmt.Println("Tests failed:\n ", err, out)
	}
	testMessage := string(out)
	return TestResult {
		isSuccess,
		testMessage,
	}
}

func deleteFolder(folderName string) {
	out, err := exec.Command("bash", "-c", "rm -rf test_area/" + folderName).Output()
	if err != nil {
		fmt.Println("Error deleting folder: ", out, err)
	}
}

type TestResult struct {
    IsSuccess bool `json:"isSuccess"`
    Message string `json:"message"`
}