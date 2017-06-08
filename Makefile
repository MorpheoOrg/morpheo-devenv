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
#
BIN_TARGETS = compute storage
LINK_TARGETS = $(foreach TARGET,$(BIN_TARGETS),$(TARGET)-link)

COMPOSE_CMD = STORAGE_PORT=8081 COMPUTE_PORT=8082 ORCHESTRATOR_PORT=8083 \
							NSQ_ADMIN_PORT=8085 STORAGE_AUTH_USER=u \
							STORAGE_AUTH_PASSWORD=p docker-compose

# Target configuration
.DEFAULT: devenv-start
.PHONY: bin $(BIN_TARGETS) link $(LINK_TARGETS) devenv-start devenv-clean

bin: $(BIN_TARGETS)

$(BIN_TARGETS): %: %-link
	$(MAKE) -C $@ bin

link: $(LINK_TARGETS)

$(LINK_TARGETS):
	$(MAKE) -C $(subst -link,,$@) vendor
	@echo "Symlinking local go-packages repo in $(subst -link,,$@) vendor dir"
	rm -rf $(subst -link,,$@)/vendor/github.com/MorpheoOrg/go-packages
	cp -Rf go-packages $(subst -link,,$@)/vendor/github.com/MorpheoOrg/go-packages

devenv-start: link bin
	$(COMPOSE_CMD) up -d --build
devenv-clean:
	$(COMPOSE_CMD) down
devenv-logs:
	$(COMPOSE_CMD) logs --follow storage compute compute-worker orchestrator dind-executor
