#!/bin/bash
cd src
rm -r ../bin/
# uses the user architecture and os
go build -o ../bin/

read  -n 1 -p "Press a key..."