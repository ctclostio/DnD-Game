#!/bin/sh
set -e

echo "Injecting runtime environment variables..."

# Define the environment variables to inject
# These can be overridden at runtime
REACT_APP_API_URL=${REACT_APP_API_URL:-"http://localhost:8080"}
REACT_APP_WS_URL=${REACT_APP_WS_URL:-"ws://localhost:8080"}
REACT_APP_ENVIRONMENT=${REACT_APP_ENVIRONMENT:-"production"}
REACT_APP_VERSION=${REACT_APP_VERSION:-"1.0.0"}
REACT_APP_SENTRY_DSN=${REACT_APP_SENTRY_DSN:-""}
REACT_APP_GA_TRACKING_ID=${REACT_APP_GA_TRACKING_ID:-""}

# Create runtime config file
cat > /usr/share/nginx/html/runtime-env.js <<EOF
window._env_ = {
  REACT_APP_API_URL: "${REACT_APP_API_URL}",
  REACT_APP_WS_URL: "${REACT_APP_WS_URL}",
  REACT_APP_ENVIRONMENT: "${REACT_APP_ENVIRONMENT}",
  REACT_APP_VERSION: "${REACT_APP_VERSION}",
  REACT_APP_SENTRY_DSN: "${REACT_APP_SENTRY_DSN}",
  REACT_APP_GA_TRACKING_ID: "${REACT_APP_GA_TRACKING_ID}"
};
EOF

# Update index.html to include runtime config
if [ -f /usr/share/nginx/html/index.html ]; then
  # Check if runtime-env.js is already included
  if ! grep -q "runtime-env.js" /usr/share/nginx/html/index.html; then
    # Insert runtime-env.js script tag before closing head tag
    sed -i 's|</head>|<script src="/runtime-env.js"></script></head>|' /usr/share/nginx/html/index.html
  fi
fi

# Update CSP header with actual API URL if needed
if [ ! -z "$REACT_APP_API_URL" ]; then
  # Extract domain from API URL
  API_DOMAIN=$(echo $REACT_APP_API_URL | awk -F[/:] '{print $4}')
  
  # Update nginx config to include API domain in CSP
  if [ -f /etc/nginx/conf.d/security-headers.conf ]; then
    sed -i "s|connect-src 'self'|connect-src 'self' $REACT_APP_API_URL|g" /etc/nginx/conf.d/security-headers.conf
  fi
fi

echo "Runtime environment injection completed."