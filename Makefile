SHELL=/bin/bash
NIX_OS_ARCHS?=darwin-amd64 linux-amd64
WIN_OS_ARCHS?=windows-amd64
VERSION=latest
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_PATH=$(shell git rev-parse --show-toplevel)
GITHUB_REPO=$(shell basename $(GIT_PATH))
GITHUB_USER=$(shell dirname `git remote get-url --all origin` | sed 's/.*[:/]//')
GIT_TAG=$(VERSION)


DOCKER_CMD=docker
DOCKER_USER=stratumn
DOCKER_FILE=Dockerfile
DOCKER_BUILD=$(DOCKER_CMD) build
DOCKER_PUSH=$(DOCKER_CMD) push

GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_LIST=$(GO_CMD) list
GO_TEST=$(GO_CMD) test
GO_LINT_CMD=golangci-lint
GO_LINT=$(GO_LINT_CMD) run --build-tags="lint" --deadline=4m --disable="ineffassign" --disable="gas" --tests=false

GITHUB_RELEASE_FLAGS=--user '$(GITHUB_USER)' --repo '$(GITHUB_REPO)' --tag '$(GIT_TAG)'
GITHUB_RELEASE_RELEASE_FLAGS=$(GITHUB_RELEASE_FLAGS) --name '$(RELEASE_NAME)' --description "$$(cat $(RELEASE_NOTES_FILE))"
GITHUB_RELEASE_RELEASE=$(GITHUB_RELEASE_COMMAND) release $(GITHUB_RELEASE_RELEASE_FLAGS)
GITHUB_RELEASE_UPLOAD=$(GITHUB_RELEASE_COMMAND) upload $(GITHUB_RELEASE_FLAGS)
GITHUB_RELEASE_EDIT=$(GITHUB_RELEASE_COMMAND) edit $(GITHUB_RELEASE_RELEASE_FLAGS)
KEYBASE_CMD=keybase
KEYBASE_SIGN=$(KEYBASE_CMD) pgp sign

CMD=connector
DIST_DIR=dist
BUILD_SOURCES=$(shell find . -name '*.go' | grep -v 'mock' | grep -v 'test' | grep -v '_test.go')
TEST_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v 'mock' | grep -v 'test')
TEST_LIST=$(foreach package, $(TEST_PACKAGES), test_$(package))

NIX_EXECS=$(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD))
WIN_EXECS=$(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).exe)
EXECS=$(NIX_EXECS) $(WIN_EXECS)
DOCKER_EXEC=$(DIST_DIR)/linux-amd64/$(CMD)
SIGNATURES=$(foreach exec, $(EXECS), $(exec).sig)
NIX_ZIP_FILES=$(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).zip)
WIN_ZIP_FILES=$(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(CMD).zip)
ZIP_FILES=$(NIX_ZIP_FILES) $(WIN_ZIP_FILES)
GITHUB_UPLOAD_LIST=$(foreach file, $(ZIP_FILES), github_upload_$(firstword $(subst ., ,$(file))))


# == .PHONY ===================================================================
.PHONY: golangcilint test lint install build git_tag github_draft github_upload github_publish docker_image docker_push $(TEST_LIST) $(GITHUB_UPLOAD_LIST)

# == all ======================================================================
all: build

# == deps =====================================================================

golangcilint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# == build ====================================================================
build: $(EXECS)

BUILD_OS_ARCH=$(word 2, $(subst /, ,$@))
BUILD_OS=$(firstword $(subst -, ,$(BUILD_OS_ARCH)))
BUILD_ARCH=$(lastword $(subst -, ,$(BUILD_OS_ARCH)))

$(EXECS): $(BUILD_SOURCES)
	GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) $(GO_BUILD) -o $@

# == install ==================================================================
install:
	$(GO_CMD) build -o stratumn-connector
	mv stratumn-connector $(GOPATH)/bin

# == release ==================================================================
release: test lint build git_tag github_draft github_upload github_publish docker_image docker_push

# == test =====================================================================
test: $(TEST_LIST)

$(TEST_LIST): test_%:
	@$(GO_TEST) $*

# == lint =====================================================================
lint:
	@$(GO_LINT) ./...

# == sign =====================================================================
sign: $(SIGNATURES)

%.sig: %
	$(KEYBASE_SIGN) -d -i $* -o $@

# == zip ======================================================================
zip: $(ZIP_FILES)

ZIP_TMP_OS_ARCH_DIR=$(TMP_DIR)/$(BUILD_OS_ARCH)
ZIP_TMP_CMD_DIR=$(ZIP_TMP_OS_ARCH_DIR)/$(CMD)

%.zip: %.exe %.exe.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $*.exe $(ZIP_TMP_CMD_DIR)
	cp $*.exe.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	mv $(ZIP_TMP_CMD_DIR)/LICENSE $(ZIP_TMP_CMD_DIR)/LICENSE.txt
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(CMD){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

%.zip: % %.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $* $(ZIP_TMP_CMD_DIR)
	cp $*.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(CMD){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

# == git_tag ==================================================================
git_tag:
	git tag $(GIT_TAG)
	git push origin --tags

# == github_draft =============================================================
github_draft:
	@if [[ $prerelease != "false" ]]; then \
		echo $(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
		$(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
	else \
		echo $(GITHUB_RELEASE_RELEASE) --draft; \
		$(GITHUB_RELEASE_RELEASE) --draft; \
	fi

# == github_upload ============================================================
github_upload: $(GITHUB_UPLOAD_LIST)

$(GITHUB_UPLOAD_LIST): github_upload_%: %.zip
	$(GITHUB_RELEASE_UPLOAD) --file $*.zip --name $(CMD)-$(BUILD_OS_ARCH).zip

# == github_publish ===========================================================
github_publish:
	@if [[ "$(PRERELEASE)" != "false" ]]; then \
		echo $(GITHUB_RELEASE_EDIT) --pre-release; \
		$(GITHUB_RELEASE_EDIT) --pre-release; \
	else \
		echo $(GITHUB_RELEASE_EDIT); \
		$(GITHUB_RELEASE_EDIT); \
	fi

# == docker_image =============================================================
docker_image:
	$(DOCKER_BUILD) -f $(DOCKER_FILE) -t $(DOCKER_USER)/$(CMD) .

# == docker_push ==============================================================
docker_push:
	$(DOCKER_PUSH) $(DOCKER_USER)/$(CMD):$(VERSION)
