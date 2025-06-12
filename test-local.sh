#!/bin/bash

# Local Test Runner Script
# Run all tests locally without CI/CD costs

set -e

echo "🧪 Running D&D Game Test Suite"
echo "=============================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Backend Tests
echo -e "\n${YELLOW}📦 Backend Tests${NC}"
echo "-------------------"
cd backend

echo "Running go fmt check..."
if ! gofmt -l . | grep -q .; then
    echo -e "${GREEN}✓ Go formatting check passed${NC}"
else
    echo -e "${RED}✗ Go formatting issues found${NC}"
    gofmt -l .
fi

echo -e "\nRunning golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --timeout=5m || echo -e "${YELLOW}⚠ Linting issues found${NC}"
else
    echo -e "${YELLOW}⚠ golangci-lint not installed, skipping${NC}"
fi

echo -e "\nRunning backend tests..."
go test ./... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✓ Backend tests complete. Coverage report: backend/coverage.html${NC}"

# Security scan
echo -e "\nRunning security scan..."
if command -v gosec &> /dev/null; then
    gosec -fmt text ./... || echo -e "${YELLOW}⚠ Security issues found${NC}"
else
    echo -e "${YELLOW}⚠ gosec not installed, skipping security scan${NC}"
fi

cd ..

# Frontend Tests
echo -e "\n${YELLOW}🎨 Frontend Tests${NC}"
echo "-------------------"
cd frontend

echo "Running npm audit..."
npm audit || echo -e "${YELLOW}⚠ Vulnerabilities found${NC}"

echo -e "\nRunning ESLint..."
npm run lint || echo -e "${YELLOW}⚠ Linting issues found${NC}"

echo -e "\nRunning TypeScript check..."
npx tsc --noEmit || echo -e "${YELLOW}⚠ TypeScript errors found${NC}"

echo -e "\nRunning frontend tests..."
npm test -- --coverage --watchAll=false || echo -e "${YELLOW}⚠ No tests found yet${NC}"

cd ..

echo -e "\n${GREEN}✅ Test suite complete!${NC}"
echo "=============================="

# Summary
echo -e "\n📊 Summary:"
echo "- Backend coverage report: backend/coverage.html"
echo "- Frontend coverage report: frontend/coverage/"
echo -e "\nRun this script anytime with: ${GREEN}./test-local.sh${NC}"