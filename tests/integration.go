package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/MorpheoOrg/morpheo-go-packages/client"
	"github.com/MorpheoOrg/morpheo-go-packages/common"
)

var (
	pathFixturesYAML = "/go/src/github.com/MorpheoOrg/morpheo-devenv/tests/fixtures/metadata.yaml"
	pathPeerConfig   = "/secrets/config.yaml"

	storage = &client.StorageAPI{
		Hostname: "storage",
		Port:     80,
		User:     "u",
		Password: "p",
	}
	compute = &client.ComputeAPI{
		Hostname: "compute",
		Port:     80,
		// User:     "u",
		// Password: "p",
	}
	peer *client.PeerAPI
	err  error
)

func main() {
	log.Println("Integration Tests Starting!")

	// Connecting to the peer client
	peer, err = client.NewPeerAPI(pathPeerConfig, "Aphp", "mychannel", "mycc")
	check(err, "[peer-API] Failed to create peerAPI")

	testLearnPred()

	log.Println("GREAT SUCCESS!")
}

// testLearnPred tests learning and prediction on the devenv
func testLearnPred() {
	// Load the fixtures
	fixtures, err := common.ParseDataFromFile(pathFixturesYAML)
	check(err, "Error loading Fixtures")

	// Post the fixtures to Storage
	check(postFixturesStorage(fixtures), "Error posting Fixtures")

	// Post the fixtures to the Chaincode
	check(registerFixturesChaincode(fixtures), "[Chaincode] Error posting Fixtures")

	// Wait for the first pending learnuplet
	var pendingList []string
	for {
		time.Sleep(4 * time.Second)
		pendingList, err = getPendingLearnupletList("pending")
		check(err, "[peer-api] Error getting pending learnuplets")

		log.Printf("[peer-api] %d learnuplet(s) with status \"pending\" detected", len(pendingList))
		if len(pendingList) > 0 {
			break
		}
	}

	// Wait for the learnuplet done status
	pendingKey := pendingList[0]
	for {
		// Get learnuplet Status
		var learnuplet common.LearnupletChaincode
		learnupletBytes, err := peer.Query("queryItem", []string{pendingKey})
		check(err, fmt.Sprintf("[peer-api] Error queryItem %s", pendingKey))

		check(json.Unmarshal(learnupletBytes, &learnuplet), "Error Unmarshalling pending learnuplet")

		if learnuplet.Status == "failed" {
			check(fmt.Errorf("Error in the worker learning task"), "[integration-tests] Learnuplet status is failed")
		}
		if learnuplet.Status == "done" {
			break
		}
		log.Printf("[learn] Waiting for learnuplet status \"done\". Last status: %s. Checking again in 20s...", learnuplet.Status)
		time.Sleep(20 * time.Second)
	}
	log.Println("[learn] SUCCESSFUL! Learnuplet status is DONE.")

	// // Request prediction to Chaincode
	// log.Println("[pred][Chaincode] Posting prediction request")
	// check(requestPredictionsChaincode(fixtures.Chaincode.Prediction), "[Chaincode] Error posting prediction to Chaincode")

	// // Wait for prediction to complete
	// time.Sleep(2 * time.Second)
	// for {
	// 	status, err := getLastPredupletStatus()
	// 	check(err, "[pred][getUpletStatus] Error getting status")

	// 	if status == "done" {
	// 		break
	// 	}
	// 	if status != "pending" {
	// 		log.Fatalln("[learn] Error: Learnuplet status is %s, whereas it should be 'pending' or 'done'.\n", status)
	// 	}
	// 	log.Printf("[pred] Waiting for preduplet status DONE on Chaincode. Last status: %s. Checking again in 20s...", status)
	// 	time.Sleep(20 * time.Second)
	// }
	// log.Println("[pred] SUCCESSFUL! Preduplet status is DONE.")
	// log.Println("SUCESSFULLY LEARNED AND PREDICTED!")
}

// ================================================================
// Storage functions
// ================================================================

func postFixturesStorage(fixtures *common.DataParser) error {

	// Post Problems
	for _, resource := range fixtures.Storage.Problem {
		log.Printf("[storage] Posting problem/%s...", resource.ID)
		file, err := fixtures.GetData("problem", resource.ID.String())
		if err != nil {
			log.Printf("[storage]%s", err)
			continue
		}
		if err := storage.PostProblem(resource, 666, file); err != nil {
			if !resourceAlreadyExist(err) {
				return err
			}
			log.Printf("[storage] problem/%s already exists", resource.ID)
		}
	}
	// Post Data
	for _, resource := range fixtures.Storage.Data {
		log.Printf("[storage] Posting data/%s...", resource.ID)
		file, err := fixtures.GetData("data", resource.ID.String())
		if err != nil {
			log.Printf("[storage]%s", err)
			continue
		}
		if err := storage.PostData(resource, 666, file); err != nil {
			if !resourceAlreadyExist(err) {
				return err
			}
			log.Printf("[storage] data/%s already exists", resource.ID)
		}
	}

	// Post Algo
	for _, resource := range fixtures.Storage.Algo {
		log.Printf("[storage] Posting algo/%s...", resource.ID)
		file, err := fixtures.GetData("algo", resource.ID.String())
		if err != nil {
			log.Printf("[storage]%s", err)
			continue
		}
		if err := storage.PostAlgo(resource, 666, file); err != nil {
			if !resourceAlreadyExist(err) {
				return err
			}
			log.Printf("[storage] algo/%s already exists", resource.ID)
		}
	}

	// // Post Model
	// for _, resource := range fixtures.Storage.Model {
	// 	log.Printf("[storage] POST Model %s", resource.ID)
	// 	file, err := fixtures.GetData("model", resource.ID.String())
	// 	if err != nil {
	// 		log.Printf("[storage]%s", err)
	//		continue
	// 	}
	// 	if err := storage.PostModel(&resource, file, 3933); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// ================================================================
// Chaincode functions
// ================================================================

func registerFixturesChaincode(fixtures *common.DataParser) error {
	// Register Problem
	for _, resource := range fixtures.Chaincode.Problem {
		log.Printf("[peer-API] Registering problem %s...", resource.StorageAddress)
		_, _, err := peer.RegisterProblem(resource.StorageAddress, resource.SizeTrainDataset, resource.TestData)
		if err != nil {
			return fmt.Errorf("[peer-API] Error registering problem %s: %s", resource.StorageAddress, err)
		}
	}

	// Register Data
	for _, resource := range fixtures.Chaincode.Data {
		log.Printf("[peer-API] Registering data %s...", resource.StorageAddress)
		_, _, err := peer.RegisterItem("data", resource.StorageAddress, resource.ProblemKeys, resource.Name)
		if err != nil {
			return fmt.Errorf("[peer-API] Error registering data %s: %s", resource.StorageAddress, err)
		}
	}

	// Register Algo
	for _, resource := range fixtures.Chaincode.Algo {
		log.Printf("[peer-API] Registering algo %s...", resource.StorageAddress)
		_, _, err := peer.RegisterItem("algo", resource.StorageAddress, resource.ProblemKeys, resource.Name)
		if err != nil {
			return fmt.Errorf("[peer-API] Error registering algo %s: %s", resource.StorageAddress, err)
		}
	}

	return nil
}

func getPendingLearnupletList(status string) (pendingList []string, err error) {
	learnupletsByte, err := peer.QueryStatusLearnuplet("pending")
	if err != nil {
		return nil, fmt.Errorf("[peer-API] Error getting pending learnpulets: %s", err)
	}
	var learnuplets []common.LearnupletChaincode
	err = json.Unmarshal(learnupletsByte, &learnuplets)
	if err != nil {
		return nil, fmt.Errorf("[peer-API] Error Unmarshal-ing pendind learnuplets: %s", err)
	}
	for _, learnuplet := range learnuplets {
		pendingList = append(pendingList, learnuplet.Key)
	}
	return pendingList, nil
}

func getLastPredupletStatus() (string, error) {
	// body, err := Chaincode.GetList("preduplet")
	// if err != nil {
	// 	return "", fmt.Errorf("Error retrieving preduplet list: %s", err)
	// }

	// // Unmarshal the learnuplet
	// var preduplets map[string][]common.Preduplet
	// err = json.NewDecoder(bytes.NewReader(body)).Decode(&preduplets)
	// if err != nil {
	// 	return "", fmt.Errorf("Error un-marshaling learnuplet: %s. Body: %s", err, string(body))
	// }
	// return preduplets["preduplets"][0].Status, nil
	return "", nil
}

func getLastLearnupletStatus() (string, error) {
	// body, err := Chaincode.GetList("learnuplet")
	// if err != nil {
	// 	return "", fmt.Errorf("Error retrieving learnuplet list: %s", err)
	// }

	// // Unmarshal the learnuplet
	// var learnuplets map[string][]common.Learnuplet
	// err = json.NewDecoder(bytes.NewReader(body)).Decode(&learnuplets)
	// if err != nil {
	// 	return "", fmt.Errorf("Error un-marshaling learnuplet: %s. Body: %s", err, string(body))
	// }
	// return learnuplets["learnuplets"][0].Status, nil
	return "", nil
}

// ============================================
// Utils
// ============================================

func check(err error, msg string) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s%s: %s\n", "[FATAL ERROR]", msg, err))
	}
}

func resourceAlreadyExist(err error) bool {
	re := regexp.MustCompile("409 Conflict")
	return re.MatchString(err.Error())
}

// ================================================================
// Client Tests
// ================================================================

func testReportLearnFailed(learnuplet string) {
	var m map[string]float64
	var f float64
	_, _, err := peer.ReportLearn(learnuplet, common.TaskStatusFailed, f, m, m)
	if err != nil {
		log.Fatalf("[FATAL ERROR] Error in testReportLearn: %s", err)
	}
}

func testReportLearnDone(learnuplet string) {
	m := map[string]float64{"p": 0.5}
	f := 0.5
	_, _, err := peer.ReportLearn(learnuplet, common.TaskStatusDone, f, m, m)
	if err != nil {
		log.Fatalf("[FATAL ERROR] Error in testReportLearn: %s", err)
	}
}
