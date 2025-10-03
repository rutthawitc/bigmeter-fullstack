# Big Meter Frontend - Project Overview

## Project Description

**big-meter-frontend** is a React-based web application for the Provincial Waterworks Authority (PWA) of Thailand. It provides a dashboard system for monitoring and analyzing large-scale water usage data across different branches. The application enables users to:

- View detailed water consumption reports by branch and time period
- Track and identify significant changes in water usage patterns
- Filter customers by usage decrease thresholds
- Export reports to Excel format
- Visualize historical usage trends with interactive sparkline charts

The application is written in **Thai language** and integrates with PWA's Intranet authentication system.

## Tech Stack

### Core Technologies

- **React 19.0.0** - UI framework (latest version)
- **TypeScript 5.5.4** - Type-safe JavaScript
- **Vite 5.4.8** - Build tool and dev server
- **pnpm 10.17.0** - Package manager

### Key Dependencies

- **@tanstack/react-query 5.55.0** - Data fetching and caching
- **react-router-dom 6.26.2** - Client-side routing
- **@visx/visx 3.12.0** - Data visualization (sparkline charts)
- **xlsx 0.18.5** - Excel export functionality
- **Tailwind CSS 4.1.0** - Utility-first CSS framework
- **@vitejs/plugin-react-swc** - Fast React refresh using SWC compiler

### Build Tools

- **TypeScript** with strict mode enabled
- **SWC** - Fast JavaScript/TypeScript compiler
- **Prettier** - Code formatting

## Project Structure

```
big-meter-frontend/
├── src/
│   ├── api/                    # API layer
│   │   ├── http.ts            # Base HTTP utilities (fetch wrapper, URL builder)
│   │   ├── branches.ts        # Branch data API
│   │   ├── custcodes.ts       # Customer code API
│   │   └── details.ts         # Water usage details API
│   ├── lib/                   # Utilities and hooks
│   │   ├── auth.tsx           # Authentication context and hooks
│   │   ├── useMediaQuery.ts   # Responsive design hook
│   │   └── exportDetailsXlsx.ts # Excel export utility
│   ├── screens/               # Page components
│   │   ├── LoginPage.tsx      # Login/authentication page
│   │   └── DetailPage.tsx     # Main dashboard page
│   ├── styles/
│   │   └── index.css          # Global styles (Tailwind imports)
│   ├── routes.tsx             # React Router configuration
│   ├── App.tsx                # Root app component
│   └── main.tsx               # Application entry point
├── vite.config.ts             # Vite configuration
├── tsconfig.json              # TypeScript configuration
└── package.json               # Dependencies and scripts
```

## Architecture & Patterns

### 1. **Authentication Flow**

- **AuthProvider** (`src/lib/auth.tsx`) provides global authentication state
- User credentials stored in localStorage (`big-meter.auth.user`)
- Integration with PWA Intranet authentication endpoint
- Protected routes redirect to login if user is not authenticated
- Auth state includes user profile data (name, position, organization, etc.)

**Key files:**

- `src/lib/auth.tsx:36-90` - Context provider and hooks
- `src/screens/LoginPage.tsx:5-138` - Login UI and authentication logic

### 2. **Data Fetching Strategy**

Uses **TanStack Query (React Query)** for:

- Automatic caching and background refetching
- Loading and error states management
- Parallel data fetching for multiple months
- Query key-based cache invalidation

**Configuration:**

```typescript
// src/main.tsx:9-11
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 60_000,
      refetchOnWindowFocus: false,
    },
  },
});
```

**Key patterns:**

- `useQuery` for single data sources (branches, custcodes)
- `useQueries` for parallel fetching of multiple months' data
- Enabled/disabled queries based on authentication and filter state

### 3. **API Layer Design**

Centralized HTTP utilities with:

- Base URL configuration from environment variables
- Automatic error handling with status codes
- Type-safe response interfaces
- URL builder with query parameter support

**Base utilities** (`src/api/http.ts`):

- `apiBase()` - Returns API base URL from env
- `buildUrl()` - Constructs URLs with query params
- `fetchJson<T>()` - Type-safe fetch wrapper with error handling

### 4. **State Management**

Combination of:

- **React Query** - Server state (API data)
- **React Context** - Authentication state
- **Local component state** - UI state (filters, pagination, search)
- **localStorage** - Persistent user preferences (threshold, months)

**Local storage keys:**

- `detail.threshold` - Usage decrease threshold percentage
- `detail.months` - Number of historical months to display

### 5. **Routing Structure**

Simple two-page application:

```typescript
// src/routes.tsx:5-8
const router = createBrowserRouter([
  { path: "/", element: <App /> },           // Login page
  { path: "/details", element: <DetailPage /> }, // Dashboard
])
```

### 6. **Responsive Design**

- Mobile-first Tailwind CSS approach
- `useMediaQuery` hook for conditional rendering
- Adaptive sparkline sizes (160x60 mobile, 260x90 desktop)
- Responsive table with hidden columns on mobile
- Fixed 3-month history on mobile devices

**Key implementation:**

```typescript
// src/screens/DetailPage.tsx:66-67
const isMobile = useMediaQuery("(max-width: 767px)");
const effectiveMonths = isMobile ? 3 : historyMonths;
```

### 7. **Data Transformation Pipeline**

Complex data processing in DetailPage:

1. **Fetch** - Parallel queries for multiple months
2. **Combine** - Merge detail records with customer metadata
3. **Filter** - Apply threshold and search criteria
4. **Paginate** - Slice for current page
5. **Display** - Render with sparklines and visual indicators

**Key functions:**

- `combineRows()` - Merges monthly data with customer metadata
- `filterRows()` - Applies threshold and search filters
- `buildMonths()` - Generates month list going backwards from latest

### 8. **Visualization Components**

Interactive sparkline charts using visx:

- Hover-triggered tooltip displays
- Area charts with gradient fills
- Line paths with circular end markers
- Responsive sizing based on device

**Components:**

- `LatestUsageCell` - Cell with hover sparkline
- `TrendSparkline` - visx-based chart component

### 9. **Excel Export**

Lazy-loaded export functionality to reduce bundle size:

```typescript
// src/screens/DetailPage.tsx:217
const { exportDetailsToXlsx } = await import("../lib/exportDetailsXlsx");
```

Dynamic import reduces initial bundle by ~700KB.

### 10. **Error Handling**

Multi-level error interpretation:

- HTTP status code detection (400, 404, 500+)
- Custom Thai language error messages
- Specific handling for database field description errors
- Warning vs error states (custcodes warnings, details errors)

**Error types:**

- `KnownError` - Error with optional status code
- Warning state - Non-critical custcode failures
- Error state - Critical detail fetch failures

## Environment Variables

```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8089   # Backend API base URL
VITE_LOGIN_API=https://intranet.pwa.co.th/login/webservice_login6.php  # PWA login endpoint
```

**Vite proxy configuration** (`vite.config.ts:22-37`):

- `/api` → Proxies to `VITE_API_BASE_URL`
- `/auth/login` → Proxies to `VITE_LOGIN_API`

## API Endpoints

### Branches API

**Endpoint:** `GET /api/v1/branches`

**Query params:** `q`, `limit`, `offset`

**Response:**

```typescript
{ items: BranchItem[], total: number, limit: number, offset: number }
```

### Customer Codes API

**Endpoint:** `GET /api/v1/custcodes`

**Query params:** `branch`, `ym`, `fiscal_year`, `q`, `limit`, `offset`

**Response:**

```typescript
{ items: CustCodeItem[], total?: number, limit?: number, offset?: number }
```

### Details API

**Endpoint:** `GET /api/v1/details`

**Query params:** `ym`, `branch`, `q`, `limit`, `offset`, `order_by`, `sort`

**Response:**

```typescript
{ items: DetailItem[], total: number, limit: number, offset: number }
```

## Key Features

### 1. **Filter Controls**

- Month/Year selector (Thai Buddhist calendar)
- Branch dropdown (defaults to user's assigned branch)
- Threshold percentage (0-100%, default 10%)
- History months toggle (3/6/12 months, desktop only)

### 2. **Data Table**

- Sortable, paginated results
- Conditional row highlighting based on usage decrease:
  - 5-15%: Yellow
  - 15-30%: Orange
  - > 30%: Red
- Responsive column hiding on mobile
- Search across all visible fields

### 3. **Sparkline Visualization**

- Hover/focus to reveal historical trend
- Area + line chart combination
- Shows oldest to latest month
- Accessible with keyboard navigation

### 4. **Export Functionality**

- Exports filtered results to Excel
- Includes all columns and monthly data
- Dynamic filename: `big-meter-{branch}-{ym}.xlsx`
- Loading state prevents duplicate exports

### 5. **Pagination**

- Configurable page sizes (10/25/50)
- Ellipsis for large page counts
- Previous/Next navigation
- Page status indicator

## Important Patterns & Conventions

### Date Format

- **Year-Month format:** `YYYYMM` (e.g., `202412`)
- **Thai Buddhist year display:** Gregorian year + 543
- **Month abbreviations:** Thai abbreviated month names

### Component Organization

- **Screen components** - Full page layouts
- **Inline components** - DataTable, TrendSparkline, Pager, etc.
- **Utility functions** - Colocated with usage in DetailPage

### TypeScript Patterns

- Strict mode enabled
- Explicit return types for API functions
- Nullable fields properly typed (`| null`)
- Type guards for error handling

### Styling Approach

- Tailwind utility classes
- Responsive modifiers (`md:`, `lg:`)
- Custom color palette (slate, blue, red, orange, yellow)
- No CSS modules or styled-components

## Development Scripts

```bash
pnpm dev       # Start development server on port 5173
pnpm build     # Build for production (TypeScript check + Vite build)
pnpm preview   # Preview production build
pnpm lint      # Run linter (currently placeholder)
pnpm format    # Format code with Prettier
```

## Recent Changes (from git history)

1. **Lazy export** - XLSX module now dynamically imported (~700KB bundle reduction)
2. **CSS fix** - Replaced unsupported `@theme` rule with `@layer base`
3. **XLSX export** - Added Excel export functionality for detail reports
4. **Responsive sparklines** - Added hover tooltips with historical trends

## Known Issues & Considerations

### Backend Dependencies

- Requires specific error format: `{ error: string }`
- Expects field descriptions to match in database responses
- Top-200 data may not be immediately available for new branches

### Browser Compatibility

- Modern browsers only (ES2020 target)
- Requires localStorage support
- Fetch API required (no polyfills)

### Performance Considerations

- Up to 12 parallel API requests on report load
- Large datasets may slow pagination
- Excel export blocks UI during file generation
- Initial bundle includes large xlsx library (mitigated by lazy loading)

## Future Enhancement Opportunities

1. **Caching improvements** - More aggressive caching strategies
2. **Virtualization** - Windowing for large datasets
3. **CSV export** - Lighter alternative to XLSX
4. **Advanced filters** - Additional query parameters
5. **Chart types** - More visualization options
6. **PWA features** - Offline support, notifications
7. **Testing** - Unit/integration test coverage
8. **Accessibility** - Enhanced ARIA labels and keyboard navigation

## Deployment Notes

### Build Configuration

- TypeScript compilation required before build
- Vite build optimizes for modern browsers
- SWC compiler provides faster build times
- Environment variables must be prefixed with `VITE_`

### Production Considerations

- Set `VITE_API_BASE_URL` to production backend
- Ensure `VITE_LOGIN_API` points to PWA Intranet
- CORS must be configured on backend
- HTTPS recommended for production

## Developer Notes

### Working with this codebase

1. **Main logic** is concentrated in `src/screens/DetailPage.tsx` (~1163 lines)
2. **API layer** is minimal and straightforward - extend by adding new files in `src/api/`
3. **Styling** uses Tailwind v4 with Vite plugin - no PostCSS config needed
4. **State** is mostly local to DetailPage - consider extracting if complexity grows
5. **Types** are colocated with API definitions - keep in sync with backend

### Adding new features

- **New page:** Add route in `src/routes.tsx` and create component in `src/screens/`
- **New API:** Create file in `src/api/` following existing pattern
- **New utility:** Add to `src/lib/` and import where needed
- **Authentication:** Use `useAuth()` hook, access user data from context

### Debugging tips

- React Query DevTools can be added for debugging queries
- Check Network tab for API response formats
- localStorage contains auth data and preferences
- Console warnings for invalid environment variables

---

**Last Updated:** 2025-10-03
**Project Version:** 0.1.0
**Maintainer:** PWA Development Team
