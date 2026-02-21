APP=truco

.PHONY: build run test windows clean

build:
	go build -o bin/$(APP) ./cmd/truco

run:
	go run ./cmd/truco

test:
	go test ./...

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP).exe ./cmd/truco

clean:
	rm -rf bin
