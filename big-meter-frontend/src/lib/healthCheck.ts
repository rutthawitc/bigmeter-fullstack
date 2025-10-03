/**
 * Health Check Utility
 * Verifies API endpoint availability and connectivity
 */

export type HealthCheckStatus = "healthy" | "unhealthy" | "unknown";

export type HealthCheckResult = {
  status: HealthCheckStatus;
  message: string;
  timestamp: string;
  latency?: number;
  details?: {
    endpoint?: string;
    statusCode?: number;
    error?: string;
  };
};

/**
 * Check if the API base URL is configured
 */
export function isApiConfigured(): boolean {
  const base = import.meta.env.VITE_API_BASE_URL as string | undefined;
  if (!base) return false;
  if (!/^https?:\/\//i.test(base)) return false;
  return true;
}

/**
 * Perform a health check on the API endpoint
 * @param endpoint - Optional specific endpoint to check (defaults to /api/health or /api/v1/branches)
 * @param timeout - Request timeout in milliseconds (default: 5000)
 */
export async function checkApiHealth(
  endpoint?: string,
  timeout = 5000
): Promise<HealthCheckResult> {
  const timestamp = new Date().toISOString();

  // Check if API is configured
  if (!isApiConfigured()) {
    return {
      status: "unhealthy",
      message: "API base URL not configured",
      timestamp,
      details: {
        error: "VITE_API_BASE_URL environment variable is missing or invalid",
      },
    };
  }

  const apiBase = import.meta.env.VITE_API_BASE_URL as string;
  const healthEndpoint = endpoint || "/api/health";
  const url = `${apiBase}${healthEndpoint}`;

  const startTime = performance.now();

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout);

    const response = await fetch(url, {
      method: "GET",
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
      },
    });

    clearTimeout(timeoutId);

    const endTime = performance.now();
    const latency = Math.round(endTime - startTime);

    if (response.ok) {
      return {
        status: "healthy",
        message: "API is healthy and responding",
        timestamp,
        latency,
        details: {
          endpoint: url,
          statusCode: response.status,
        },
      };
    }

    return {
      status: "unhealthy",
      message: `API returned non-OK status: ${response.status}`,
      timestamp,
      latency,
      details: {
        endpoint: url,
        statusCode: response.status,
      },
    };
  } catch (error) {
    const endTime = performance.now();
    const latency = Math.round(endTime - startTime);

    if (error instanceof Error) {
      if (error.name === "AbortError") {
        return {
          status: "unhealthy",
          message: "API health check timed out",
          timestamp,
          latency,
          details: {
            endpoint: url,
            error: `Request exceeded ${timeout}ms timeout`,
          },
        };
      }

      return {
        status: "unhealthy",
        message: "API health check failed",
        timestamp,
        latency,
        details: {
          endpoint: url,
          error: error.message,
        },
      };
    }

    return {
      status: "unknown",
      message: "Unknown error during health check",
      timestamp,
      latency,
      details: {
        endpoint: url,
        error: String(error),
      },
    };
  }
}

/**
 * Check multiple endpoints for availability
 */
export async function checkMultipleEndpoints(
  endpoints: string[],
  timeout = 5000
): Promise<Record<string, HealthCheckResult>> {
  const results: Record<string, HealthCheckResult> = {};

  await Promise.all(
    endpoints.map(async (endpoint) => {
      results[endpoint] = await checkApiHealth(endpoint, timeout);
    })
  );

  return results;
}

/**
 * Get overall health status from multiple endpoint results
 */
export function getOverallHealth(
  results: Record<string, HealthCheckResult>
): HealthCheckStatus {
  const statuses = Object.values(results).map((r) => r.status);

  if (statuses.every((s) => s === "healthy")) return "healthy";
  if (statuses.some((s) => s === "unknown")) return "unknown";
  return "unhealthy";
}

/**
 * Format health check result for display
 */
export function formatHealthResult(result: HealthCheckResult): string {
  let output = `Status: ${result.status.toUpperCase()}\n`;
  output += `Message: ${result.message}\n`;
  output += `Timestamp: ${result.timestamp}\n`;

  if (result.latency !== undefined) {
    output += `Latency: ${result.latency}ms\n`;
  }

  if (result.details) {
    output += `\nDetails:\n`;
    if (result.details.endpoint) {
      output += `  Endpoint: ${result.details.endpoint}\n`;
    }
    if (result.details.statusCode) {
      output += `  Status Code: ${result.details.statusCode}\n`;
    }
    if (result.details.error) {
      output += `  Error: ${result.details.error}\n`;
    }
  }

  return output;
}
