#!/bin/sh
# Health check script for Docker container

# Check if health endpoint responds
curl -f http://localhost:8080/health || exit 1