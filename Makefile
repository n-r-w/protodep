ROOT_PACKAGE := github.com/n-r-w/protodep
VERSION_PACKAGE := $(ROOT_PACKAGE)/version
LDFLAG_GIT_COMMIT := "$(VERSION_PACKAGE).gitCommit"
LDFLAG_GIT_COMMIT_FULL := "$(VERSION_PACKAGE).gitCommitFull"
LDFLAG_BUILD_DATE := "$(VERSION_PACKAGE).buildDate"
LDFLAG_VERSION := "$(VERSION_PACKAGE).version"

.PHONY: tidy
tidy:
	go mod tidy

APP := protodep
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

build: tidy version
		$(eval GIT_COMMIT := $(shell git describe --tags --always))
		$(eval GIT_COMMIT_FULL := $(shell git rev-parse HEAD))
		$(eval BUILD_DATE := $(shell date '+%Y%m%d'))
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s -X $(LDFLAG_GIT_COMMIT)=$(GIT_COMMIT) -X $(LDFLAG_GIT_COMMIT_FULL)=$(GIT_COMMIT_FULL) -X $(LDFLAG_BUILD_DATE)=$(BUILD_DATE) -X $(LDFLAG_VERSION)=$(GIT_COMMIT)" \
			-o bin/protodep main.go

define build-artifact
		$(eval GIT_COMMIT := $(shell git describe --tags --always))
		$(eval GIT_COMMIT_FULL := $(shell git rev-parse HEAD))
		$(eval BUILD_DATE := $(shell date '+%Y%m%d'))
		GOOS=$(1) GOARCH=$(2) go build -ldflags="-w -s -X $(LDFLAG_GIT_COMMIT)=$(GIT_COMMIT) -X $(LDFLAG_GIT_COMMIT_FULL)=$(GIT_COMMIT_FULL) -X $(LDFLAG_BUILD_DATE)=$(BUILD_DATE) -X $(LDFLAG_VERSION)=$(BUILD_DATE)-$(GIT_COMMIT)" \
			-o artifacts/$(3) main.go
		cd artifacts && tar cvzf $(APP)_$(1)_$(2).tar.gz $(3)
		rm ./artifacts/$(3)
		@echo [INFO]build success: $(1)_$(2)
endef

.PHONY: build-all
build-all: tidy
		$(call build-artifact,linux,amd64,$(APP))
		$(call build-artifact,darwin,arm64,$(APP))

.PHONY: clean
clean:
	rm -rf bin
	rm -rf vendor
	rm -rf artifacts

.PHONY: test
test:
	go test -v ./...

