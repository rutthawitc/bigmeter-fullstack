export function apiBase(): string {
  const base = import.meta.env.VITE_API_BASE_URL as string | undefined
  if (!base) return ''
  if (!/^https?:\/\//i.test(base)) return ''
  return base.replace(/\/$/, '')
}

export function buildUrl(path: string, params?: Record<string, string | number | boolean | undefined>) {
  const url = new URL(`${apiBase()}${path}`, window.location.origin)
  if (params) for (const [k, v] of Object.entries(params)) if (v !== undefined && v !== '') url.searchParams.set(k, String(v))
  return url.toString()
}

export async function fetchJson<T>(input: RequestInfo, init?: RequestInit): Promise<T> {
  const res = await fetch(input, init)
  const text = await res.text()
  if (!res.ok) {
    let message = `HTTP ${res.status}`
    if (text) {
      try {
        const body = JSON.parse(text)
        if (typeof body.error === 'string' && body.error.trim()) message = body.error
      } catch {
        message = text
      }
    }
    const error = new Error(message)
    ;(error as Error & { status?: number }).status = res.status
    throw error
  }
  return text ? (JSON.parse(text) as T) : ({} as T)
}
