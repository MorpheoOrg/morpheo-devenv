package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
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
	pathFixturesUntargeted = "/fixtures/untargetedTest"
	pathPerf               string
)

// Perfuplet describes the performance.json file, an output of learning tasks
type Perfuplet struct {
	Perf      float64            `json:"perf"`
	TrainPerf map[string]float64 `json:"train_perf"`
	TestPerf  map[string]float64 `json:"test_perf"`
}

func main() {
	// Parse args -T detarget/perf -i /hidden_path -s /submission_path
	var task, hiddenPath, submissionPath string
	flag.StringVar(&task, "T", "", "task: detarget/perf")
	flag.StringVar(&hiddenPath, "i", "", "hidden_path")
	flag.StringVar(&submissionPath, "s", "", "submission_path")
	flag.Parse()

	// Check args are properly set
	if (task != "detarget" && task != "perf") || hiddenPath == "" || submissionPath == "" {
		log.Fatalf("Missing or invalid arguments: -T: %s, -i: %s, -s: %s", task, hiddenPath, submissionPath)
	}
	log.Printf("Starting task '%s' with hidden_path '%s' and submission_path '%s'...", task, hiddenPath, submissionPath)

	// Setup directory structures
	dir_true_test_files := filepath.Join(hiddenPath, "/test/")
	dir_detargeted_test_files := filepath.Join(submissionPath, "/test/")
	dir_true_train_files := filepath.Join(submissionPath, "/train/")

	dir_pred_test_files := filepath.Join(submissionPath, "/test/pred")
	dir_pred_train_files := filepath.Join(submissionPath, "/train/pred")

	dirPerf := filepath.Join(hiddenPath, "/perf")
	filePerf := "performance.json"

	// Detarget
	if task == "detarget" {
		detarget(dir_true_test_files, dir_detargeted_test_files)
	}
	if task == "perf" {
		perf(dir_true_test_files, dir_true_train_files,
			dir_pred_test_files, dir_pred_train_files,
			dirPerf, filePerf)
	}
}

func detarget(prevDir, newDir string) {
	log.Printf("Removing targets from %s into %s...", prevDir, newDir)

	// Read prevDir and check it is properly set
	prevFiles, err := ioutil.ReadDir(prevDir)
	check(err, "")
	if len(prevFiles) == 0 {
		log.Fatalf("[FATAL ERROR] Missing test files in %s", prevDir)
	}

	// Check files and copy untargeted files from local path
	for _, f := range prevFiles {
		if f.IsDir() {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(prevDir, f.Name()))
		check(err, "")
		checksum := fmt.Sprintf("%x", md5.Sum(data))
		fName, ok := validHashData[checksum]
		if !ok {
			log.Fatalf("[FATAL ERROR] Invalid checksum for file %s (%s)", f.Name(), checksum)
		}
		check(copyFile(filepath.Join(pathFixturesUntargeted, fName), filepath.Join(newDir, f.Name())), "[SCRIPT ERROR] Failed to copy untargetedTest data")
		log.Printf("Removed target from %s", f.Name())
	}
}

func perf(dirTest, dirTrain, dirTestPred, dirTrainPred, dirPerf, filePerf string) {
	// Checking datas
	testFiles := checkData(dirTest)
	trainFiles := checkData(dirTrain)

	testPredFiles := checkData(dirTestPred)
	trainPredFiles := checkData(dirTrainPred)

	if len(testFiles) != len(testPredFiles) || len(trainFiles) != len(trainPredFiles) {
		log.Fatalf("[FATAL ERROR] Missing files. Test: %d, Test Pred: %d, Train: %d, Train Pred: %d",
			len(testFiles), len(testPredFiles), len(trainFiles), len(trainPredFiles))
	}

	// Generate performance
	rand.Seed(time.Now().Unix())

	testPerf := make(map[string]float64)
	trainPerf := make(map[string]float64)
	for _, f := range testPredFiles {
		testPerf[f.Name()] = rand.Float64()
		log.Printf("[perf] Computed perf on %s", f.Name())
	}
	for _, f := range trainPredFiles {
		trainPerf[f.Name()] = rand.Float64()
		log.Printf("[perf] Computed perf on %s", f.Name())
	}

	// Saving performance
	p := Perfuplet{
		Perf:      rand.Float64(),
		TrainPerf: trainPerf,
		TestPerf:  testPerf,
	}
	perfBytes, err := json.Marshal(p)
	check(err, "[SCRIPT ERROR] Failed to Marshal perf")
	_ = os.MkdirAll(dirPerf, 0777)
	check(ioutil.WriteFile(filepath.Join(dirPerf, filePerf), perfBytes, 0777), "[SCRIPT ERROR] Failed to WriteFile on perf")
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

func check(err error, msg string) {
	if err != nil {
		if msg == "" {
			log.Fatalln(fmt.Sprintf("%s: %s", "[FATAL ERROR]", err))
		}
		log.Fatalln(fmt.Sprintf("%s %s: %s", "[FATAL ERROR]", msg, err))
	}
}
