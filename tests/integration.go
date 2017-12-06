package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/MorpheoOrg/morpheo-go-packages/client"
	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

var (
	// PathResourceData is the path where the fixtures metadata are stored in a .yaml
	defaultPathMetadata = path.Join(os.Getenv("GOPATH"), "src/github.com/MorpheoOrg/morpheo-devenv/tests/fixtures.yaml")
	defaultPathData     = path.Join(os.Getenv("GOPATH"), "src/github.com/MorpheoOrg/morpheo-devenv/data/fixtures")

	pathMetadata = getenv("PATH_METADATA", defaultPathMetadata)
	pathData     = getenv("PATH_DATA", defaultPathData)

	orchestrator = &client.OrchestratorAPI{
		Hostname: "0.0.0.0",
		Port:     8083,
		User:     "u",
		Password: "p",
	}
	storage = &client.StorageAPI{
		Hostname: "0.0.0.0",
		Port:     8081,
		User:     "u",
		Password: "p",
	}
	compute = &client.ComputeAPI{
		Hostname: "0.0.0.0",
		Port:     8082,
		// User:     "u",
		// Password: "p",
	}
)

func main() {
	log.Println("Integration Tests Starting!")
	testLearnPred()
	log.Println("GREAT SUCCESS!")
}

// testLearnPred tests learning and prediction on the devenv
func testLearnPred() {
	// Load the fixtures
	fixtures, err := common.NewDataParser(pathMetadata, pathData)
	check(err, "[fixtures] Error loading Fixtures")

	// Post the fixtures to Storage
	check(postFixturesStorage(fixtures), "[storage] Error posting Fixtures")

	// Post the fixtures to the Orchestrator
	check(postFixturesOrchestrator(fixtures), "[orchestrator] Error posting Fixtures")

	// Wait for learning to complete
	time.Sleep(2 * time.Second)
	for {
		status, err := getLastLearnupletStatus()
		check(err, "[learn][getUpletStatus] Error getting status")

		if status == "done" {
			break
		}
		if status != "pending" {
			log.Fatalln("[learn] Error: Learnuplet status is %s, whereas it should be 'pending' or 'done'.\n", status)
		}
		log.Printf("[learn] Waiting for learnuplet status DONE on Orchestrator. Last status: %s. Checking again in 20s...", status)
		time.Sleep(20 * time.Second)
	}
	log.Println("[learn] SUCCESSFUL! Learnuplet status is DONE.")

	// Request prediction to Orchestrator
	log.Println("[pred][orchestrator] Posting prediction request")
	check(requestPredictionsOrchestrator(fixtures.Orchestrator.Prediction), "[orchestrator] Error posting prediction to Orchestrator")

	// Wait for prediction to complete
	time.Sleep(2 * time.Second)
	for {
		status, err := getLastPredupletStatus()
		check(err, "[pred][getUpletStatus] Error getting status")

		if status == "done" {
			break
		}
		if status != "pending" {
			log.Fatalln("[learn] Error: Learnuplet status is %s, whereas it should be 'pending' or 'done'.\n", status)
		}
		log.Printf("[pred] Waiting for preduplet status DONE on Orchestrator. Last status: %s. Checking again in 20s...", status)
		time.Sleep(20 * time.Second)
	}
	log.Println("[pred] SUCCESSFUL! Preduplet status is DONE.")
	log.Println("SUCESSFULLY LEARNED AND PREDICTED!")
}

func postFixturesOrchestrator(fixtures *common.DataParser) error {

	// Post Problem
	for _, resource := range fixtures.Orchestrator.Problem {
		log.Printf("[orchestrator] POST problem/%s", resource.ID)
		if err := orchestrator.PostProblem(resource); err != nil {
			return fmt.Errorf("[orchestrator] Error posting Problem %s: %s", resource, err)
		}
	}
	// Post Data
	for _, resource := range fixtures.Orchestrator.Data {
		log.Printf("[orchestrator] POST data/%s", resource.ID)
		if err := orchestrator.PostData(resource); err != nil {
			return fmt.Errorf("[orchestrator] Error posting Data %s: %s", resource, err)
		}
	}
	// Post Algo
	for _, resource := range fixtures.Orchestrator.Algo {
		log.Printf("[orchestrator] POST algo/%s", resource.ID)
		if err := orchestrator.PostAlgo(resource); err != nil {
			return fmt.Errorf("[orchestrator] Error posting Algo %s: %s", resource, err)
		}
	}

	return nil
}

func postFixturesStorage(fixtures *common.DataParser) error {

	// Post Problems
	for _, resource := range fixtures.Storage.Problem {
		log.Printf("[storage] POST problem/%s", resource.ID)
		file, err := fixtures.GetData("problem", resource.ID.String())
		if err != nil {
			log.Println(err)
		}
		if err := storage.PostProblem(resource, 666, file); err != nil {
			return err
		}
	}
	// Post Data
	for _, resource := range fixtures.Storage.Data {
		log.Printf("[storage] POST data/%s", resource.ID)
		file, err := fixtures.GetData("data", resource.ID.String())
		if err != nil {
			log.Println(err)
		}
		if err := storage.PostData(resource, 666, file); err != nil {
			return err
		}
	}

	// Post Algo
	for _, resource := range fixtures.Storage.Algo {
		log.Printf("[storage] POST algo/%s", resource.ID)
		file, err := fixtures.GetData("algo", resource.ID.String())
		if err != nil {
			log.Println(err)
		}
		if err := storage.PostAlgo(resource, 666, file); err != nil {
			return err
		}
	}

	// // Post Model
	// for _, resource := range fixtures.Storage.Model {
	// 	log.Printf("[storage] POST Model %s", resource.ID)
	// 	file, err := fixtures.GetData("model", resource.ID.String())
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	if err := storage.PostModel(&resource, file, 3933); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func requestPredictionsOrchestrator(predictions []common.OrchestratorPrediction) error {
	for _, resource := range predictions {
		if err := orchestrator.PostPrediction(resource); err != nil {
			return fmt.Errorf("Error posting Prediction %s: %s", resource, err)
		}
	}
	return nil
}

func getLastPredupletStatus() (string, error) {
	body, err := orchestrator.GetList("preduplet")
	if err != nil {
		return "", fmt.Errorf("Error retrieving preduplet list: %s", err)
	}

	// Unmarshal the learnuplet
	var preduplets map[string][]common.Preduplet
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&preduplets)
	if err != nil {
		return "", fmt.Errorf("Error un-marshaling learnuplet: %s. Body: %s", err, string(body))
	}
	return preduplets["preduplets"][0].Status, nil
}

func getLastLearnupletStatus() (string, error) {
	body, err := orchestrator.GetList("learnuplet")
	if err != nil {
		return "", fmt.Errorf("Error retrieving learnuplet list: %s", err)
	}

	// Unmarshal the learnuplet
	var learnuplets map[string][]common.Learnuplet
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&learnuplets)
	if err != nil {
		return "", fmt.Errorf("Error un-marshaling learnuplet: %s. Body: %s", err, string(body))
	}
	return learnuplets["learnuplets"][0].Status, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func check(err error, msg string) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s%s: %s\n", "[FATAL ERROR]", msg, err))
	}
}
