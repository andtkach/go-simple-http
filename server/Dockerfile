# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

# Install CA certificates for module downloads over TLS.
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o /out/server .

FROM alpine:3.21 AS runtime
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/server /app/server

ENV HOST=0.0.0.0
ENV PORT=8081
EXPOSE 8081

USER app
ENTRYPOINT ["/app/server"]
