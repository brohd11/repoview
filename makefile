
APP_NAME  = repoview
OUT_DIR   = build
DIST_DIR  = dist
VERSION   = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   = -ldflags "-s -w -X github.com/brohd11/repoview/cmd.version=$(VERSION)"
PLATFORMS = darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

HOST_OS   = $(shell go env GOOS)
HOST_ARCH = $(shell go env GOARCH)

.PHONY: build all test package clean $(PLATFORMS)

# Host build -> build/<os>-<arch>/repoview. Default target so the dev loop
# (../builds.sh -> make && ./install_unix.sh) compiles one target, not five.
build:
	go build $(LDFLAGS) -o $(OUT_DIR)/$(HOST_OS)-$(HOST_ARCH)/$(APP_NAME) .

test:
	go test ./...

# Cross-compile every release target.
all: $(PLATFORMS)

$(PLATFORMS):
	@os=$(word 1,$(subst /, ,$@)); arch=$(word 2,$(subst /, ,$@)); \
	ext=$$( [ "$$os" = "windows" ] && echo .exe || echo ); \
	echo "building $$os/$$arch"; \
	GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
	  go build $(LDFLAGS) -o $(OUT_DIR)/$$os-$$arch/$(APP_NAME)$$ext .

# Build all targets, then archive each into dist/ for a GitHub release.
#
# Archive names are version-less on purpose: it lets install.sh use GitHub's
# /releases/latest/download/<name> redirect and skip the API entirely. The release
# tag carries the version. Archives are flat -- one bare executable at the root.
package: all
	@mkdir -p $(DIST_DIR); \
	for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; \
	  ext=$$( [ "$$os" = "windows" ] && echo .exe || echo ); \
	  name=$(APP_NAME)-$$os-$$arch.zip; \
	  echo "packaging $$name"; \
	  rm -f $(DIST_DIR)/$$name; \
	  ( cd $(OUT_DIR)/$$os-$$arch && zip -q -j ../../$(DIST_DIR)/$$name $(APP_NAME)$$ext ); \
	done; \
	echo "done -> $(DIST_DIR)/"

clean:
	rm -rf $(OUT_DIR) $(DIST_DIR)
