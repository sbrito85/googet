#! /bin/sh
GOOS=windows go build -ldflags '-X main.version=1.6.0@1'
