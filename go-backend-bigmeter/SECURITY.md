# Security and Secrets Policy

This repository must not contain secrets. Keep credentials and sensitive values out of source control.

What to do
- Use local `.env` files for development, never commit them. A safe template is provided as `.env.example` (root) and `configs/.env.example`.
- For Docker Compose, copy `.env.example` to `.env` and fill in values locally.
- For application configuration, copy `configs/.env.example` into `configs/.env` (or rely on exported environment variables), and set DSNs there.

Oracle and Postgres DSNs
- Oracle (godror, EZCONNECT): `USER/PASS@host:1521/SERVICE` or `USER/PASS@host:1521/SID`
- Postgres (pgx): `postgres://user:pass@host:5432/db?sslmode=disable`

Operational guidance
- If a secret is accidentally committed: immediately rotate it in the upstream system (Oracle/Postgres), then remove or neutralize the file in git history if necessary.
- `.gitignore` already excludes `.env` and `configs/.env`. Do not override this.
- Review PRs for accidental inclusions of credentials, tokens, or private endpoints.

Reporting
- If you suspect a leak, open a private channel to the maintainers and rotate credentials right away.

