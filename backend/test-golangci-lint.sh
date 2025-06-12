#!/bin/bash
# Test script to verify golangci-lint configuration

echo "Testing golangci-lint configuration..."
echo "======================================="

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint is not installed. Installing v1.62.2..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2
fi

# Check version
echo "Current golangci-lint version:"
golangci-lint --version

# Test the configuration
echo -e "\nTesting configuration file..."
golangci-lint config path

# Run a dry run to check for configuration errors
echo -e "\nRunning dry run..."
golangci-lint run --no-config --disable-all --enable=typecheck ./... 2>&1 | head -20

echo -e "\nRunning with full configuration..."
golangci-lint run --timeout=5m ./... 2>&1 | head -50

echo -e "\nTest complete!"