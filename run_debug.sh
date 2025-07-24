#!/bin/bash

echo "Building and running chess engine debug tests..."

# Build the debug runner
go build -o debug_runner debug_runner.go

if [ $? -eq 0 ]; then
    echo "Build successful. Running debug tests..."
    ./debug_runner
else
    echo "Build failed!"
    exit 1
fi