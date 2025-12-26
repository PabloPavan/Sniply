# syntax=docker/dockerfile:1

FROM golang:alpine AS build
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM golang:alpine AS debug
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /app/api ./cmd/api

EXPOSE 8080 40000
CMD ["/go/bin/dlv", "exec", "/app/api", "--headless", "--listen=:40000", "--api-version=2", "--accept-multiclient", "--log"]

FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /bin/api /app/api

EXPOSE 8080
CMD ["/app/api"]
