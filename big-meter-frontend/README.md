# Big Meter Frontend

ระบบแสดงผลผู้ใช้น้ำรายใหญ่ - Provincial Waterworks Authority (PWA) Water Usage Dashboard

A React-based web application for monitoring and analyzing large-scale water usage data across different branches of the Provincial Waterworks Authority of Thailand.

## Features

- 📊 **Water Usage Dashboard** - Monitor consumption patterns across branches
- 📈 **Historical Trends** - Interactive sparkline charts showing usage over time
- 🔍 **Advanced Filtering** - Filter by branch, time period, and usage decrease threshold
- 📥 **Excel Export** - Export filtered reports with full historical data
- 🔐 **PWA Intranet Auth** - Integration with existing PWA authentication system
- 📱 **Responsive Design** - Works seamlessly on desktop and mobile devices
- 🇹🇭 **Thai Language** - Full Thai localization with Buddhist calendar

## Tech Stack

- **React 19** - Modern UI framework
- **TypeScript** - Type-safe development
- **Vite** - Lightning-fast build tool
- **TanStack Query** - Powerful data fetching and caching
- **Tailwind CSS 4** - Utility-first styling
- **visx** - Data visualization library
- **pnpm** - Fast, disk-efficient package manager

## Prerequisites

- **Node.js** 18+ (recommended: 20 LTS)
- **pnpm** 10+ (install via `npm install -g pnpm`)
- Access to Big Meter backend API
- PWA Intranet credentials for authentication

## Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd big-meter-frontend
```

### 2. Install Dependencies

```bash
pnpm install
```

### 3. Configure Environment Variables

Copy the example environment file:

```bash
cp .env.example .env.local
```

Edit `.env.local` with your configuration:

```env
# Backend API URL
VITE_API_BASE_URL=http://localhost:8089

# Optional: Override PWA login endpoint
# VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php
```

### 4. Start Development Server

```bash
pnpm dev
```

The application will open automatically at `http://localhost:5173`

## Available Scripts

```bash
pnpm dev         # Start development server with hot reload
pnpm build       # Build for production (includes TypeScript check)
pnpm preview     # Preview production build locally
pnpm lint        # Run linter (configure ESLint first)
pnpm format      # Format code with Prettier
```

## Building for Production

### 1. Set Production Environment Variables

Create `.env.production`:

```env
VITE_API_BASE_URL=https://api.yourdomain.pwa.co.th
VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php
```

### 2. Build the Application

```bash
pnpm build
```

This will:

- Run TypeScript type checking
- Build optimized production bundle
- Output to `dist/` directory

### 3. Test the Production Build

```bash
pnpm preview
```

Visit `http://localhost:4173` to test the production build locally.

## Deployment

**📖 See the complete [Deployment Guide →](./DEPLOYMENT.md)** for detailed step-by-step instructions.

### Quick Overview

### Option 1: Static Hosting (Nginx, Apache)

1. Build the application:

   ```bash
   pnpm build
   ```

2. Deploy the `dist/` folder to your web server

3. Configure your web server for SPA routing:

   **Nginx example:**

   ```nginx
   server {
       listen 80;
       server_name yourdomain.pwa.co.th;
       root /var/www/big-meter-frontend/dist;
       index index.html;

       location / {
           try_files $uri $uri/ /index.html;
       }

       # Optional: Gzip compression
       gzip on;
       gzip_types text/css application/javascript application/json;
   }
   ```

   **Apache example (.htaccess):**

   ```apache
   <IfModule mod_rewrite.c>
       RewriteEngine On
       RewriteBase /
       RewriteRule ^index\.html$ - [L]
       RewriteCond %{REQUEST_FILENAME} !-f
       RewriteCond %{REQUEST_FILENAME} !-d
       RewriteRule . /index.html [L]
   </IfModule>
   ```

### Option 2: Docker Deployment

1. Create `Dockerfile`:

   ```dockerfile
   FROM node:20-alpine AS builder
   WORKDIR /app
   COPY package.json pnpm-lock.yaml ./
   RUN npm install -g pnpm && pnpm install --frozen-lockfile
   COPY . .
   ARG VITE_API_BASE_URL
   ARG VITE_LOGIN_API
   ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
   ENV VITE_LOGIN_API=$VITE_LOGIN_API
   RUN pnpm build

   FROM nginx:alpine
   COPY --from=builder /app/dist /usr/share/nginx/html
   COPY nginx.conf /etc/nginx/conf.d/default.conf
   EXPOSE 80
   CMD ["nginx", "-g", "daemon off;"]
   ```

2. Create `nginx.conf`:

   ```nginx
   server {
       listen 80;
       root /usr/share/nginx/html;
       index index.html;

       location / {
           try_files $uri $uri/ /index.html;
       }

       gzip on;
       gzip_types text/css application/javascript application/json;
   }
   ```

3. Build and run:

   ```bash
   docker build \
     --build-arg VITE_API_BASE_URL=https://api.yourdomain.pwa.co.th \
     --build-arg VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php \
     -t big-meter-frontend .

   docker run -p 80:80 big-meter-frontend
   ```

### Option 3: Vercel/Netlify

1. Connect your repository to Vercel or Netlify

2. Configure build settings:
   - **Build command:** `pnpm build`
   - **Output directory:** `dist`
   - **Install command:** `pnpm install`

3. Add environment variables in the platform dashboard:
   - `VITE_API_BASE_URL`
   - `VITE_LOGIN_API`

4. Deploy automatically on push to main branch

## Environment Variables

| Variable            | Description                 | Required | Default                                                  |
| ------------------- | --------------------------- | -------- | -------------------------------------------------------- |
| `VITE_API_BASE_URL` | Backend API base URL        | Yes      | `http://localhost:8089`                                  |
| `VITE_LOGIN_API`    | PWA Intranet login endpoint | No       | `https://intranet.pwa.co.th/login/webservice_login6.php` |

**Important:** All environment variables must be prefixed with `VITE_` to be exposed to the client.

## API Requirements

The application expects the following backend endpoints:

### Authentication

- `POST /auth/login` - PWA Intranet authentication proxy

### Data Endpoints

- `GET /api/v1/branches` - List of branches
- `GET /api/v1/custcodes` - Customer metadata
- `GET /api/v1/details` - Water usage details

### Health Check (Recommended)

- `GET /api/health` - Health check endpoint for monitoring and uptime verification

**📖 See the complete [Health Check Guide →](./HEALTH_CHECK.md)** for:

- Backend implementation examples (Go, Node.js, Python)
- Kubernetes/Docker integration
- Monitoring setup and best practices
- Frontend health check utility usage

### CORS Configuration

The backend **must** be configured to allow CORS from your frontend domain. This is critical for the application to function.

**Quick example:**

```javascript
// Express.js
app.use(
  cors({
    origin: "https://yourdomain.pwa.co.th",
    credentials: true,
  }),
);
```

**📖 For detailed CORS configuration including:**

- Framework-specific examples (Go, Node.js, Python, Nginx)
- Security best practices
- Testing procedures
- Troubleshooting common issues

**See the complete [CORS Configuration Guide →](./CORS.md)**

## Security Considerations

### Production Checklist

- ✅ Use HTTPS for all production deployments
- ✅ Configure proper CORS on backend (see [CORS.md](./CORS.md))
- ✅ Configure Content Security Policy (see [CSP.md](./CSP.md))
- ✅ Set secure HTTP headers (HSTS, X-Frame-Options, etc.)
- ✅ Implement rate limiting on backend
- ✅ Regular security updates for dependencies
- ✅ Monitor for vulnerabilities: `pnpm audit`

### Content Security Policy

The application includes a CSP meta tag in `index.html` that allows localhost connections for development. **Server headers are strongly recommended** for production.

**⚠️ Development Note:** The CSP in `index.html` includes `http://localhost:8089` in `connect-src` for local API access. Remove this in production and use environment-specific server headers instead.

**📖 See the complete [CSP Configuration Guide →](./CSP.md)** for:

- Production-ready CSP headers
- Server configuration examples (Nginx, Apache, Node.js)
- Environment-specific policies (dev, staging, production)
- Testing and troubleshooting procedures
- Self-hosted fonts CSP adjustments

### Recommended Security Headers

```nginx
# Content Security Policy (see CSP.md for details)
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'" always;

# Prevent clickjacking
add_header X-Frame-Options "DENY" always;

# Prevent MIME sniffing
add_header X-Content-Type-Options "nosniff" always;

# XSS Protection (legacy browsers)
add_header X-XSS-Protection "1; mode=block" always;

# Referrer Policy
add_header Referrer-Policy "strict-origin-when-cross-origin" always;

# HSTS (HTTPS only)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

# Permissions Policy
add_header Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=()" always;
```

## Troubleshooting

### Build Fails

**Issue:** TypeScript errors during build

```bash
pnpm build
# Error: Type 'X' is not assignable to type 'Y'
```

**Solution:** Fix TypeScript errors or temporarily disable strict checks:

```bash
tsc -b --noEmit false && vite build
```

### Login Not Working

**Issue:** CORS errors when logging in

**Solution:** Ensure backend CORS is configured and `VITE_LOGIN_API` is correct

### Data Not Loading

**Issue:** API requests fail in production

**Solution:**

1. Check `VITE_API_BASE_URL` is set correctly
2. Verify backend is accessible from production domain
3. Check browser console for CORS errors
4. Verify API endpoints are responding

### White Screen After Deployment

**Issue:** Blank page in production

**Solution:**

1. Check browser console for errors
2. Verify build output in `dist/` directory
3. Ensure server is configured for SPA routing
4. Check that environment variables were set during build

## Development

### Project Structure

```
big-meter-frontend/
├── src/
│   ├── api/              # API client layer
│   ├── lib/              # Utilities and hooks
│   ├── screens/          # Page components
│   ├── styles/           # Global styles
│   ├── App.tsx           # Root component
│   ├── main.tsx          # Entry point
│   └── routes.tsx        # Route configuration
├── dist/                 # Build output (generated)
├── .env.example          # Environment template
├── vite.config.ts        # Vite configuration
├── tsconfig.json         # TypeScript config
└── package.json          # Dependencies
```

### Adding New Features

1. Create feature branch
2. Add types in `src/api/`
3. Create components in `src/screens/`
4. Update routes in `src/routes.tsx`
5. Test locally
6. Create pull request

### Code Quality

```bash
# Lint code
pnpm lint

# Format code
pnpm format

# Type check
pnpm build

# Health check API
node -e "import('./src/lib/healthCheck.ts').then(m => m.checkApiHealth().then(console.log))"

# Update dependencies
pnpm update --latest
```

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Mobile browsers (iOS Safari 14+, Chrome Android)

**Note:** Internet Explorer is not supported.

## Performance

### Bundle Size

- Main bundle: ~422 KB (gzipped: ~132 KB)
- XLSX export (lazy): ~284 KB (gzipped: ~96 KB)
- CSS: ~27 KB (gzipped: ~6 KB)

### Optimization Tips

1. **Code splitting** - XLSX export is already lazy-loaded
2. **Image optimization** - Use WebP format for images
3. **CDN** - Serve static assets from CDN
4. **Caching** - Configure aggressive caching for build assets
5. **Compression** - Enable gzip/brotli on server

## License

Provincial Waterworks Authority - Internal Use Only

## Support

For issues and questions:

- Create an issue in the repository
- Contact the PWA Development Team
- Email: support@pwa.co.th

## Changelog

See [CLAUDE.md](./CLAUDE.md) for detailed project documentation and recent changes.

---

**Built with ❤️ by PWA Development Team**
