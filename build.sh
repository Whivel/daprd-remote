#!/bin/bash

echo "Proxy..."
cd src
rm -r ../bin/
# uses the container architecture and os
export GOARCH="amd64"
export GOOS="linux"
go build -o ../bin/

echo "Docker..."
cd ../

docker build -f ./docker/Dockerfile -t whivel/remote-daprd:v0.1 ./bin/

read  -n 1 -p "Press a key..."