# go-simple-http

Start server
docker build -t go-simple-http-server .
docker run --rm -p 8081:8081 go-simple-http-server

Start client
docker build -t go-simple-http-client .
docker run --rm -e BASE_URL=http://host.docker.internal:8081 go-simple-http-client