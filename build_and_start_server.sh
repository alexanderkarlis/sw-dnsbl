#!/bin/sh

# add env vars here
. ./config.env
cat ./config.env

echo "downloading go deps..."
go mod download
echo "running tests..."
go test ./... -test.short
# go test ./server_test.go -test.short -v
echo "building..."
go build -o server server.go
echo "starting server"
./server
