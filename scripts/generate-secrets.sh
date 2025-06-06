#!/bin/bash

# Generate secure secrets for .env file

echo "Generating secure secrets..."

# Generate a secure JWT secret (32 bytes, base64 encoded)
JWT_SECRET=$(openssl rand -base64 32)

# Generate a secure database password (16 characters, alphanumeric)
DB_PASSWORD=$(openssl rand -base64 16 | tr -d "=+/" | cut -c1-16)

echo ""
echo "Add these to your .env file:"
echo "=========================="
echo "JWT_SECRET=$JWT_SECRET"
echo "DB_PASSWORD=$DB_PASSWORD"
echo ""
echo "IMPORTANT: Never commit these values to version control!"
echo ""
echo "To generate a new API key:"
echo "1. Go to https://openrouter.ai/keys"
echo "2. Create a new API key"
echo "3. Add it to your .env file as AI_API_KEY=your-new-key"
echo ""