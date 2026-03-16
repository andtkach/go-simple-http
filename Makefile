.PHONY: help tidy run-server build-server build-client run-client docker-build-server docker-build-client docker-build docker-run-server docker-run-client

help:
	@echo Available targets:
	@echo   make tidy        - Run go mod tidy for server and client
	@echo   make run-server  - Run the Go server locally
	@echo   make build-server - Build the Go server binary
	@echo   make build-client - Build the Go client binary
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

build-server:
	go -C server build .

build-client:
	go -C client build .

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