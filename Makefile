APP=truco

.PHONY: build run test windows browser clean browser-clean

build:
	go build -o bin/$(APP) ./cmd/truco

run:
	go run ./cmd/truco

test:
	go test ./...

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP).exe ./cmd/truco

browser:
	bash browser-edition/scripts/build-web.sh

browser-clean:
	rm -rf browser-edition/dist

clean:
	rm -rf bin
