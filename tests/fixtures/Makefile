ALGO_UUID=8f5c97ff-ee61-4cf1-a0ac-6852bac08408
PB_UUID=c89d0eb7-2336-48d7-873b-27073ccd363f

.PHONY: train pred detarget perf tar-image cp-data clean generate-fixtures register-algo orchestrator-clean-test

# Algo submission
train: cp-data algo/fastest/fastest
	@echo "\nRUNNING TRAIN"
	@echo "================================================================================="
	@docker run -v ${shell pwd}/data/train:/data/train \
			    -v ${shell pwd}/data/untargetedTest:/data/test \
			    -v ${shell pwd}/data/model:/data/model \
		        algo-fastest -T train -V /data
	@echo "=================================================================================\n"

pred: cp-data algo/fastest/fastest
	@echo "\nRUNNING PREDICT"
	@echo "================================================================================="
	@docker run -v ${shell pwd}/data/test:/data/test \
			    -v ${shell pwd}/data/test/pred:/data/test/pred \
			    -v ${shell pwd}/data/model:/data/model \
			    algo-fastest -T predict -V /data
	@echo "=================================================================================\n"


# Workflow
detarget: cp-data problem/fastest/problem_fastest
	@echo "\nRUNNING DETARGET"
	@echo "================================================================================="
	@docker run -v ${shell pwd}/data/test:/hidden_data/test \
			    -v ${shell pwd}/data/untargetedTest:/submission_data/test \
				problem-fastest \
				-T detarget -i /hidden_data -s /submission_data
	@echo "=================================================================================\n"


perf: cp-data problem/fastest/problem_fastest
	@echo "\nRUNNING PERF"
	@echo "================================================================================="
	@docker run -v ${shell pwd}/data:/hidden_data \
			    -v ${shell pwd}/data/untargetedTest:/submission_data/test \
			    -v ${shell pwd}/data/train:/submission_data/train \
				problem-fastest \
				-T perf -i /hidden_data -s /submission_data
	@echo "=================================================================================\n"


# Builds
algo/fastest/fastest: algo/fastest/Dockerfile algo/fastest/fastest.go
	go build --installsuffix cgo --ldflags '-extldflags \"-static\"' -o algo/fastest/fastest algo/fastest/fastest.go
	docker build -t algo-fastest algo/fastest/.

problem/fastest/problem_fastest: problem/fastest/Dockerfile problem/fastest/problem_fastest.go
	go build --installsuffix cgo --ldflags '-extldflags \"-static\"' -o problem/fastest/problem_fastest problem/fastest/problem_fastest.go
	docker build -t problem-fastest problem/fastest/.

cp-data:
	cp -r data_fastest data

# Build tar
tar-gz: algo/fastest/fastest problem/fastest/problem_fastest
	cd algo/fastest && tar cvzf fastest.tar.gz Dockerfile fastest fixtures
	cd problem/fastest && tar cvzf problem_fastest.tar.gz Dockerfile problem_fastest fixtures

# Cleaning Makefile outputs
clean:
	sudo rm -rf data
	sudo rm -rf algo/fastest/fastest algo/fastest/*.tar.gz
	sudo rm -rf problem/fastest/problem_fastest problem/fastest/*.tar.gz

# Generate Fixtures for tests
gen-fixtures: tar-gz
	@mkdir -p ../../data/fixtures/algo/fastest/
	@mkdir -p ../../data/fixtures/problem/fastest/
	@mkdir -p ../../data/fixtures/data/fastest/
	mv algo/fastest/fastest.tar.gz ../../data/fixtures/algo/fastest/${ALGO_UUID}
	mv problem/fastest/problem_fastest.tar.gz ../../data/fixtures/problem/fastest/${PB_UUID}
	cp -r data_fastest/train ../../data/fixtures/data/fastest
	cp -r data_fastest/test ../../data/fixtures/data/fastest

# Register the algo to the orchestrator, cleaning previous tests
register-algo: orchestrator-clean-test
	@echo "\nRegistering new algo..."
	@echo "================================================================================="
	@curl -H "Content-Type: application/json" -X POST -u $${USER_AUTH}:$${PWD_AUTH} -d '{"name":"fastest","problem":"2869781a-c481-4ed7-b88a-a5073bae8326","uuid":"8f5c97ff-ee61-4cf1-a0ac-6852bac08408"}' http://orchestrator.morpheo.co/algo

orchestrator-clean-test:
	@echo "Deleting learnuplets and algo from previous tests..."
	@echo "================================================================================="
	@echo 'db.learnuplet.deleteMany({"algo": "8f5c97ff-ee61-4cf1-a0ac-6852bac08408"})' | kubectl --namespace staging exec -i orchestrator-db-0 mongo orchestrator
	@echo 'db.algo.deleteMany({"uuid": "8f5c97ff-ee61-4cf1-a0ac-6852bac08408"})' | kubectl --namespace staging exec -i orchestrator-db-0 mongo orchestrator
