#!/bin/bash

set -e

echo "Building ghissues..."
go build -o ghissues ./cmd/ghissues

echo -e "\n=== Test 1: Running without config (should prompt for setup) ==="
# Create a temporary directory for test
TEST_DIR=$(mktemp -d)
export HOME="$TEST_DIR"
echo "Using test directory: $TEST_DIR"

# Run ghissues - should show setup message
echo -e "\nRunning ghissues for the first time:"
./ghissues || true  # Allow non-zero exit for now since TUI not implemented

echo -e "\n=== Test 2: Check config command ==="
echo -e "\nRunning ghissues config --help:"
./ghissues config --help

echo -e "\n=== Test 3: Cleanup ==="
rm -rf "$TEST_DIR"
rm -f ghissues

echo -e "\n✅ All tests completed!"