#!/bin/bash

# Big Meter Frontend Deployment Script
# This script builds the app with production CSP and deploys to web server

set -e  # Exit on error

# ============= CONFIGURATION =============
PRODUCTION_API="https://api.yourdomain.pwa.co.th"
PRODUCTION_LOGIN="https://intranet.pwa.co.th/login/webservice_login6.php"

# Optional: SSH deployment settings (uncomment to use)
# SERVER_USER="your-user"
# SERVER_HOST="your-server.com"
# SERVER_PATH="/var/www/big-meter-frontend"

# ============= BUILD =============
echo "🔨 Building for production..."

# Build with production environment variables
VITE_API_BASE_URL=$PRODUCTION_API \
VITE_LOGIN_API=$PRODUCTION_LOGIN \
pnpm build

# ============= FIX CSP =============
echo "🔒 Updating CSP (removing localhost)..."

# Remove localhost:8089 from CSP in dist/index.html
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' 's/http:\/\/localhost:8089 //g' dist/index.html
else
    # Linux
    sed -i 's/http:\/\/localhost:8089 //g' dist/index.html
fi

echo "✅ CSP updated - localhost removed from connect-src"

# ============= DEPLOY (Optional) =============
# Uncomment and configure for automatic deployment

# if [ -n "$SERVER_USER" ] && [ -n "$SERVER_HOST" ] && [ -n "$SERVER_PATH" ]; then
#     echo "🚀 Deploying to $SERVER_HOST..."
#     rsync -avz --delete dist/ ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/
#     echo "✅ Deployment complete!"
# else
#     echo "⚠️  Deployment skipped (not configured)"
#     echo "📦 Build ready in dist/ folder"
# fi

echo ""
echo "============================================"
echo "✅ Build complete!"
echo "📦 Output directory: ./dist/"
echo "============================================"
echo ""
echo "Next steps:"
echo "1. Copy dist/ folder to your web server"
echo "2. Configure web server (see README.md)"
echo "3. IMPORTANT: Use server CSP headers (recommended)"
echo ""
