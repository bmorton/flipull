FROM golang:1.20.0-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o flipull *.go

FROM alpine:edge
COPY --from=builder ["/build/flipull", "/bin/flipull"]
