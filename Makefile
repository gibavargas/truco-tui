APP=truco

.PHONY: build run relay test windows browser clean browser-clean ffi ffi-macos ffi-linux

build:
	go build -o bin/$(APP) ./cmd/truco

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

ffi-linux:
	GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o bin/libtruco_core.so ./cmd/truco-core-ffi

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP).exe ./cmd/truco

browser:
	bash browser-edition/scripts/build-web.sh

browser-clean:
	rm -rf browser-edition/dist

clean:
	rm -rf bin
