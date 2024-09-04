#!/bin/bash

# Run the Go application excluding _test.go files
go run $(ls *.go | grep -v '_test.go')
