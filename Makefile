APP=truco
TUI_DIR=bin/tui
GUI_DIR=bin/gui
RELAY_DIR=bin
WAILS_DIR=desktop/wails
HOST_GOOS := $(shell go env GOOS)
HOST_GOARCH := $(shell go env GOARCH)
HOST_TUI_NAME := $(APP)-tui-core-$(HOST_GOOS)-$(HOST_GOARCH)
HOST_RELAY_NAME := $(APP)-relay-$(HOST_GOOS)-$(HOST_GOARCH)
WIN_TUI_NAME := $(APP)-tui-core-windows-amd64
HOST_WAILS_NAME := $(APP)-gui-wails-$(HOST_GOOS)-$(HOST_GOARCH)
LINUX_WAILS_NAME := $(APP)-gui-wails-linux-amd64
WINDOWS_WAILS_NAME := $(APP)-gui-wails-windows-amd64.exe
define bin_ext
$(if $(filter windows,$(1)),.exe,)
endef
HOST_TUI_BIN := $(TUI_DIR)/$(HOST_TUI_NAME)$(call bin_ext,$(HOST_GOOS))
HOST_RELAY_BIN := $(RELAY_DIR)/$(HOST_RELAY_NAME)$(call bin_ext,$(HOST_GOOS))
WIN_TUI_BIN := $(TUI_DIR)/$(WIN_TUI_NAME).exe
HOST_WAILS_BIN := $(GUI_DIR)/wails/$(HOST_WAILS_NAME)$(call bin_ext,$(HOST_GOOS))
HOST_WAILS_APP := $(GUI_DIR)/wails/$(HOST_WAILS_NAME).app
LINUX_WAILS_BIN := $(GUI_DIR)/wails/$(LINUX_WAILS_NAME)
WINDOWS_WAILS_BIN := $(GUI_DIR)/wails/$(WINDOWS_WAILS_NAME)

.PHONY: build build-relay run relay test windows browser clean browser-clean ffi ffi-macos ffi-linux ffi-windows linux-gtk flatpak-linux verify-artifacts verify-browser-dist wails-frontend wails-typecheck wails-frontend-test wails-test wails-dev wails-build wails-build-linux wails-build-windows verify-wails

# Naming rule: `truco-<type>-<client>-<platform>-<arch>[-<variant>]` with TUI binaries under `bin/tui`
# and any GUI bundles under `bin/gui`. This keeps CLI/GUI outputs distinct.
build:
	@mkdir -p $(TUI_DIR)
	go build -o $(HOST_TUI_BIN) ./cmd/truco

build-relay:
	@mkdir -p $(RELAY_DIR)
	go build -o $(HOST_RELAY_BIN) ./cmd/truco-relay

run:
	go run ./cmd/truco

relay:
	go run ./cmd/truco-relay

test:
	go test ./internal/... ./browser-edition/... ./cmd/... | grep -v ebitengine || go test ./internal/... ./browser-edition/... ./cmd/...

ffi:
	go build -buildmode=c-shared -o bin/libtruco_core.dylib ./cmd/truco-core-ffi

ffi-macos:
	go build -buildmode=c-shared -o bin/libtruco_core.dylib ./cmd/truco-core-ffi
	install_name_tool -id "@rpath/libtruco_core.dylib" bin/libtruco_core.dylib
	cp bin/libtruco_core.h native/macos/Truco/Truco/truco_core.h

ffi-linux:
	GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o bin/libtruco_core.so ./cmd/truco-core-ffi

ffi-windows:
	GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -o bin/truco-core-ffi.dll ./cmd/truco-core-ffi

linux-gtk: ffi-linux
	mkdir -p native/linux-gtk/lib
	cp bin/libtruco_core.so native/linux-gtk/lib/libtruco_core.so
	cargo build --manifest-path native/linux-gtk/Cargo.toml --release

flatpak-linux:
	flatpak-builder --force-clean .flatpak-builder build/flatpak native/linux-gtk/dev.truco.Native.yaml

windows:
	@mkdir -p $(TUI_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(WIN_TUI_BIN) ./cmd/truco

browser:
	bash browser-edition/scripts/build-web.sh

wails-frontend:
	npm install --prefix $(WAILS_DIR)
	npm run build --prefix $(WAILS_DIR)

wails-typecheck:
	npm install --prefix $(WAILS_DIR)
	npm run typecheck --prefix $(WAILS_DIR)

wails-frontend-test:
	npm install --prefix $(WAILS_DIR)
	npm run test:frontend --prefix $(WAILS_DIR)

wails-test:
	cd $(WAILS_DIR) && go test -tags=wails .

wails-dev:
	cd $(WAILS_DIR) && wails dev

wails-build: wails-frontend
	@mkdir -p $(GUI_DIR)/wails
	cd $(WAILS_DIR) && wails build -clean -platform $(HOST_GOOS)/$(HOST_GOARCH) -o ../../$(HOST_WAILS_BIN)
	@if [ "$(HOST_GOOS)" = "darwin" ] && [ -d "$(WAILS_DIR)/build/bin/truco-gui-wails.app" ]; then \
		rm -rf "$(HOST_WAILS_APP)"; \
		cp -R "$(WAILS_DIR)/build/bin/truco-gui-wails.app" "$(HOST_WAILS_APP)"; \
	fi

wails-build-linux: wails-frontend
	@mkdir -p $(GUI_DIR)/wails
	cd $(WAILS_DIR) && wails build -clean -platform linux/amd64 -o ../../$(LINUX_WAILS_BIN)

wails-build-windows: wails-frontend
	@mkdir -p $(GUI_DIR)/wails
	cd $(WAILS_DIR) && wails build -clean -platform windows/amd64 -o ../../$(WINDOWS_WAILS_BIN)

verify-browser-dist:
	bash scripts/validate-browser-dist.sh

verify-wails: wails-typecheck wails-frontend wails-frontend-test wails-test
	@mkdir -p $(GUI_DIR)/wails
	cd $(WAILS_DIR) && wails build -dryrun -platform linux/amd64 -o ../../$(LINUX_WAILS_BIN)
	cd $(WAILS_DIR) && wails build -dryrun -platform windows/amd64 -o ../../$(WINDOWS_WAILS_BIN)

verify-artifacts: verify-browser-dist verify-wails
	bash scripts/validate-artifacts.sh

browser-clean:
	rm -rf browser-edition/dist

clean:
	rm -rf $(TUI_DIR) $(GUI_DIR) $(WAILS_DIR)/frontend/dist
