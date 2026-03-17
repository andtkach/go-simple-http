.PHONY: help tidy run-server build-windows-server build-windows-client build-linux-server build-linux-client build-linux run-client docker-build-server docker-build-client docker-build docker-run-server docker-run-client

help:
	@echo Available targets:
	@echo   make tidy        - Run go mod tidy for server and client
	@echo   make run-server  - Run the Go server locally
	@echo   make build-windows-server - Build the Go server binary for Windows
	@echo   make build-windows-client - Build the Go client binary for Windows
	@echo   make build-linux-server - Build the Go server binary for Linux
	@echo   make build-linux-client - Build the Go client binary for Linux
	@echo   make build-linux - Build both server and client binaries for Linux
	@echo   make run-client  - Run the Go client locally
	@echo   make docker-build-server - Build the server Docker image
	@echo   make docker-build-client - Build the client Docker image
	@echo   make docker-build - Build both Docker images
	@echo   make docker-run-server - Run the server Docker container
	@echo   make docker-run-client - Run the client Docker container

tidy:
	go -C server mod tidy
	go -C client mod tidy

run-server:
	go -C server run .

build-windows-server:
	go -C server build .

build-windows-client:
	go -C client build .

build-linux-server:
	GOOS=linux GOARCH=amd64 go -C server build -o server-linux .

build-linux-client:
	GOOS=linux GOARCH=amd64 go -C client build -o client-linux .

build-linux: build-linux-server build-linux-client

run-client:
	go -C client run .

docker-build-server:
	docker build -t go-simple-http-server ./server

docker-build-client:
	docker build -t go-simple-http-client ./client

docker-build: docker-build-server docker-build-client

docker-run-server:
	docker run --rm -p 8081:8081 go-simple-http-server

docker-run-client:
	docker run --rm -e BASE_URL=http://host.docker.internal:8081 go-simple-http-client