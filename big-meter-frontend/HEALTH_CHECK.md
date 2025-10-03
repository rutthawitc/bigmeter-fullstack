# Health Check Documentation

## Overview

This document describes the health check functionality for the Big Meter Frontend application, including how to verify API connectivity, implement backend health endpoints, and monitor system status.

## Frontend Health Check Utility

The application includes a built-in health check utility at `src/lib/healthCheck.ts` that can verify API endpoint availability.

### Basic Usage

```typescript
import { checkApiHealth } from "./lib/healthCheck";

// Check default health endpoint
const result = await checkApiHealth();
console.log(result);
// {
//   status: "healthy",
//   message: "API is healthy and responding",
//   timestamp: "2025-10-03T10:00:00.000Z",
//   latency: 45,
//   details: {
//     endpoint: "https://api.pwa.co.th/api/health",
//     statusCode: 200
//   }
// }
```

### Advanced Usage

```typescript
import {
  checkApiHealth,
  checkMultipleEndpoints,
  getOverallHealth,
  formatHealthResult,
} from "./lib/healthCheck";

// Check a specific endpoint with custom timeout
const result = await checkApiHealth("/api/v1/branches", 3000);

// Check multiple endpoints
const results = await checkMultipleEndpoints([
  "/api/health",
  "/api/v1/branches",
  "/api/v1/custcodes",
  "/api/v1/details",
]);

// Get overall health status
const overallStatus = getOverallHealth(results);
console.log(`Overall: ${overallStatus}`); // "healthy" | "unhealthy" | "unknown"

// Format for display
console.log(formatHealthResult(result));
```

### API Reference

#### `checkApiHealth(endpoint?, timeout?)`

Performs a health check on the specified API endpoint.

**Parameters:**

- `endpoint` (string, optional): API endpoint to check. Default: `/api/health`
- `timeout` (number, optional): Request timeout in milliseconds. Default: `5000`

**Returns:** `Promise<HealthCheckResult>`

**Example:**

```typescript
const result = await checkApiHealth("/api/v1/branches", 10000);
```

#### `checkMultipleEndpoints(endpoints, timeout?)`

Checks multiple endpoints concurrently.

**Parameters:**

- `endpoints` (string[]): Array of endpoint paths to check
- `timeout` (number, optional): Request timeout in milliseconds. Default: `5000`

**Returns:** `Promise<Record<string, HealthCheckResult>>`

**Example:**

```typescript
const results = await checkMultipleEndpoints(
  ["/api/health", "/api/v1/branches"],
  3000,
);
```

#### `getOverallHealth(results)`

Determines overall health status from multiple endpoint results.

**Parameters:**

- `results` (Record<string, HealthCheckResult>): Results from `checkMultipleEndpoints`

**Returns:** `HealthCheckStatus` - "healthy" | "unhealthy" | "unknown"

#### `isApiConfigured()`

Checks if the API base URL is properly configured.

**Returns:** `boolean`

**Example:**

```typescript
if (!isApiConfigured()) {
  console.error("API base URL not configured!");
}
```

## Backend Health Endpoint Implementation

### Recommended Health Endpoint

Create a `/api/health` endpoint on your backend that returns the following response:

```json
{
  "status": "healthy",
  "timestamp": "2025-10-03T10:00:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "cache": "healthy",
    "externalServices": "healthy"
  }
}
```

### Implementation Examples

#### Go (Golang) with Gin

```go
package main

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

type HealthResponse struct {
    Status    string            `json:"status"`
    Timestamp string            `json:"timestamp"`
    Version   string            `json:"version"`
    Checks    map[string]string `json:"checks"`
}

func healthCheck(c *gin.Context) {
    // Perform actual health checks here
    dbHealthy := checkDatabase()
    cacheHealthy := checkCache()

    status := "healthy"
    if !dbHealthy || !cacheHealthy {
        status = "unhealthy"
    }

    response := HealthResponse{
        Status:    status,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Version:   "1.0.0",
        Checks: map[string]string{
            "database": boolToHealth(dbHealthy),
            "cache":    boolToHealth(cacheHealthy),
        },
    }

    statusCode := http.StatusOK
    if status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    c.JSON(statusCode, response)
}

func boolToHealth(healthy bool) string {
    if healthy {
        return "healthy"
    }
    return "unhealthy"
}

func checkDatabase() bool {
    // Implement database connectivity check
    // Example: db.Ping()
    return true
}

func checkCache() bool {
    // Implement cache connectivity check
    return true
}

func main() {
    router := gin.Default()
    router.GET("/api/health", healthCheck)
    router.Run(":8089")
}
```

#### Node.js with Express

```javascript
const express = require("express");
const app = express();

app.get("/api/health", async (req, res) => {
  try {
    // Perform health checks
    const dbHealthy = await checkDatabase();
    const cacheHealthy = await checkCache();

    const status = dbHealthy && cacheHealthy ? "healthy" : "unhealthy";
    const statusCode = status === "healthy" ? 200 : 503;

    res.status(statusCode).json({
      status,
      timestamp: new Date().toISOString(),
      version: "1.0.0",
      checks: {
        database: dbHealthy ? "healthy" : "unhealthy",
        cache: cacheHealthy ? "healthy" : "unhealthy",
      },
    });
  } catch (error) {
    res.status(503).json({
      status: "unhealthy",
      timestamp: new Date().toISOString(),
      error: error.message,
    });
  }
});

async function checkDatabase() {
  // Implement database check
  return true;
}

async function checkCache() {
  // Implement cache check
  return true;
}

app.listen(8089);
```

#### Python with Flask

```python
from flask import Flask, jsonify
from datetime import datetime
import sys

app = Flask(__name__)

@app.route('/api/health')
def health_check():
    try:
        db_healthy = check_database()
        cache_healthy = check_cache()

        status = 'healthy' if db_healthy and cache_healthy else 'unhealthy'
        status_code = 200 if status == 'healthy' else 503

        return jsonify({
            'status': status,
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'version': '1.0.0',
            'checks': {
                'database': 'healthy' if db_healthy else 'unhealthy',
                'cache': 'healthy' if cache_healthy else 'unhealthy'
            }
        }), status_code
    except Exception as e:
        return jsonify({
            'status': 'unhealthy',
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'error': str(e)
        }), 503

def check_database():
    # Implement database check
    return True

def check_cache():
    # Implement cache check
    return True

if __name__ == '__main__':
    app.run(port=8089)
```

## Health Check Best Practices

### 1. Endpoint Design

- **Path:** Use `/api/health` or `/health` for consistency
- **Method:** Use `GET` for health checks
- **Authentication:** Health checks should NOT require authentication
- **CORS:** Health endpoints should be accessible from monitoring tools

### 2. Response Format

**Successful Health Check (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2025-10-03T10:00:00Z",
  "version": "1.0.0"
}
```

**Failed Health Check (503 Service Unavailable):**

```json
{
  "status": "unhealthy",
  "timestamp": "2025-10-03T10:00:00Z",
  "error": "Database connection failed"
}
```

### 3. What to Check

#### Essential Checks

- ✅ Database connectivity
- ✅ Critical external service availability
- ✅ Application startup status

#### Optional Checks

- Cache/Redis connectivity
- File system access
- Memory/CPU usage
- Queue system status

### 4. Performance Considerations

- **Fast Response:** Health checks should complete in <1 second
- **Lightweight:** Don't perform heavy operations
- **Caching:** Cache dependency status for 5-10 seconds
- **Timeout:** Set reasonable timeouts for dependency checks

**Example with caching:**

```typescript
let lastCheck = { status: "unknown", timestamp: 0 };
const CACHE_TTL = 5000; // 5 seconds

app.get("/api/health", (req, res) => {
  const now = Date.now();

  if (now - lastCheck.timestamp < CACHE_TTL) {
    return res.json(lastCheck);
  }

  // Perform actual health check
  const status = performHealthCheck();
  lastCheck = { ...status, timestamp: now };

  res.json(lastCheck);
});
```

### 5. Monitoring Integration

Health endpoints should be compatible with:

- **Kubernetes:** Liveness and readiness probes
- **Docker:** HEALTHCHECK instruction
- **Load Balancers:** Health check configuration
- **Monitoring Tools:** Uptime monitoring (UptimeRobot, Pingdom, etc.)

## Kubernetes Liveness/Readiness Probes

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: big-meter-backend
spec:
  containers:
    - name: api
      image: big-meter-api:latest
      ports:
        - containerPort: 8089
      livenessProbe:
        httpGet:
          path: /api/health
          port: 8089
        initialDelaySeconds: 10
        periodSeconds: 10
        timeoutSeconds: 5
        failureThreshold: 3
      readinessProbe:
        httpGet:
          path: /api/health
          port: 8089
        initialDelaySeconds: 5
        periodSeconds: 5
        timeoutSeconds: 3
        failureThreshold: 2
```

## Docker Health Check

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN npm install
EXPOSE 8089

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD node healthcheck.js || exit 1

CMD ["node", "server.js"]
```

**healthcheck.js:**

```javascript
const http = require("http");

const options = {
  hostname: "localhost",
  port: 8089,
  path: "/api/health",
  timeout: 5000,
};

const req = http.request(options, (res) => {
  if (res.statusCode === 200) {
    process.exit(0);
  } else {
    process.exit(1);
  }
});

req.on("error", () => process.exit(1));
req.on("timeout", () => {
  req.destroy();
  process.exit(1);
});

req.end();
```

## Manual Testing

### Using curl

```bash
# Basic health check
curl -i http://localhost:8089/api/health

# With timeout
curl -i --max-time 5 http://localhost:8089/api/health

# Formatted JSON output
curl -s http://localhost:8089/api/health | jq .
```

### Using Browser

Simply navigate to: `https://api.pwa.co.th/api/health`

### Using Frontend Utility

Open browser console on the application and run:

```javascript
// Import dynamically
const { checkApiHealth } = await import("/src/lib/healthCheck.ts");

// Run health check
const result = await checkApiHealth();
console.log(result);
```

## Monitoring Setup

### Uptime Monitoring Services

Configure services like UptimeRobot or Pingdom to:

- Check `/api/health` endpoint every 1-5 minutes
- Alert on non-200 responses
- Alert on timeout (>5 seconds)
- Track uptime percentage

### Custom Monitoring Script

```bash
#!/bin/bash
# health-monitor.sh

API_URL="https://api.pwa.co.th/api/health"
TIMEOUT=5
LOG_FILE="/var/log/api-health.log"

while true; do
  TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

  # Perform health check
  RESPONSE=$(curl -s -w "\n%{http_code}" --max-time $TIMEOUT "$API_URL")
  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

  if [ "$HTTP_CODE" -eq 200 ]; then
    echo "[$TIMESTAMP] API healthy (HTTP $HTTP_CODE)" >> "$LOG_FILE"
  else
    echo "[$TIMESTAMP] API unhealthy (HTTP $HTTP_CODE)" >> "$LOG_FILE"
    # Send alert (e.g., email, Slack, PagerDuty)
    # send-alert "API health check failed"
  fi

  sleep 60 # Check every minute
done
```

## Troubleshooting

### Health Check Returns 404

- Verify `/api/health` endpoint exists on backend
- Check backend route configuration
- Ensure CORS allows health check requests

### Health Check Times Out

- Backend may be overloaded
- Check database/dependency connectivity
- Review backend logs for errors
- Increase timeout threshold

### Always Returns "Unhealthy"

- Check all dependency connections (database, cache, etc.)
- Review backend health check logic
- Verify environment variables are set
- Check backend logs for specific errors

### CORS Error on Health Check

- Health endpoint needs CORS headers
- Add health endpoint to CORS allowed origins
- See [CORS.md](./CORS.md) for configuration

## Integration with Frontend

The health check utility can be integrated into the frontend application for:

1. **Startup verification** - Check API availability before rendering
2. **Error boundaries** - Display connectivity status in error states
3. **Admin dashboard** - Show real-time system status
4. **Development tools** - Debug API connectivity issues

**Example integration:**

```typescript
// In main.tsx or App.tsx
import { checkApiHealth } from "./lib/healthCheck";

async function verifyApiHealth() {
  const result = await checkApiHealth();

  if (result.status !== "healthy") {
    console.warn("API health check failed:", result);
    // Show warning to user or redirect to maintenance page
  }
}

// Call on app startup
verifyApiHealth();
```

## Further Reading

- [Kubernetes Health Checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [Docker HEALTHCHECK](https://docs.docker.com/engine/reference/builder/#healthcheck)
- [Health Check Pattern](https://microservices.io/patterns/observability/health-check-api.html)

---

**Last Updated:** 2025-10-03
**Maintained by:** PWA Development Team
