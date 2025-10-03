# Repository Guidelines

## Project Structure & Module Organization

- Stack: Vite + React 19, TypeScript, TailwindCSS (no daisyUI).
- Layout:
  - `index.html`, `vite.config.ts`, `tsconfig.json`.
  - `src/main.tsx` (bootstraps `QueryClientProvider`), `src/App.tsx`.
  - `src/components/` (UI), `src/lib/` (utils). Details page will be re-created.
  - `src/styles/` with `index.css` (Tailwind base, components, utilities).
  - `public/` static assets; `docs/` API and UI references.
  - Tests colocated as `*.test.ts(x)` or under `tests/`.

## Build, Test, and Development Commands

- Init (if not set up): `pnpm create vite@latest . -- --template react-ts`.
- Install UI: `pnpm add -D tailwindcss @tailwindcss/vite` (Tailwind v4). No PostCSS config needed.
- Dev: `pnpm dev` (Vite HMR).
- Build/Preview: `pnpm build` → `dist/`, `pnpm preview`.
- Test/Lint/Format: `pnpm test` (Vitest), `pnpm lint`, `pnpm format`.
- Tooling: `packageManager` pins pnpm 10.17.0. Use `corepack enable` and `pnpm --version` (update via `pnpm self-update`) to match.
- Auto-approve builds: this repo whitelists build scripts for `esbuild` and `@swc/core` (and their platform variants) via `package.json > pnpm.allowedScripts`, so installs are non-interactive.
- Dev proxy: Vite proxies `/api` to `VITE_API_BASE_URL` (default `http://localhost:8089`). Keep API calls relative (e.g., `fetch('/api/v1/healthz')`).

## Coding Style & Naming Conventions

- TypeScript strict, 2-space indent, UTF-8, LF.
- Components `PascalCase.tsx` in `src/components`; utilities `camelCase.ts` in `src/lib`.
- Queries: define in `src/api/`; name keys like `['meters', id]`. Mutations in feature folders.
- Tailwind: Tailwind v4 uses CSS-first config. In `src/styles/index.css` use `@import "tailwindcss";`.
- Vite env vars: prefix with `VITE_` and document in `.env.example`.

## Testing Guidelines

- Vitest + React Testing Library. Add `src/setupTests.ts` and configure in Vitest.
- Name tests `*.test.ts`/`*.test.tsx`; focus on logic and critical UI flows.
- Mock network with MSW when exercising TanStack Query.

## Commit & Pull Request Guidelines

- Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`). Subject ≤ 72 chars.
- PRs: clear description, linked issues, screenshots/GIFs for UI, and test notes.
- Keep diffs focused; one logical change per PR.

## Security & Configuration Tips

- Never commit secrets; use `.env.local` (gitignored) and provide `.env.example`.
- Align API usage with `docs/API-Spec.md`/`docs/api_spec.md` and verify request/response shapes.
