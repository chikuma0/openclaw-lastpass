GO ?= go
BINARY ?= openclaw-lastpass
BUILD_DIR ?= dist

.PHONY: build fmt test release-assets

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY) ./cmd/openclaw-lastpass

release-assets:
	./scripts/build-release.sh $(VERSION) $(BUILD_DIR)/release
