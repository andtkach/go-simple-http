# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/client .

FROM alpine:3.21 AS runtime
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/client /app/client

USER app
ENTRYPOINT ["/app/client"]
