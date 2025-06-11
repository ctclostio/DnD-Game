#!/bin/bash

# Run GitHub Actions locally with act
# Install act first: https://github.com/nektos/act

echo "🎭 Running GitHub Actions Locally with act"
echo "=========================================="

# Check if act is installed
if ! command -v act &> /dev/null; then
    echo "act is not installed. Install it with:"
    echo ""
    echo "  # On Linux/macOS:"
    echo "  curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    echo ""
    echo "  # Or with brew:"
    echo "  brew install act"
    echo ""
    echo "  # Or download from: https://github.com/nektos/act/releases"
    exit 1
fi

# Run the CI workflow locally
echo "Running CI workflow..."
act -W .github/workflows/ci.yml \
    --container-architecture linux/amd64 \
    -P ubuntu-latest=catthehacker/ubuntu:act-latest \
    --secret GITHUB_TOKEN="${GITHUB_TOKEN:-dummy}" \
    -j backend-test \
    -j backend-lint \
    -j frontend-test \
    -j frontend-lint

echo "✅ Local CI run complete!"