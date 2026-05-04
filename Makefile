# Copyright (c) 2026, Ampere Computing LLC.
#  
# SPDX-License-Identifier: BSD-3-Clause

GO ?= go
BINARY ?= acr
OUT_DIR ?= dist
BUILD_FLAGS ?= -trimpath -ldflags="-s -w"
NATIVE_GOARCH := $(shell $(GO) env GOARCH)

.PHONY: build build-amd64 build-arm64 build-all clean

build: $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=$(NATIVE_GOARCH) $(GO) build $(BUILD_FLAGS) -o $(OUT_DIR)/$(BINARY)-linux-$(NATIVE_GOARCH) .

build-amd64: $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(OUT_DIR)/$(BINARY)-linux-amd64 .

build-arm64: $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o $(OUT_DIR)/$(BINARY)-linux-arm64 .

build-all: build-amd64 build-arm64

$(OUT_DIR):
	mkdir -p $(OUT_DIR)

clean:
	rm -rf $(OUT_DIR)
