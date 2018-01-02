package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	validHashData = map[string]string{
		"7182575d2f1fe035c0ce8cea70f93cd7": "af7fcc0f-7a58-4a74-bfa2-8fb6e12008eb",
		"86564dc69f8c9b081b5174ef562e1ac1": "8bc11648-d983-4a62-9ea2-590901f374ff",
		"f09058c9de0b55d482f8575f8e8e7628": "48557ec1-3205-403a-b82c-843fd9b03f5b",
		"ca7a7b23a4d5cb655df97378e571f7c6": "cbddd90c-f574-43d9-8d1f-b4989678a09b",
		// Detargeted Files
		"41b26abe9fd63ea31a8ea325bb9fb47c": "48557ec1-3205-403a-b82c-843fd9b03f5b",
		"de5a7197e3e61c7987d3a64731608187": "cbddd90c-f574-43d9-8d1f-b4989678a09b",
		// Pred files
		"3b38668b9e0d1a8931e57d01235de01d": "48557ec1-3205-403a-b82c-843fd9b03f5b",
		"7be08f2103cc61f50eb485dd5c7ef0df": "8bc11648-d983-4a62-9ea2-590901f374ff",
		"a479fb72d25cff24112328433e39915f": "af7fcc0f-7a58-4a74-bfa2-8fb6e12008eb",
		"63f9156ec639f5384c069fe3c7807429": "cbddd90c-f574-43d9-8d1f-b4989678a09b",
	}
	pathFixturesPred = "/fixtures/pred"
)

type Model struct {
	ID        int    `json:"id"`
	Msg       string `json:"msg"`
	Timestamp int    `json:"timestamp"`
}

var (
	pathTrain     string
	pathTrainPred string
	pathTest      string
	pathTestPred  string
	pathModel     string
)

func main() {
	// Parse args
	var task, volume string
	flag.StringVar(&task, "T", "", "Task: train/predict")
	flag.StringVar(&volume, "V", "", "Volume")
	flag.Parse()

	// Check args are properly set
	if (task != "train" && task != "predict") || volume == "" {
		check(fmt.Errorf("task: %s, volume: %s", task, volume), "Missing or invalid arguments")
	}
	log.Printf("Starting task '%s' with volume '%s'...", task, volume)

	// Set up paths
	pathTrain = filepath.Join(volume, "train")
	pathTrainPred = filepath.Join(pathTrain, "pred")
	pathTest = filepath.Join(volume, "test")
	pathTestPred = filepath.Join(pathTest, "pred")
	pathModel = filepath.Join(volume, "model/model_trained.json")

	// Perform training if needed
	if task == "train" {
		train()
		predict(pathTrain, pathTrainPred)
	}

	// Always perform predict test
	predict(pathTest, pathTestPred)
}

func train() {
	// Check data
	files := checkData(pathTrain)
	log.Printf("[train] Starting training with %d data files", len(files))

	// Simulate training
	updateModel()
}

func predict(srcDir, saveDir string) {
	// Check model is here
	if !checkFileExists(pathModel) {
		log.Fatalln("Missing model_trained.json file for predicting task")
	}

	// Read srcDir and check it is properly set
	prevFiles, err := ioutil.ReadDir(srcDir)
	check(err, "")
	if len(prevFiles) == 0 {
		log.Fatalf("[FATAL ERROR] Missing files for predict task in directory %s", srcDir)
	}

	// Check files and copy predict files
	for _, f := range prevFiles {
		if f.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(srcDir, f.Name()))
		check(err, "")
		checksum := fmt.Sprintf("%x", md5.Sum(data))
		fName, ok := validHashData[checksum]
		if !ok {
			log.Fatalf("[FATAL ERROR] Invalid checksum for file %s (%s)", f.Name(), checksum)
		}
		check(os.MkdirAll(saveDir, 0755), fmt.Sprintf("Failed to create directory %s", saveDir))
		check(copyFile(filepath.Join(pathFixturesPred, fName), filepath.Join(saveDir, f.Name())), "[SCRIPT ERROR] Failed to copy predict data")
		log.Printf("[predict] Sucessfully predicted on data %s", f.Name())
	}
}

func updateModel() {
	var model []Model
	newEntry := Model{Timestamp: int(time.Now().Unix()), Msg: "Train"}

	// Get ID
	if checkFileExists(pathModel) {
		data, err := ioutil.ReadFile(pathModel)
		check(err, "")
		check(json.Unmarshal(data, &model), "")
		newEntry.ID = model[len(model)-1].ID + 1
	} else {
		newEntry.ID = 0
	}
	model = append(model, newEntry)

	modelBytes, err := json.Marshal(model)
	check(err, "Failed to Marshal Model")
	check(ioutil.WriteFile(pathModel, modelBytes, 0775), "[SCRIPT ERROR] Failed to WriteFile on Model")
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// checkData reads a dir, and for each file verify that checksums is valid
// it returns the list of files (removing directories)
func checkData(path string) (fileInfos []os.FileInfo) {
	dirFiles, err := ioutil.ReadDir(path)
	check(err, "")
	for _, f := range dirFiles {
		if f.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
		check(err, "")
		checksum := fmt.Sprintf("%x", md5.Sum(data))
		if _, ok := validHashData[checksum]; !ok {
			log.Fatalf("[FATAL ERROR] Invalid checksum for file %s (%s)", f.Name(), checksum)
		}
		fileInfos = append(fileInfos, f)
	}
	if len(fileInfos) == 0 {
		log.Fatalf("[FATAL ERROR] Missing data in folder %s", path)
	}
	return fileInfos
}

func checkFileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func check(err error, msg string) {
	if err != nil {
		if msg == "" {
			log.Fatalln(fmt.Sprintf("%s: %s", "[FATAL ERROR]", err))
		}
		log.Fatalln(fmt.Sprintf("%s %s: %s", "[FATAL ERROR]", msg, err))
	}
}
