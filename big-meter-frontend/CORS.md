# CORS Configuration Guide

## Overview

The Big Meter Frontend application requires proper CORS (Cross-Origin Resource Sharing) configuration on the backend API to function correctly. This document outlines the CORS requirements and provides configuration examples for various backend frameworks.

## Why CORS is Required

The frontend application runs on a different domain than the backend API:

- **Frontend:** `https://big-meter.pwa.co.th` (example)
- **Backend API:** `https://api.big-meter.pwa.co.th` (example)
- **PWA Intranet:** `https://intranet.pwa.co.th`

Modern browsers enforce the Same-Origin Policy, which blocks requests to different domains unless the server explicitly allows them via CORS headers.

## Required CORS Headers

### Minimum Required Headers

```
Access-Control-Allow-Origin: https://big-meter.pwa.co.th
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

### For Development (Less Restrictive)

```
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 3600
```

### Multiple Origins (Development + Production)

If your backend needs to support multiple origins (e.g., localhost for development and production domain):

```
Access-Control-Allow-Origin: <dynamically set based on request origin>
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

## Backend Framework Examples

### Go (Golang) with Gin

```go
package main

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    router := gin.Default()

    // CORS configuration
    config := cors.Config{
        AllowOrigins:     []string{"https://big-meter.pwa.co.th"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }

    // For development, add localhost
    if gin.Mode() == gin.DebugMode {
        config.AllowOrigins = append(config.AllowOrigins, "http://localhost:5173")
    }

    router.Use(cors.New(config))

    // Your routes here
    router.GET("/api/v1/branches", getBranches)
    router.GET("/api/v1/custcodes", getCustCodes)
    router.GET("/api/v1/details", getDetails)

    router.Run(":8089")
}
```

### Go (Golang) with Standard Library

```go
package main

import (
    "net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        // Allow specific origins
        allowedOrigins := map[string]bool{
            "https://big-meter.pwa.co.th": true,
            "http://localhost:5173":       true, // for development
        }

        if allowedOrigins[origin] {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Max-Age", "86400")

        // Handle preflight requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()

    // Your routes
    mux.HandleFunc("/api/v1/branches", getBranches)
    mux.HandleFunc("/api/v1/custcodes", getCustCodes)
    mux.HandleFunc("/api/v1/details", getDetails)

    // Wrap with CORS middleware
    handler := corsMiddleware(mux)

    http.ListenAndServe(":8089", handler)
}
```

### Node.js with Express

```javascript
const express = require("express");
const cors = require("cors");
const app = express();

// CORS configuration
const corsOptions = {
  origin: function (origin, callback) {
    const allowedOrigins = [
      "https://big-meter.pwa.co.th",
      "http://localhost:5173", // for development
    ];

    if (!origin || allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error("Not allowed by CORS"));
    }
  },
  credentials: true,
  methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  allowedHeaders: ["Content-Type", "Authorization"],
  maxAge: 86400,
};

app.use(cors(corsOptions));

// Your routes
app.get("/api/v1/branches", getBranches);
app.get("/api/v1/custcodes", getCustCodes);
app.get("/api/v1/details", getDetails);

app.listen(8089);
```

### Python with Flask

```python
from flask import Flask
from flask_cors import CORS

app = Flask(__name__)

# CORS configuration
cors_config = {
    "origins": [
        "https://big-meter.pwa.co.th",
        "http://localhost:5173"  # for development
    ],
    "methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
    "allow_headers": ["Content-Type", "Authorization"],
    "supports_credentials": True,
    "max_age": 86400
}

CORS(app, resources={r"/api/*": cors_config})

# Your routes
@app.route('/api/v1/branches')
def get_branches():
    pass

@app.route('/api/v1/custcodes')
def get_custcodes():
    pass

@app.route('/api/v1/details')
def get_details():
    pass

if __name__ == '__main__':
    app.run(port=8089)
```

## Nginx Reverse Proxy Configuration

If you're using Nginx as a reverse proxy in front of your backend:

```nginx
server {
    listen 80;
    server_name api.big-meter.pwa.co.th;

    location / {
        # Proxy to backend
        proxy_pass http://localhost:8089;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # CORS headers
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' 'https://big-meter.pwa.co.th' always;
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
            add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization' always;
            add_header 'Access-Control-Max-Age' 86400 always;
            add_header 'Content-Length' 0;
            add_header 'Content-Type' 'text/plain charset=UTF-8';
            return 204;
        }

        add_header 'Access-Control-Allow-Origin' 'https://big-meter.pwa.co.th' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
        add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization' always;
        add_header 'Access-Control-Allow-Credentials' 'true' always;
    }
}
```

## Testing CORS Configuration

### 1. Using curl

Test preflight request:

```bash
curl -X OPTIONS \
  -H "Origin: https://big-meter.pwa.co.th" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -I https://api.big-meter.pwa.co.th/api/v1/branches
```

Expected response headers:

```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://big-meter.pwa.co.th
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

### 2. Using Browser DevTools

1. Open the frontend application in Chrome/Firefox
2. Open Developer Tools (F12)
3. Go to Network tab
4. Perform an action that triggers an API request
5. Check the request headers and response headers
6. Look for `Access-Control-Allow-*` headers in the response

### 3. Common CORS Errors

**Error:** `No 'Access-Control-Allow-Origin' header is present`

- **Cause:** Backend is not sending CORS headers
- **Fix:** Add CORS middleware to your backend

**Error:** `The 'Access-Control-Allow-Origin' header contains multiple values`

- **Cause:** Multiple CORS configurations are conflicting
- **Fix:** Ensure only one CORS configuration is active (either backend or proxy, not both)

**Error:** `Credentials flag is 'true', but 'Access-Control-Allow-Credentials' is not`

- **Cause:** Frontend sends credentials but backend doesn't allow them
- **Fix:** Add `Access-Control-Allow-Credentials: true` header

**Error:** `Method PUT is not allowed by Access-Control-Allow-Methods`

- **Cause:** Backend doesn't include PUT in allowed methods
- **Fix:** Add PUT to `Access-Control-Allow-Methods`

## Security Considerations

### Production Best Practices

1. **Never use wildcard (`*`) in production:**

   ```
   ❌ Access-Control-Allow-Origin: *
   ✅ Access-Control-Allow-Origin: https://big-meter.pwa.co.th
   ```

2. **Validate origins dynamically:**

   ```go
   allowedOrigins := []string{
       "https://big-meter.pwa.co.th",
       "https://staging.big-meter.pwa.co.th",
   }
   ```

3. **Use HTTPS in production:**
   - Always use `https://` for production origins
   - Never allow `http://` origins in production (except localhost for dev)

4. **Set appropriate Max-Age:**
   - Development: 3600 seconds (1 hour)
   - Production: 86400 seconds (24 hours)

5. **Limit allowed headers:**
   - Only include headers your API actually needs
   - Don't allow all headers with `*`

6. **Handle credentials carefully:**
   - Only set `Access-Control-Allow-Credentials: true` if you're using cookies/auth
   - When using credentials, you cannot use wildcard origins

## Environment-Specific Configuration

### Development Environment

```
Allowed Origins: http://localhost:5173
Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
Max-Age: 3600
Credentials: false (unless testing auth)
```

### Staging Environment

```
Allowed Origins: https://staging.big-meter.pwa.co.th
Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
Max-Age: 86400
Credentials: true
```

### Production Environment

```
Allowed Origins: https://big-meter.pwa.co.th
Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
Max-Age: 86400
Credentials: true
```

## PWA Intranet Login Endpoint

The login endpoint at `https://intranet.pwa.co.th/login/webservice_login6.php` must also have CORS configured to accept requests from your frontend domain.

If you cannot modify the PWA Intranet server, use the Vite dev server proxy (already configured):

```typescript
// vite.config.ts - already configured
server: {
  proxy: {
    '/auth/login': {
      target: 'https://intranet.pwa.co.th',
      changeOrigin: true,
      secure: false,
      rewrite: () => '/login/webservice_login6.php',
    },
  },
}
```

For production, you'll need either:

1. Backend proxy endpoint that forwards to PWA Intranet
2. CORS configuration on PWA Intranet server
3. Same-domain deployment (frontend and backend on same domain with path-based routing)

## Troubleshooting Checklist

- [ ] Backend is sending `Access-Control-Allow-Origin` header
- [ ] Origin matches exactly (including protocol: `https://` vs `http://`)
- [ ] All required methods are in `Access-Control-Allow-Methods`
- [ ] All required headers are in `Access-Control-Allow-Headers`
- [ ] Preflight OPTIONS requests return 200/204 status code
- [ ] If using credentials, `Access-Control-Allow-Credentials: true` is set
- [ ] If using credentials, origin is not `*`
- [ ] No conflicting CORS configurations (check both backend and proxy)
- [ ] Caching isn't causing issues (check `Max-Age` or clear cache)

## Further Reading

- [MDN Web Docs: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [CORS Specification](https://fetch.spec.whatwg.org/#http-cors-protocol)
- [Enable CORS](https://enable-cors.org/)

## Support

If you encounter CORS issues:

1. Check browser console for specific error messages
2. Verify backend CORS configuration using curl tests
3. Check network tab in browser DevTools
4. Ensure environment variables are set correctly
5. Contact backend team if PWA Intranet CORS needs configuration

---

**Last Updated:** 2025-10-03
**Maintained by:** PWA Development Team
