#!/bin/bash
cd src
rm -r ../bin/
# uses the container architecture and os
export GOARCH="amd64"
export GOOS="linux"
go build -o ../bin/

read  -n 1 -p "Press a key..."