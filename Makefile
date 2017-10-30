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

BIN_TARGETS = morpheo-compute morpheo-storage
VENDOR_TARGETS = $(foreach TARGET, $(BIN_TARGETS), $(TARGET)-vendor)

COMPOSE_CMD = STORAGE_PORT=8081 COMPUTE_PORT=8082 ORCHESTRATOR_PORT=8083 NSQ_ADMIN_PORT=8085 \
							STORAGE_AUTH_USER=u STORAGE_AUTH_PASSWORD=p \
							ORCHESTRATOR_AUTH_USER=test ORCHESTRATOR_AUTH_PASSWORD=test \
							docker-compose

# Target configuration
.DEFAULT: devenv-start
.PHONY: $(BIN_TARGETS) $(VENDOR_TARGETS) devenv-start devenv-stop devenv-clean devenv-logs

$(BIN_TARGETS):
	@echo "\n**** $(subst morpheo-,,$@): builds ****" | tr a-z A-Z
	@$(MAKE) -C ../$@ bin

$(VENDOR_TARGETS):
	@echo "\n**** $(patsubst morpheo-%-vendor,%,$@): vendor update ****" | tr a-z A-Z
	@echo "[$(patsubst morpheo-%-vendor,%,$@)] Updating vendor..."
	@$(MAKE) -C ../$(subst -vendor,,$@) vendor

	@echo "[devenv] Replacing vendor/morpheo-go-packages by local repository..."
	@rm -rf ../$(subst -vendor,,$@)/vendor/github.com/MorpheoOrg
	@mkdir ../$(subst -vendor,,$@)/vendor/github.com/MorpheoOrg
	@cp -Rf ../morpheo-go-packages ../$(subst -vendor,,$@)/vendor/github.com/MorpheoOrg/morpheo-go-packages
	@rm -rf ../$(subst -vendor,,$@)/vendor/github.com/MorpheoOrg/morpheo-go-packages/vendor

devenv-start: $(VENDOR_TARGETS) $(BIN_TARGETS)
	@echo  "\n**** DEVENV: DOCKER-COMPOSE UP ****"
	$(COMPOSE_CMD) up -d --build
devenv-stop:
	$(COMPOSE_CMD) stop
devenv-clean:
	$(COMPOSE_CMD) down
devenv-logs:
	$(COMPOSE_CMD) logs --follow storage compute compute-worker dind-executor
