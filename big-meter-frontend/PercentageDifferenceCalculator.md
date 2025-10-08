# Percentage Difference Calculator in Frontend

## Overview

The frontend calculates **percentage decrease/increase** in water usage by comparing the **latest month** vs **previous month**.

---

## Core Formula

**Location:** `utils.ts:160-165`

```typescript
export function computePct(prev?: number, curr?: number) {
  if (prev == null || prev === 0) return null;
  const prevVal = prev;
  const currVal = curr ?? 0;
  return ((currVal - prevVal) / prevVal) * 100;
}
```

### Formula Breakdown:

```
Percentage Change = ((Current - Previous) / Previous) × 100
```

**Examples:**
- Previous: 1000, Current: 900 → **-10%** (10% decrease)
- Previous: 1000, Current: 1100 → **+10%** (10% increase)
- Previous: 1000, Current: 700 → **-30%** (30% decrease)
- Previous: 0 → **null** (can't divide by zero)

---

## How It's Used

### 1. **Filtering Rows** (`utils.ts:242-274`)

```typescript
export function filterRows(
  rows: Row[],
  latestYm: string | undefined,
  prevYm: string | undefined,
  threshold: number,
  search: string
) {
  return rows.filter((row) => {
    const current = latestYm ? (row.values[latestYm] ?? 0) : 0;
    const previous = prevYm ? (row.values[prevYm] ?? 0) : 0;
    const pct = previous > 0 ? ((current - previous) / previous) * 100 : null;

    // Only show rows where decrease >= threshold
    const passesThreshold = threshold <= 0 || (pct != null && pct <= -threshold);

    if (!passesThreshold) return false;
    // ... search filtering
  });
}
```

**Key Logic:**
- `pct <= -threshold` → Only **negative** percentages (decreases) pass
- Threshold 10% → Shows rows with ≥10% decrease
- Threshold 0% → Shows all rows

### 2. **Visual Color Coding** (`utils.ts:276-284`)

```typescript
export function resolveBadgeClass(pct: number | null) {
  if (pct == null || !isFinite(pct) || pct >= 0)
    return "bg-slate-100 text-slate-800";  // Gray (no change/increase)

  const drop = Math.abs(pct);
  if (drop > 30)  return "bg-red-500/20 text-red-700";      // Red
  if (drop >= 15) return "bg-orange-500/20 text-orange-700"; // Orange
  if (drop >= 5)  return "bg-yellow-400/30 text-yellow-800";  // Yellow
  return "bg-slate-100 text-slate-800";                       // Gray
}
```

**Color Bands:**
| Decrease % | Color | Meaning |
|------------|-------|---------|
| > 30% | 🔴 Red | Critical drop |
| 15-30% | 🟠 Orange | High drop |
| 5-15% | 🟡 Yellow | Moderate drop |
| < 5% or increase | ⚪ Gray | Normal |

### 3. **Display in Table** (`DataTable.tsx:103-115`)

```typescript
const pct = prevYm
  ? computePct(row.values[prevYm], value)
  : null;

<LatestUsageCell
  value={value}
  pct={pct}      // -10.5 for 10.5% decrease
  history={historyData}
  isMobile={isMobile}
/>
```

Displays as: `900 (-10.5%)` with color-coded badge

---

## Example Scenario

**User selects:**
- Latest month: **202501** (January 2025)
- Threshold: **10%**

**Data for Customer A:**
- December 2024 (`202412`): 1000 cubic meters
- January 2025 (`202501`): 850 cubic meters

**Calculation:**
```
pct = ((850 - 1000) / 1000) × 100 = -15%
```

**Result:**
- ✅ Passes threshold (15% ≥ 10%)
- 🟠 Shows in **orange** badge (15-30% range)
- 📊 Display: `850 (-15.0%)`
- 📈 Hover shows sparkline trend

---

## Key Features

### Month Selection
- **Latest month** (`latestYm`): User-selected current month
- **Previous month** (`prevYm`): `monthsAll[1]` = automatically 1 month before latest

### Threshold Control
- **Default:** 10%
- **Range:** 0-100%
- **Saved:** localStorage (`detail.threshold`)
- **Logic:** Only shows rows where `decrease ≥ threshold`

### Null Handling
- Previous = 0 → Returns `null` (no percentage shown)
- Current = undefined → Treated as 0
- Null percentage → Gray badge, no filtering

---

## Visual Examples

**Data Table Display:**

| Customer | Dec 2024 | Jan 2025 | Change |
|----------|----------|----------|---------|
| A | 1000 | 850 | 🟠 850 (-15.0%) |
| B | 500 | 400 | 🟡 400 (-20.0%) |
| C | 2000 | 1200 | 🔴 1200 (-40.0%) |
| D | 800 | 780 | ⚪ 780 (-2.5%) |

**With 10% threshold:**
- A, B, C shown ✅
- D hidden (< 10% decrease)

**With 0% threshold:**
- All shown ✅

---

## Implementation Flow

```
1. User sets filters (month, branch, threshold)
   ↓
2. Load data for multiple months
   ↓
3. Combine rows with values per month
   ↓
4. Calculate percentage for each row:
   pct = ((current - previous) / previous) × 100
   ↓
5. Filter rows where pct <= -threshold
   ↓
6. Apply color badge based on severity
   ↓
7. Display in table with sparkline on hover
```

---

## Files Involved

| File | Purpose |
|------|---------|
| `utils.ts:160-165` | `computePct()` - Core calculation |
| `utils.ts:242-274` | `filterRows()` - Apply threshold filter |
| `utils.ts:276-284` | `resolveBadgeClass()` - Color coding |
| `DataTable.tsx:103` | Compute pct for latest month |
| `DataTable.tsx:179` | Display badge with percentage |
| `DetailPage.tsx:163-166` | Use prevMonth for comparison |

---

## Edge Cases Handled

### 1. Division by Zero
```typescript
if (prev == null || prev === 0) return null;
```
- **Why:** Can't calculate percentage if previous value is 0
- **Result:** No percentage shown, gray badge

### 2. Missing Current Value
```typescript
const currVal = curr ?? 0;
```
- **Why:** Missing data treated as 0 (no usage)
- **Result:** Shows 100% decrease if previous > 0

### 3. Infinite/NaN Values
```typescript
if (pct == null || !isFinite(pct) || pct >= 0)
  return "bg-slate-100 text-slate-800";
```
- **Why:** Handle bad data gracefully
- **Result:** Gray badge for invalid percentages

### 4. Threshold = 0
```typescript
const passesThreshold = threshold <= 0 || (pct != null && pct <= -threshold);
```
- **Why:** Allow viewing all data regardless of change
- **Result:** Shows all rows when threshold is 0

---

## Performance Considerations

1. **Memoization:** Percentage calculations are memoized via `useMemo` in `DetailPage.tsx`
2. **Filtering:** Happens in-memory after data fetch (no backend filtering)
3. **Re-calculation:** Only when `threshold`, `search`, or data changes
4. **Display:** Percentages computed once per row, not per render

---

## Future Enhancements

### Potential Features:
1. **Configurable color bands** - Allow users to customize thresholds
2. **Multiple comparison modes:**
   - Month-over-month (current)
   - Year-over-year
   - vs Average
   - vs Custom baseline
3. **Alert thresholds** - Separate color coding from filtering
4. **Percentage ranges** - Show both decrease AND increase
5. **Statistical analysis** - Standard deviation, outliers

---

**Last Updated:** 2025-01-08
**Maintained by:** PWA Development Team
