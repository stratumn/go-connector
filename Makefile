DIST_DIR=dist

DOCKER_CMD=docker
DOCKER_USER=stratumn
DOCKER_FILE=Dockerfile
DOCKER_BUILD=$(DOCKER_CMD) build

GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_BUILD_PLUGIN=$(GO_BUILD) -buildmode=plugin
GO_LIST=$(GO_CMD) list

EXEC=connector
BUILD_SOURCES=$(shell find . -name '*.go' | grep -v 'mock' | grep -v 'test' | grep -v '_test.go')

PLUGINS=$(shell $(GO_LIST) ./src/plugins/.../...)

# == all ======================================================================
all: build

# == build ====================================================================
build: $(EXEC) $(PLUGINS)

$(EXEC): $(BUILD_SOURCES)
	$(GO_BUILD) -o $(DIST_DIR)/$@

PLUGIN_NAME=$(lastword $(subst /, ,$@)).so
$(PLUGINS):
	$(GO_BUILD_PLUGIN) -o $(DIST_DIR)/$(PLUGIN_NAME) $@

# == docker_image =============================================================
docker_image:
	$(DOCKER_BUILD) -f $(DOCKER_FILE) -t $(DOCKER_USER)/$(EXEC) .
