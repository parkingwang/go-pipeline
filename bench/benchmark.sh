#!/bin/bash

# Run with config file: $1
go run ../cmd/main.go -k 10s \
    -with-pprof-cpu -with-pprof-cpu-output profiling-cpu.out

go-torch -b profiling-cpu.out