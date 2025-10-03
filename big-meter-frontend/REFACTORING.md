# Detail Page Refactoring Plan

## Current State

**File:** `src/screens/DetailPage.tsx`
**Lines:** 1,173 lines
**Status:** Monolithic component with multiple concerns

## Problems

1. **Too large** - Difficult to navigate and maintain
2. **Mixed concerns** - Data fetching, business logic, UI rendering all in one file
3. **Hard to test** - Tightly coupled logic
4. **Code duplication** - Similar patterns repeated
5. **Poor reusability** - Components can't be used elsewhere

## Refactoring Strategy

### Phase 1: Extract Utilities (✅ COMPLETED)

**Created:**
- `src/components/DetailPage/utils.ts` - Pure functions and types
- `src/components/DetailPage/UIComponents.tsx` - Simple presentational components

**Extracted:**
- Type definitions (Row, constants)
- Formatting functions (fmtNum, fmtPct, fmtThMonth)
- Date/time utilities (buildMonths, defaultLatestYm)
- Data transformation (combineRows, filterRows)
- Business logic (computePct, resolveBadgeClass)
- Error interpretation (interpretErrorMessage)

### Phase 2: Extract Components (TODO)

#### 2.1 Filter Section Component

**File:** `src/components/DetailPage/FilterSection.tsx`

**Props:**
```typescript
interface FilterSectionProps {
  branch: string;
  setBranch: (value: string) => void;
  latestYm: string;
  setLatestYm: (value: string) => void;
  threshold: number;
  setThreshold: (value: number) => void;
  branches: BranchItem[];
  onApply: () => void;
  onReset: () => void;
}
```

**Contains:**
- Month/Year selectors
- Branch dropdown
- Threshold input
- Apply/Reset buttons

**Lines:** ~110

#### 2.2 Results Header Component

**File:** `src/components/DetailPage/ResultsHeader.tsx`

**Props:**
```typescript
interface ResultsHeaderProps {
  filteredCount: number;
  search: string;
  setSearch: (value: string) => void;
  pageSize: number;
  setPageSize: (value: number) => void;
  historyMonths: number;
  setHistoryMonths: (value: number) => void;
  onExport: () => void;
  isExporting: boolean;
  applied: boolean;
  isMobile: boolean;
}
```

**Contains:**
- Result count display
- Color legend
- Search input
- Export button
- History months toggle
- Page size selector

**Lines:** ~80

#### 2.3 Data Table Component

**File:** `src/components/DetailPage/DataTable.tsx`

**Props:**
```typescript
interface DataTableProps {
  rows: Row[];
  months: string[];
  latestYm: string;
  baseIndex: number;
  isMobile: boolean;
}
```

**Contains:**
- Table header
- Table rows
- Cell rendering logic

**Lines:** ~110

**Sub-components to extract:**
- `MonthHeader.tsx` - Month column header
- `LatestUsageCell.tsx` - Cell with sparkline
- `TrendSparkline.tsx` - Sparkline chart

#### 2.4 Sparkline Components

**File:** `src/components/DetailPage/TrendSparkline.tsx`

**Props:**
```typescript
interface TrendSparklineProps {
  data: TrendPoint[];
  width: number;
  height: number;
}
```

**Contains:**
- Visx chart rendering
- Gradient definitions
- Scale calculations

**Lines:** ~80

### Phase 3: Extract Custom Hooks (TODO)

#### 3.1 useWaterUsageData Hook

**File:** `src/hooks/useWaterUsageData.ts`

**Purpose:** Manage data fetching and transformation

```typescript
export function useWaterUsageData(
  applied: AppliedFilters | null,
  isAuthenticated: boolean
) {
  // Branches query
  // Custcodes query
  // Details queries
  // Combined rows
  // Error handling
  return { rows, isLoading, isFetching, error, warning };
}
```

#### 3.2 useFilterState Hook

**File:** `src/hooks/useFilterState.ts`

**Purpose:** Manage filter state and persistence

```typescript
export function useFilterState(user: AuthUser | null, branches: BranchItem[]) {
  // Branch, YM, threshold state
  // LocalStorage persistence
  // Default values from user
  return { /* state and setters */ };
}
```

#### 3.3 useExportData Hook

**File:** `src/hooks/useExportData.ts`

**Purpose:** Handle export functionality

```typescript
export function useExportData() {
  const [isExporting, setIsExporting] = useState(false);

  const handleExport = async (params: ExportParams) => {
    // Dynamic import
    // File generation
    // Error handling with toast
  };

  return { isExporting, handleExport };
}
```

### Phase 4: Reorganize File Structure (TODO)

```
src/
├── screens/
│   └── DetailPage.tsx          (~200 lines - main orchestration)
├── components/
│   └── DetailPage/
│       ├── FilterSection.tsx   (~110 lines)
│       ├── ResultsHeader.tsx   (~80 lines)
│       ├── DataTable.tsx       (~110 lines)
│       ├── MonthHeader.tsx     (~20 lines)
│       ├── LatestUsageCell.tsx (~60 lines)
│       ├── TrendSparkline.tsx  (~80 lines)
│       ├── UIComponents.tsx    (~150 lines - done ✅)
│       ├── utils.ts            (~300 lines - done ✅)
│       └── types.ts            (~50 lines)
└── hooks/
    ├── useWaterUsageData.ts    (~150 lines)
    ├── useFilterState.ts       (~100 lines)
    └── useExportData.ts        (~50 lines)
```

**Total:** ~1,460 lines (organized, reusable, testable)
**Original:** 1,173 lines (monolithic)

## Refactored DetailPage Structure

After refactoring, the main `DetailPage.tsx` would look like:

```typescript
import { FilterSection } from '../components/DetailPage/FilterSection';
import { ResultsHeader } from '../components/DetailPage/ResultsHeader';
import { DataTable } from '../components/DetailPage/DataTable';
import { Pager, ErrorState, WarningState, EmptyState } from '../components/DetailPage/UIComponents';
import { useWaterUsageData } from '../hooks/useWaterUsageData';
import { useFilterState } from '../hooks/useFilterState';
import { useExportData } from '../hooks/useExportData';

export default function DetailPage() {
  const { user, hydrated, logout } = useAuth();
  const { showToast } = useToast();
  const navigate = useNavigate();
  const isMobile = useMediaQuery("(max-width: 767px)");

  // Filter state
  const {
    branch,
    setBranch,
    latestYm,
    setLatestYm,
    threshold,
    setThreshold,
    historyMonths,
    setHistoryMonths,
    applied,
    setApplied,
    handleApply,
    handleReset,
  } = useFilterState(user, branches);

  // Data fetching
  const {
    branches,
    rows,
    isLoading,
    isFetching,
    error,
    warning,
  } = useWaterUsageData(applied, isAuthenticated);

  // Filtering and pagination
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const filteredRows = /* ... */;
  const pageRows = /* ... */;

  // Export
  const { isExporting, handleExport } = useExportData();

  if (!hydrated) return <LoadingState />;
  if (!user) return <Navigate to="/" replace />;

  return (
    <div className="min-h-screen bg-gray-50 px-3 py-4 md:px-6 md:py-8">
      {/* Header */}
      <header>{/* ... */}</header>

      {/* Filter Section */}
      <FilterSection
        branch={branch}
        setBranch={setBranch}
        latestYm={latestYm}
        setLatestYm={setLatestYm}
        threshold={threshold}
        setThreshold={setThreshold}
        branches={branches}
        onApply={handleApply}
        onReset={handleReset}
      />

      {/* Results Section */}
      <section>
        <ResultsHeader
          filteredCount={filteredRows.length}
          search={search}
          setSearch={setSearch}
          pageSize={pageSize}
          setPageSize={setPageSize}
          historyMonths={historyMonths}
          setHistoryMonths={setHistoryMonths}
          onExport={() => handleExport(/* params */)}
          isExporting={isExporting}
          applied={!!applied}
          isMobile={isMobile}
        />

        {!applied && <EmptyState message="..." />}
        {applied && warning && <WarningState warning={warning} />}
        {applied && error && <ErrorState error={error} />}

        {applied && !error && (
          <>
            {isLoading ? (
              <LoadingState />
            ) : (
              <>
                <DataTable
                  rows={pageRows}
                  months={monthsToDisplay}
                  latestYm={applied.ym}
                  baseIndex={(page - 1) * pageSize}
                  isMobile={isMobile}
                />
                <Pager page={page} totalPages={totalPages} onChange={setPage} />
              </>
            )}
          </>
        )}
      </section>
    </div>
  );
}
```

**Result:** ~200 lines instead of 1,173

## Benefits

### 1. Maintainability ✅
- Each component has a single responsibility
- Easy to locate and fix bugs
- Clear separation of concerns

### 2. Testability ✅
- Utility functions can be unit tested
- Components can be tested in isolation
- Hooks can be tested independently

### 3. Reusability ✅
- Components can be used in other pages
- Hooks can be shared across features
- Utilities are pure functions

### 4. Performance ✅
- Smaller components = better React optimization
- Memoization opportunities
- Easier to identify performance bottlenecks

### 5. Developer Experience ✅
- Easier to onboard new developers
- Better IDE navigation
- Clearer code organization

## Implementation Steps

### Immediate (Done ✅)
1. ✅ Create `utils.ts` with helper functions
2. ✅ Create `UIComponents.tsx` with simple components

### Short-term (Next PR)
1. Extract `FilterSection` component
2. Extract `ResultsHeader` component
3. Extract `DataTable` and sparkline components
4. Update `DetailPage` to use new components
5. Test thoroughly

### Medium-term
1. Create custom hooks for data fetching
2. Create custom hooks for state management
3. Further split large components
4. Add unit tests for utilities
5. Add component tests

### Long-term
1. Consider state management library (Zustand/Jotai) if needed
2. Add Storybook for component documentation
3. Performance optimization with React.memo
4. Accessibility improvements

## Testing Strategy

### Unit Tests (Utilities)
```typescript
describe('fmtNum', () => {
  it('formats numbers with Thai locale', () => {
    expect(fmtNum(1234.56)).toBe('1,234.56');
  });
});

describe('filterRows', () => {
  it('filters rows by threshold', () => {
    const rows = [/* ... */];
    const result = filterRows(rows, '202412', '202411', 10, '');
    expect(result.length).toBeLessThanOrEqual(rows.length);
  });
});
```

### Component Tests
```typescript
describe('FilterSection', () => {
  it('calls onApply when apply button clicked', () => {
    const onApply = jest.fn();
    render(<FilterSection {...props} onApply={onApply} />);
    fireEvent.click(screen.getByText('แสดงรายงาน'));
    expect(onApply).toHaveBeenCalled();
  });
});
```

### Integration Tests
```typescript
describe('DetailPage', () => {
  it('loads data when filters applied', async () => {
    render(<DetailPage />);
    // Select branch
    // Click apply
    // Wait for data
    // Assert table rendered
  });
});
```

## Migration Path

To avoid breaking changes:

1. **Parallel Development**
   - Create new components alongside existing file
   - Test components independently
   - Gradually replace sections

2. **Feature Flags**
   - Use env variable to toggle new components
   - A/B test if needed

3. **Incremental Rollout**
   - Refactor one section at a time
   - Deploy and monitor
   - Continue to next section

## Checklist

- [x] Extract utility functions
- [x] Extract UI components
- [x] Extract FilterSection component
- [x] Extract ResultsHeader component
- [x] Extract DataTable component
- [x] Extract sparkline components
- [ ] Create useWaterUsageData hook
- [ ] Create useFilterState hook
- [ ] Create useExportData hook
- [x] Update DetailPage to use new structure
- [ ] Add unit tests for utilities
- [ ] Add component tests
- [ ] Update documentation
- [ ] Performance testing
- [ ] Accessibility audit

## Notes

- Keep backward compatibility during transition
- Use TypeScript strictly for type safety
- Follow existing code style and conventions
- Document complex logic
- Add comments for non-obvious code

---

**Created:** 2025-10-03
**Last Updated:** 2025-10-03
**Status:** Phase 2 Complete ✅ (Components extracted and integrated)
**Current State:** DetailPage reduced from 1,173 lines to 295 lines (75% reduction)
**Next Steps:** Phase 3 - Create custom hooks (useWaterUsageData, useFilterState, useExportData)
