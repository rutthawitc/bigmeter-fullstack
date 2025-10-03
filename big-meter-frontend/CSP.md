# Content Security Policy (CSP) Configuration

## Overview

Content Security Policy (CSP) is a security layer that helps detect and mitigate certain types of attacks, including Cross-Site Scripting (XSS) and data injection attacks. This document provides CSP configuration for the Big Meter Frontend application.

## Recommended CSP Policy

### Production CSP Header

```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
```

### Formatted for Readability

```
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
img-src 'self' data: https:;
connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th;
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
```

### Development CSP (More Permissive)

```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' ws://localhost:5173 http://localhost:8089 https://intranet.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
```

## Directive Breakdown

### `default-src 'self'`

**Purpose:** Fallback for all resource types not explicitly defined.

**Allows:** Resources from same origin only.

### `script-src 'self'`

**Purpose:** Controls JavaScript execution sources.

**Allows:**

- Scripts from same origin
- Vite bundles built by the application

**Blocks:**

- Inline scripts (prevents XSS)
- External CDN scripts
- `eval()` and similar functions

**Why 'unsafe-eval' in development:**

- Vite HMR (Hot Module Replacement) uses eval
- Remove in production

### `style-src 'self' 'unsafe-inline' https://fonts.googleapis.com`

**Purpose:** Controls CSS sources.

**Allows:**

- Stylesheets from same origin
- Inline styles (required for Tailwind CSS)
- Google Fonts CSS

**Note:** `'unsafe-inline'` is needed because:

- Tailwind CSS generates inline styles
- React components may use inline styles
- Removing would break styling

### `font-src 'self' https://fonts.gstatic.com`

**Purpose:** Controls font file sources.

**Allows:**

- Self-hosted fonts
- Google Fonts (fonts.gstatic.com)

**Why Google Fonts:**

- Application uses Sarabun from Google Fonts
- If self-hosting fonts, can remove `https://fonts.gstatic.com`

### `img-src 'self' data: https:`

**Purpose:** Controls image sources.

**Allows:**

- Images from same origin
- Data URIs (base64 encoded images)
- Any HTTPS image source

**Why HTTPS wildcard:**

- User avatars might be from external sources
- Future-proofing for CDN images
- More restrictive option: `'self' data:` only

### `connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th`

**Purpose:** Controls XHR, fetch, WebSocket connections.

**Allows:**

- API calls to same origin
- PWA Intranet login endpoint
- Backend API endpoint

**Important:** Update `https://api.pwa.co.th` to your actual API domain.

**Development additions:**

- `ws://localhost:5173` - Vite HMR WebSocket
- `http://localhost:8089` - Local API server

### `frame-ancestors 'none'`

**Purpose:** Controls where the page can be embedded.

**Prevents:**

- Clickjacking attacks
- Embedding in iframes

**Equivalent to:** `X-Frame-Options: DENY`

### `base-uri 'self'`

**Purpose:** Restricts URLs that can appear in `<base>` tag.

**Prevents:** Base tag injection attacks.

### `form-action 'self'`

**Purpose:** Restricts where forms can submit.

**Allows:** Form submissions to same origin only.

**Note:** Login form submits to `/auth/login` which is proxied to PWA Intranet.

## Implementation Methods

### Method 1: Server Headers (Recommended)

#### Nginx

```nginx
server {
    listen 80;
    server_name big-meter.pwa.co.th;

    # CSP Header
    add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'" always;

    # Additional security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

    location / {
        root /var/www/big-meter-frontend/dist;
        try_files $uri $uri/ /index.html;
    }
}
```

#### Apache (.htaccess)

```apache
<IfModule mod_headers.c>
    Header always set Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

    # Additional security headers
    Header always set X-Frame-Options "DENY"
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Referrer-Policy "strict-origin-when-cross-origin"
</IfModule>
```

#### Node.js/Express

```javascript
const helmet = require("helmet");

app.use(
  helmet.contentSecurityPolicy({
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
      fontSrc: ["'self'", "https://fonts.gstatic.com"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: [
        "'self'",
        "https://intranet.pwa.co.th",
        "https://api.pwa.co.th",
      ],
      frameAncestors: ["'none'"],
      baseUri: ["'self'"],
      formAction: ["'self'"],
    },
  }),
);
```

#### Cloudflare Workers

```javascript
addEventListener("fetch", (event) => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const response = await fetch(request);
  const newResponse = new Response(response.body, response);

  newResponse.headers.set(
    "Content-Security-Policy",
    "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
  );

  return newResponse;
}
```

### Method 2: HTML Meta Tag (Fallback)

**Note:** Meta tags are less powerful than HTTP headers. Some directives (like `frame-ancestors`) only work in headers.

Add to `index.html`:

```html
<head>
  <meta
    http-equiv="Content-Security-Policy"
    content="default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th; base-uri 'self'; form-action 'self'"
  />
</head>
```

**Limitations:**

- Cannot use `frame-ancestors`
- Cannot use `report-uri` / `report-to`
- Less flexible than headers
- Not recommended for production

### Method 3: Vite Plugin (Development)

Create `vite-plugin-csp.ts`:

```typescript
import type { Plugin } from "vite";

export function cspPlugin(): Plugin {
  return {
    name: "vite-plugin-csp",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        res.setHeader(
          "Content-Security-Policy",
          "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' ws://localhost:5173 http://localhost:8089 https://intranet.pwa.co.th; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
        );
        next();
      });
    },
  };
}
```

Use in `vite.config.ts`:

```typescript
import { cspPlugin } from "./vite-plugin-csp";

export default defineConfig({
  plugins: [react(), tailwind(), cspPlugin()],
});
```

## Testing CSP

### 1. Browser DevTools

1. Open application in browser
2. Open DevTools (F12)
3. Go to Console tab
4. Look for CSP violations (red errors)
5. Check Network tab for blocked requests

**Example violation:**

```
Refused to load the script 'https://evil.com/script.js' because it violates
the following Content Security Policy directive: "script-src 'self'".
```

### 2. CSP Evaluator

Use Google's CSP Evaluator: https://csp-evaluator.withgoogle.com/

Paste your CSP and check for:

- Syntax errors
- Security weaknesses
- Best practice violations

### 3. Report-Only Mode

Test CSP without blocking resources:

```
Content-Security-Policy-Report-Only: default-src 'self'; ...
```

Violations are reported to console but resources still load.

### 4. CSP Reports

Set up reporting endpoint:

```
Content-Security-Policy: default-src 'self'; ...; report-uri https://your-domain.pwa.co.th/csp-report
```

Backend endpoint to receive reports:

```javascript
app.post(
  "/csp-report",
  express.json({ type: "application/csp-report" }),
  (req, res) => {
    console.log("CSP Violation:", req.body);
    res.status(204).end();
  },
);
```

## Common Issues and Solutions

### Issue: Vite HMR Not Working

**Error:** `Refused to evaluate a string as JavaScript because 'unsafe-eval'...`

**Solution:** Add `'unsafe-eval'` to `script-src` in development only.

```
script-src 'self' 'unsafe-eval';  // Development only
```

### Issue: Inline Styles Blocked

**Error:** `Refused to apply inline style because it violates CSP directive...`

**Solution:** Add `'unsafe-inline'` to `style-src` (required for Tailwind).

```
style-src 'self' 'unsafe-inline';
```

### Issue: Google Fonts Blocked

**Error:** `Refused to load the stylesheet 'https://fonts.googleapis.com/...'`

**Solution:** Add both Google Fonts domains:

```
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
```

### Issue: API Calls Blocked

**Error:** `Refused to connect to 'https://api.pwa.co.th' because it violates...`

**Solution:** Add API domain to `connect-src`:

```
connect-src 'self' https://api.pwa.co.th;
```

### Issue: Data URI Images Blocked

**Error:** `Refused to load the image 'data:image/png;base64,...'`

**Solution:** Add `data:` to `img-src`:

```
img-src 'self' data:;
```

## Self-Hosted Fonts CSP

If switching to self-hosted fonts (see [FONTS.md](./FONTS.md)):

**Remove:**

```
style-src ... https://fonts.googleapis.com
font-src ... https://fonts.gstatic.com
```

**Result:**

```
style-src 'self' 'unsafe-inline';
font-src 'self';
```

## Strict CSP (Future Enhancement)

For maximum security, consider implementing nonce-based CSP:

### 1. Generate Nonce Server-Side

```javascript
const crypto = require("crypto");
const nonce = crypto.randomBytes(16).toString("base64");

res.setHeader(
  "Content-Security-Policy",
  `script-src 'nonce-${nonce}'; style-src 'self' 'unsafe-inline'`,
);
```

### 2. Add Nonce to Scripts

```html
<script nonce="${nonce}" src="/assets/main.js"></script>
```

### 3. Remove 'unsafe-inline'

```
style-src 'self';  // No 'unsafe-inline'
```

**Challenges:**

- Requires server-side rendering or template engine
- Complex with SPA + build tools
- Not compatible with Vite's default setup
- Tailwind generates inline styles

## Security Headers Bundle

Complete security headers for production:

```nginx
# Content Security Policy
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
add_header Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()" always;
```

## Environment-Specific CSP

### Development

```
default-src 'self';
script-src 'self' 'unsafe-eval';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
img-src 'self' data: https:;
connect-src 'self' ws://localhost:5173 http://localhost:8089 https://intranet.pwa.co.th;
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
```

### Staging

```
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
img-src 'self' data: https:;
connect-src 'self' https://intranet.pwa.co.th https://staging-api.pwa.co.th;
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
report-uri https://staging.pwa.co.th/csp-report;
```

### Production

```
default-src 'self';
script-src 'self';
style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
font-src 'self' https://fonts.gstatic.com;
img-src 'self' data: https:;
connect-src 'self' https://intranet.pwa.co.th https://api.pwa.co.th;
frame-ancestors 'none';
base-uri 'self';
form-action 'self';
upgrade-insecure-requests;
block-all-mixed-content;
```

## Monitoring

Set up CSP violation monitoring:

```javascript
// Add to backend
app.post("/csp-report", (req, res) => {
  const violation = req.body["csp-report"];

  // Log to monitoring service
  logger.warn("CSP Violation", {
    documentUri: violation["document-uri"],
    violatedDirective: violation["violated-directive"],
    blockedUri: violation["blocked-uri"],
    sourceFile: violation["source-file"],
    lineNumber: violation["line-number"],
  });

  // Send to error tracking (e.g., Sentry)
  // Sentry.captureMessage('CSP Violation', { extra: violation });

  res.status(204).end();
});
```

## Checklist

Before deploying with CSP:

- [ ] CSP header configured on web server
- [ ] Tested with browser DevTools (no violations)
- [ ] Tested all application features
- [ ] API calls work correctly
- [ ] Google Fonts load properly
- [ ] Images display correctly
- [ ] Forms submit successfully
- [ ] No console errors
- [ ] Report-URI configured (optional)
- [ ] Monitoring setup (optional)

## Further Reading

- [MDN: Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [CSP Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
- [Google CSP Guide](https://web.dev/csp/)
- [CSP Evaluator](https://csp-evaluator.withgoogle.com/)

---

**Last Updated:** 2025-10-03
**Maintained by:** PWA Development Team
