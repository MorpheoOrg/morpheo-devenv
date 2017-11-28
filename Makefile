#
# Copyright Morpheo Org. 2017
#
# contact@morpheo.co
#
# This software is part of the Morpheo project, an open-source machine
# learning platform.
#
# This software is governed by the CeCILL license, compatible with the
# GNU GPL, under French law and abiding by the rules of distribution of
# free software. You can  use, modify and/ or redistribute the software
# under the terms of the CeCILL license as circulated by CEA, CNRS and
# INRIA at the following URL "http://www.cecill.info".
#
# As a counterpart to the access to the source code and  rights to copy,
# modify and redistribute granted by the license, users are provided only
# with a limited warranty  and the software's author,  the holder of the
# economic rights,  and the successive licensors  have only  limited
# liability.
#
# In this respect, the user's attention is drawn to the risks associated
# with loading,  using,  modifying and/or developing or reproducing the
# software by the user in light of its specific status of free software,
# that may mean  that it is complicated to manipulate,  and  that  also
# therefore means  that it is reserved for developers  and  experienced
# professionals having in-depth computer knowledge. Users are therefore
# encouraged to load and test the software's suitability as regards their
# requirements in conditions enabling the security of their systems and/or
# data to be ensured and,  more generally, to use and operate it in the
# same conditions as regards security.
#
# The fact that you are presently reading this means that you have had
# knowledge of the CeCILL license and that you accept its terms.

BIN_TARGETS = compute storage
VENDOR_TARGETS = $(foreach TARGET, $(BIN_TARGETS), $(TARGET)-vendor)

COMPOSE_CMD = STORAGE_PORT=8081 STORAGE_AUTH_USER=u STORAGE_AUTH_PASSWORD=p \
  			  ORCHESTRATOR_PORT=8083 ORCHESTRATOR_AUTH_USER=test ORCHESTRATOR_AUTH_PASSWORD=test \
  			  MONGO_PORT=27017 \
			  COMPUTE_PORT=8082 NSQ_ADMIN_PORT=8085 \
			  docker-compose

PATH_METADATA = PATH_METADATA=$(GOPATH)/src/github.com/MorpheoOrg/morpheo-devenv/tests/fixtures.yaml
PATH_DATA = PATH_DATA=$(GOPATH)/src/github.com/MorpheoOrg/morpheo-devenv/data/fixtures

# Target configuration
.DEFAULT: up
.PHONY: $(BIN_TARGETS) $(VENDOR_TARGETS) up stop logs down clean tests full-tests

$(BIN_TARGETS):
	@echo "\n**** [$@] builds ****" | tr a-z A-Z
	@$(MAKE) -C ../morpheo-$@ bin

$(VENDOR_TARGETS):
	@echo "\n**** [$(subst -vendor,,$@)] vendor update ****" | tr a-z A-Z
	@$(MAKE) -C ../morpheo-$(subst -vendor,,$@) vendor vendor-replace-local

up: $(VENDOR_TARGETS) $(BIN_TARGETS)
	@echo  "\n**** [DEVENV] DOCKER-COMPOSE UP ****"
	$(COMPOSE_CMD) up -d --build
stop:
	$(COMPOSE_CMD) stop
logs:
	$(COMPOSE_CMD) logs --follow storage compute compute-worker orchestrator
down:
	$(COMPOSE_CMD) down
clean: down
	sudo rm -rf data/mongo data/storage data/postgresql

tests:
	@$(MAKE) -C ../morpheo-go-packages vendor
	@$(PATH_METADATA) $(PATH_DATA) go run ./tests/integration.go

full-tests:
	$(MAKE) -C ../morpheo-compute tests
	$(MAKE) -C ../morpheo-storage tests
	$(MAKE) tests
