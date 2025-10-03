# Deployment Guide

Complete guide for deploying Big Meter Frontend to production.

## Quick Start

### 1. Build for Production

```bash
# Set your production API URLs
export VITE_API_BASE_URL=https://api.yourdomain.pwa.co.th
export VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php

# Build
pnpm build
```

### 2. Deploy dist/ folder

```bash
# Copy to web server (example)
scp -r dist/* user@server:/var/www/big-meter-frontend/
```

### 3. Configure Web Server

See server-specific instructions below.

---

## Important: Content Security Policy (CSP)

⚠️ **The `index.html` includes `http://localhost:8089` for development. You MUST override this in production!**

### Two Options:

#### Option 1: Use Server Headers (✅ RECOMMENDED)

Configure your web server to send CSP headers - this automatically overrides the meta tag:

- **Nginx:** See `nginx.conf` file
- **Apache:** See `.htaccess` file

#### Option 2: Remove from index.html

Use the deployment script:

```bash
./deploy.sh
```

This automatically removes localhost from CSP during build.

---

## Method 1: Nginx Deployment (Recommended)

### Step 1: Build

```bash
# Create .env.production
cat > .env.production << EOF
VITE_API_BASE_URL=https://api.yourdomain.pwa.co.th
VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php
EOF

# Build
pnpm build
```

### Step 2: Copy Files

```bash
# Create directory
sudo mkdir -p /var/www/big-meter-frontend

# Copy dist contents
sudo cp -r dist/* /var/www/big-meter-frontend/

# Set permissions
sudo chown -R www-data:www-data /var/www/big-meter-frontend
sudo chmod -R 755 /var/www/big-meter-frontend
```

### Step 3: Configure Nginx

```bash
# Copy provided config
sudo cp nginx.conf /etc/nginx/sites-available/big-meter-frontend

# Edit the config - IMPORTANT: Update these values!
sudo nano /etc/nginx/sites-available/big-meter-frontend
# Change:
#   - server_name
#   - root path (if different)
#   - API URL in CSP header

# Enable site
sudo ln -s /etc/nginx/sites-available/big-meter-frontend /etc/nginx/sites-enabled/

# Test config
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### Step 4: Set Up SSL (Recommended)

```bash
# Using Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.pwa.co.th

# Certbot will automatically update nginx config for HTTPS
```

---

## Method 2: Apache Deployment

### Step 1: Build

```bash
# Same as Nginx - create .env.production and build
pnpm build
```

### Step 2: Copy Files

```bash
# Copy dist contents to web root
sudo cp -r dist/* /var/www/html/big-meter/

# Copy .htaccess
sudo cp .htaccess /var/www/html/big-meter/

# Edit .htaccess - Update API URL in CSP header
sudo nano /var/www/html/big-meter/.htaccess

# Set permissions
sudo chown -R www-data:www-data /var/www/html/big-meter
sudo chmod -R 755 /var/www/html/big-meter
```

### Step 3: Enable Required Modules

```bash
sudo a2enmod rewrite
sudo a2enmod headers
sudo a2enmod deflate
sudo a2enmod expires

sudo systemctl restart apache2
```

### Step 4: Configure VirtualHost

Edit `/etc/apache2/sites-available/big-meter.conf`:

```apache
<VirtualHost *:80>
    ServerName yourdomain.pwa.co.th
    DocumentRoot /var/www/html/big-meter

    <Directory /var/www/html/big-meter>
        AllowOverride All
        Require all granted
    </Directory>

    ErrorLog ${APACHE_LOG_DIR}/big-meter-error.log
    CustomLog ${APACHE_LOG_DIR}/big-meter-access.log combined
</VirtualHost>
```

Enable and restart:

```bash
sudo a2ensite big-meter
sudo systemctl restart apache2
```

---

## Method 3: Docker Deployment

### Dockerfile

Already created! Use the example from README.md:

```bash
# Build image
docker build \
  --build-arg VITE_API_BASE_URL=https://api.yourdomain.pwa.co.th \
  --build-arg VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php \
  -t big-meter-frontend .

# Run container
docker run -d -p 80:80 --name big-meter big-meter-frontend
```

---

## Method 4: Using Deployment Script

### Configure deploy.sh

Edit `deploy.sh` and update:

```bash
PRODUCTION_API="https://api.yourdomain.pwa.co.th"
PRODUCTION_LOGIN="https://intranet.pwa.co.th/login/webservice_login6.php"

# Optional: Configure automatic deployment
SERVER_USER="your-user"
SERVER_HOST="your-server.com"
SERVER_PATH="/var/www/big-meter-frontend"
```

### Run Deployment

```bash
chmod +x deploy.sh
./deploy.sh
```

This script:

1. ✅ Builds with production env vars
2. ✅ Removes localhost from CSP
3. ✅ Optionally deploys via rsync

---

## Verification Checklist

After deployment, verify:

### 1. Check CSP Headers

```bash
# Check response headers
curl -I https://yourdomain.pwa.co.th

# Should see:
# Content-Security-Policy: default-src 'self'; ...
# (WITHOUT http://localhost:8089)
```

### 2. Browser DevTools

1. Open site in browser
2. Open DevTools (F12)
3. **Console tab:** No CSP errors
4. **Network tab:**
   - API calls go to production URL
   - Status 200 for index.html
5. **Application tab:**
   - Check localStorage works
   - Check session persistence

### 3. Functionality Tests

- [ ] Login works
- [ ] Branch dropdown loads
- [ ] Data displays correctly
- [ ] Export to Excel works
- [ ] Responsive on mobile

### 4. Security Headers

Check all security headers are present:

```bash
curl -I https://yourdomain.pwa.co.th | grep -E "Content-Security-Policy|X-Frame-Options|X-Content-Type-Options"
```

Should see:

- ✅ Content-Security-Policy
- ✅ X-Frame-Options: DENY
- ✅ X-Content-Type-Options: nosniff

---

## Troubleshooting

### Issue: White screen after deployment

**Check:**

1. Browser console for errors
2. Network tab for 404s
3. Web server error logs

**Fix:**

- Ensure SPA routing is configured (try_files in Nginx, mod_rewrite in Apache)

### Issue: API calls fail

**Check:**

1. CSP headers (should allow your API domain)
2. CORS configuration on backend
3. API URL is correct in build

**Fix:**

```bash
# Rebuild with correct API URL
VITE_API_BASE_URL=https://correct-api-url pnpm build
```

### Issue: CSP errors in browser

**Check:**

```bash
# View actual CSP header
curl -I https://yourdomain.pwa.co.th | grep Content-Security-Policy
```

**Fix:**

- Update CSP in nginx.conf or .htaccess
- Reload web server

### Issue: Assets not loading

**Check:**

1. File permissions (should be readable)
2. Path in server config

**Fix:**

```bash
sudo chmod -R 755 /var/www/big-meter-frontend
sudo chown -R www-data:www-data /var/www/big-meter-frontend
```

---

## Production Checklist

Before going live:

- [ ] Environment variables set correctly
- [ ] Built with `pnpm build` (not dev server)
- [ ] Localhost removed from CSP
- [ ] HTTPS configured with valid certificate
- [ ] Security headers configured
- [ ] CORS configured on backend
- [ ] API endpoints are accessible
- [ ] Caching headers set for assets
- [ ] Gzip/Brotli compression enabled
- [ ] Error logging configured
- [ ] Backup plan in place

---

## Rollback Procedure

If deployment fails:

### Option 1: Keep previous version

```bash
# Backup before deploying
sudo cp -r /var/www/big-meter-frontend /var/www/big-meter-frontend.backup

# Rollback if needed
sudo rm -rf /var/www/big-meter-frontend
sudo mv /var/www/big-meter-frontend.backup /var/www/big-meter-frontend
```

### Option 2: Redeploy previous build

```bash
# Tag builds with git
git tag -a v1.0.0 -m "Release 1.0.0"
git push origin v1.0.0

# Rollback to previous tag
git checkout v1.0.0
pnpm build
# Deploy...
```

---

## Files Summary

| File                    | Purpose              | When to Use                   |
| ----------------------- | -------------------- | ----------------------------- |
| `.env.production`       | Production env vars  | Before building               |
| `deploy.sh`             | Automated deployment | Quick deployments             |
| `nginx.conf`            | Nginx configuration  | Nginx server                  |
| `.htaccess`             | Apache configuration | Apache server                 |
| `index.production.html` | Clean CSP for prod   | Alternative to server headers |

---

## Additional Resources

- [CSP.md](./CSP.md) - Detailed CSP configuration
- [CORS.md](./CORS.md) - Backend CORS setup
- [README.md](./README.md) - General documentation
- [HEALTH_CHECK.md](./HEALTH_CHECK.md) - Health monitoring

---

**Need Help?**

Create an issue or contact the PWA Development Team.
