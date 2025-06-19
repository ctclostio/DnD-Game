#!/bin/bash
# Production build script for D&D Game
# This script ensures the correct Docker stages are used for production builds

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Building D&D Game for Production${NC}"

# Check if we're in the project root
if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}❌ Error: Must run from project root directory${NC}"
    exit 1
fi

# Get build version (git commit or timestamp)
VERSION=${VERSION:-$(git rev-parse --short HEAD 2>/dev/null || date +%Y%m%d%H%M%S)}
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo -e "${YELLOW}📋 Build Information:${NC}"
echo "  Version: $VERSION"
echo "  Date: $BUILD_DATE"

# Build frontend for production
echo -e "\n${YELLOW}🏗️  Building Frontend (Production Stage)...${NC}"
docker build \
    --target production \
    --build-arg VERSION=$VERSION \
    --build-arg BUILD_DATE=$BUILD_DATE \
    --build-arg NODE_ENV=production \
    -t dnd-frontend:prod \
    -t dnd-frontend:$VERSION \
    -f frontend/Dockerfile.optimized \
    .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Frontend production build successful${NC}"
else
    echo -e "${RED}❌ Frontend build failed${NC}"
    exit 1
fi

# Build backend for production
echo -e "\n${YELLOW}🏗️  Building Backend (Final Stage)...${NC}"
docker build \
    --target final \
    --build-arg VERSION=$VERSION \
    --build-arg BUILD_DATE=$BUILD_DATE \
    -t dnd-backend:prod \
    -t dnd-backend:$VERSION \
    -f backend/Dockerfile.optimized \
    .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Backend production build successful${NC}"
else
    echo -e "${RED}❌ Backend build failed${NC}"
    exit 1
fi

# Verify builds
echo -e "\n${YELLOW}🔍 Verifying Production Builds...${NC}"

# Check frontend doesn't have source maps
echo -n "Checking frontend for source maps... "
if docker run --rm dnd-frontend:prod ls /usr/share/nginx/html/js/ 2>/dev/null | grep -q "\.map$"; then
    echo -e "${RED}❌ WARNING: Source maps found in production build!${NC}"
    exit 1
else
    echo -e "${GREEN}✅ No source maps found${NC}"
fi

# Check backend has no shell (using scratch base)
echo -n "Checking backend security... "
if docker run --rm --entrypoint sh dnd-backend:prod -c "echo test" 2>&1 | grep -q "not found"; then
    echo -e "${GREEN}✅ No shell access in production image${NC}"
else
    echo -e "${YELLOW}⚠️  Shell access available (consider using scratch base)${NC}"
fi

# Display image information
echo -e "\n${YELLOW}📊 Image Information:${NC}"
docker images | grep -E "REPOSITORY|dnd-" | grep -E "prod|$VERSION"

# Security scan reminder
echo -e "\n${YELLOW}🔒 Security Recommendations:${NC}"
echo "1. Run security scans:"
echo "   docker scan dnd-frontend:prod"
echo "   docker scan dnd-backend:prod"
echo ""
echo "2. Set production environment variables:"
echo "   - ENV=production"
echo "   - JWT_SECRET (64+ characters)"
echo "   - DB_SSLMODE=require"
echo ""
echo "3. Never use development stages in production"

echo -e "\n${GREEN}✨ Production build complete!${NC}"
echo -e "Tagged as: ${YELLOW}dnd-frontend:prod${NC} and ${YELLOW}dnd-backend:prod${NC}"
echo -e "Also tagged with version: ${YELLOW}$VERSION${NC}"