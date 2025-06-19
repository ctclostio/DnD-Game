#!/bin/bash
# Script to validate Docker security configurations

set -e

echo "🔒 Validating Docker Security Configurations..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for issues
ISSUES=0

# Function to check if file contains pattern
check_pattern() {
    local file=$1
    local pattern=$2
    local description=$3
    
    if grep -q "$pattern" "$file" 2>/dev/null; then
        echo -e "${RED}❌ $file: $description${NC}"
        ((ISSUES++))
    else
        echo -e "${GREEN}✓ $file: No $description found${NC}"
    fi
}

# Function to check if .dockerignore exists and contains patterns
check_dockerignore() {
    local dir=$1
    local dockerignore="$dir/.dockerignore"
    
    echo -e "\n${YELLOW}Checking $dockerignore...${NC}"
    
    if [ ! -f "$dockerignore" ]; then
        echo -e "${RED}❌ Missing $dockerignore${NC}"
        ((ISSUES++))
        return
    fi
    
    # Check for essential exclusions
    local required_patterns=(".env" ".git" "*.key" "*.pem" "node_modules" "*.log")
    for pattern in "${required_patterns[@]}"; do
        if grep -q "$pattern" "$dockerignore"; then
            echo -e "${GREEN}✓ $dockerignore excludes: $pattern${NC}"
        else
            echo -e "${YELLOW}⚠ $dockerignore should exclude: $pattern${NC}"
        fi
    done
}

# Function to check Dockerfile security
check_dockerfile() {
    local dockerfile=$1
    
    echo -e "\n${YELLOW}Checking $dockerfile...${NC}"
    
    if [ ! -f "$dockerfile" ]; then
        echo -e "${YELLOW}⚠ Skipping $dockerfile (not found)${NC}"
        return
    fi
    
    # Check for recursive COPY
    if grep -E "^COPY \. \." "$dockerfile" >/dev/null; then
        echo -e "${RED}❌ Found recursive COPY . . in $dockerfile${NC}"
        ((ISSUES++))
    else
        echo -e "${GREEN}✓ No recursive COPY . . found${NC}"
    fi
    
    # Check for USER directive
    if grep -q "^USER " "$dockerfile"; then
        echo -e "${GREEN}✓ USER directive found (non-root execution)${NC}"
    else
        echo -e "${RED}❌ No USER directive found (runs as root)${NC}"
        ((ISSUES++))
    fi
    
    # Check for secret exposure in ENV or ARG
    if grep -E "^(ENV|ARG).*(password|secret|token|key)" "$dockerfile" >/dev/null; then
        echo -e "${YELLOW}⚠ Potential secret in ENV/ARG directives${NC}"
    fi
    
    # Check for latest tags
    if grep -E "FROM.*:latest" "$dockerfile" >/dev/null; then
        echo -e "${YELLOW}⚠ Using :latest tag (consider pinning versions)${NC}"
    fi
    
    # Check for insecure COPY --chown patterns
    if grep -E "^COPY.*--chown.*[^/]$" "$dockerfile" | grep -v "\-\-chmod" >/dev/null; then
        echo -e "${RED}❌ Found COPY --chown without explicit permissions${NC}"
        ((ISSUES++))
    else
        echo -e "${GREEN}✓ No insecure COPY --chown patterns found${NC}"
    fi
    
    # Check for proper use of --chmod
    if grep -E "^COPY.*--chmod=[0-9]+" "$dockerfile" >/dev/null; then
        echo -e "${GREEN}✓ Using --chmod to set explicit permissions${NC}"
    fi
}

# Main checks
echo "Starting security validation..."

# Check backend
check_dockerignore "backend"
check_dockerfile "backend/Dockerfile"
check_dockerfile "backend/Dockerfile.optimized"

# Check frontend
check_dockerignore "frontend"
check_dockerfile "frontend/Dockerfile"
check_dockerfile "frontend/Dockerfile.optimized"

# Check root
check_dockerignore "."

# Summary
echo -e "\n${YELLOW}========================================${NC}"
if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ All security checks passed!${NC}"
else
    echo -e "${RED}❌ Found $ISSUES security issues${NC}"
    exit 1
fi