package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"simulation-test-server/pkg/controller"

	"github.com/joho/godotenv"
)


func main() {
	// Load environment variables
	godotenv.Load()
	port := os.Getenv("PORT")
	testAreaDirectory := os.Getenv("TEST_AREA")

	// Create test area directory
	controller.DeleteFolderInTestArea("")
	err := os.Mkdir(testAreaDirectory, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating test area directory: ", err.Error())
	}

	// Define routes
	http.HandleFunc("/api/sim", controller.HomeFunc)
	http.HandleFunc("/api/sim/test/{commitId}/commit", controller.TestFunc)
	http.HandleFunc("/api/sim/test/{testResultId}/result", controller.GetTestResultFunc)

	// Start the server
	fmt.Println("Test server starting on http://127.0.0.1:" + port)
	log.Fatal(http.ListenAndServe("127.0.0.1:" + port, nil))
}




