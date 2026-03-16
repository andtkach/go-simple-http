# go-simple-http

## Run locally
make run-server
make run-client

make build-server
make build-client

## Docker
make docker-build-server
make docker-run-server

make docker-build-client
make docker-run-client

make docker-build

## Start server
docker build -t go-simple-http-server ./server
docker run --rm -p 8081:8081 go-simple-http-server

Start client
docker build -t go-simple-http-client ./client
docker run --rm -e BASE_URL=http://host.docker.internal:8081 go-simple-http-client