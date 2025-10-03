# TODO Plan — Validation, Logging, Metrics

Scope: Track follow-up tasks to enhance API robustness and observability.

Items

1) Define validation rules per endpoint — [in_progress]
   - Required params, formats (e.g., `ym=YYYYMM`), ranges, enumerations
   - Whitelist order fields and sort directions

2) Implement request validation middleware — [pending]
   - Centralize parsing + error responses (400 w/ details)

3) Add structured request logging middleware — [pending]
   - Method, path, status, latency, client IP, request ID

4) Expose Prometheus metrics endpoint — [pending]
   - `/metrics` (separate mux or guarded route)

5) Instrument handlers and DB timings — [pending]
   - Counters, histograms (durations), per-endpoint labels

6) Add unit tests for handlers — [pending]
   - Happy-path, invalid inputs, pagination, zeroed-row detection

7) Document configs and dashboards — [pending]
   - ENV knobs, sample Grafana dashboards, alert suggestions

Notes

- Keep changes minimal and aligned with current project structure.
- Avoid adding secrets; configs should use `.env.example` only.
