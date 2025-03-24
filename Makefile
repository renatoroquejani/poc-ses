.PHONY: build run clean swagger deps

# Variáveis
APP_NAME=poc-ses
MAIN_PATH=./cmd

# Comandos
deps:
	go mod tidy

swagger:
	swag init -g cmd/main.go

build:
	go build -o $(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)/main.go

clean:
	go clean
	rm -f $(APP_NAME)

test:
	go test -v ./...

all: deps swagger build
