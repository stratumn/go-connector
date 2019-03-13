DIST_DIR=dist

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

CMD=connector
BUILD_SOURCES=$(shell find . -name '*.go' | grep -v 'mock' | grep -v 'test' | grep -v '_test.go')
TEST_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v 'mock' | grep -v 'test')

TEST_LIST=$(foreach package, $(TEST_PACKAGES), test_$(package))

# == .PHONY ===================================================================
.PHONY: golangcilint test lint build docker_image $(TEST_LIST)

# == all ======================================================================
all: build

# == deps =====================================================================

golangcilint:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# == build ====================================================================
build: $(CMD)

$(CMD): $(BUILD_SOURCES)
	$(GO_BUILD) -o $(DIST_DIR)/$@

# == test =====================================================================
test: $(TEST_LIST)

$(TEST_LIST): test_%:
	@$(GO_TEST) $*

# == lint =====================================================================
lint:
	@$(GO_LINT) ./...


# == docker_image =============================================================
docker_image:
	$(DOCKER_BUILD) -f $(DOCKER_FILE) -t $(DOCKER_USER)/$(CMD) .
