APP=truco
TUI_DIR=bin/tui
GUI_DIR=bin/gui
HOST_GOOS := $(shell go env GOOS)
HOST_GOARCH := $(shell go env GOARCH)
HOST_TUI_NAME := $(APP)-tui-core-$(HOST_GOOS)-$(HOST_GOARCH)
WIN_TUI_NAME := $(APP)-tui-core-windows-amd64
define bin_ext
$(if $(filter windows,$(1)),.exe,)
endef
HOST_TUI_BIN := $(TUI_DIR)/$(HOST_TUI_NAME)$(call bin_ext,$(HOST_GOOS))
WIN_TUI_BIN := $(TUI_DIR)/$(WIN_TUI_NAME).exe

.PHONY: build run relay test windows browser clean browser-clean ffi ffi-macos ffi-linux ffi-windows linux-gtk flatpak-linux

# Naming rule: `truco-<type>-<client>-<platform>-<arch>[-<variant>]` with TUI binaries under `bin/tui`
# and any GUI bundles under `bin/gui`. This keeps CLI/GUI outputs distinct.
build:
	@mkdir -p $(TUI_DIR)
	go build -o $(HOST_TUI_BIN) ./cmd/truco

run:
	go run ./cmd/truco

relay:
	go run ./cmd/truco-relay

test:
	go test ./...

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

browser-clean:
	rm -rf browser-edition/dist

clean:
	rm -rf $(TUI_DIR) $(GUI_DIR)
