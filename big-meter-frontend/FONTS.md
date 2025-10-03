# Font Configuration Guide

## Current Setup

The application currently uses **Google Fonts** with optimized loading:

- **Font:** Sarabun (Thai web font)
- **Weights:** 300, 400, 500, 600, 700
- **Loading Strategy:** `font-display: swap` (already configured)
- **Optimization:** DNS preconnect for faster loading

### Preconnect Tags (Already Added)

```html
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
```

These tags enable the browser to establish early connections to Google Fonts servers, reducing latency.

## Why font-display: swap?

The current Google Fonts URL already includes `&display=swap`:

```css
@import url("https://fonts.googleapis.com/css2?family=Sarabun:wght@300;400;500;600;700&display=swap");
```

**Benefits:**
- ✅ Text is immediately visible using fallback font
- ✅ Custom font swaps in when loaded (no blocking)
- ✅ Eliminates Flash of Invisible Text (FOIT)
- ✅ Better Core Web Vitals (FCP, LCP)

## Self-Hosting Fonts (Alternative)

If you want to eliminate the Google Fonts dependency entirely, follow these steps:

### Step 1: Download Font Files

```bash
# Create fonts directory
mkdir -p public/fonts

# Download Sarabun from Google Fonts
# Visit: https://fonts.google.com/specimen/Sarabun
# Click "Download family"
# Or use google-webfonts-helper: https://gwfh.mranftl.com/fonts/sarabun
```

### Step 2: Add Font Files to Project

Place the downloaded `.woff2` files in `public/fonts/`:

```
public/
└── fonts/
    ├── sarabun-300.woff2
    ├── sarabun-400.woff2
    ├── sarabun-500.woff2
    ├── sarabun-600.woff2
    └── sarabun-700.woff2
```

### Step 3: Update CSS

Replace the Google Fonts import in `src/styles/index.css`:

**Before:**
```css
@import url("https://fonts.googleapis.com/css2?family=Sarabun:wght@300;400;500;600;700&display=swap");
@import "tailwindcss";
```

**After:**
```css
@import "tailwindcss";

@layer base {
  /* Sarabun - Light 300 */
  @font-face {
    font-family: "Sarabun";
    font-style: normal;
    font-weight: 300;
    font-display: swap;
    src: url("/fonts/sarabun-300.woff2") format("woff2");
  }

  /* Sarabun - Regular 400 */
  @font-face {
    font-family: "Sarabun";
    font-style: normal;
    font-weight: 400;
    font-display: swap;
    src: url("/fonts/sarabun-400.woff2") format("woff2");
  }

  /* Sarabun - Medium 500 */
  @font-face {
    font-family: "Sarabun";
    font-style: normal;
    font-weight: 500;
    font-display: swap;
    src: url("/fonts/sarabun-500.woff2") format("woff2");
  }

  /* Sarabun - Semi-Bold 600 */
  @font-face {
    font-family: "Sarabun";
    font-style: normal;
    font-weight: 600;
    font-display: swap;
    src: url("/fonts/sarabun-600.woff2") format("woff2");
  }

  /* Sarabun - Bold 700 */
  @font-face {
    font-family: "Sarabun";
    font-style: normal;
    font-weight: 700;
    font-display: swap;
    src: url("/fonts/sarabun-700.woff2") format("woff2");
  }

  :root {
    --font-sans: "Sarabun", ui-sans-serif, system-ui, -apple-system,
      BlinkMacSystemFont, "Segoe UI", sans-serif;
  }

  /* ... rest of your CSS */
}
```

### Step 4: Remove Preconnect Tags

In `index.html`, remove the Google Fonts preconnect tags:

```html
<!-- Remove these lines -->
<link rel="preconnect" href="https://fonts.googleapis.com" />
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
```

### Step 5: Add Preload (Optional)

For critical font weights (400 and 600), add preload in `index.html`:

```html
<head>
  <!-- ... other meta tags ... -->

  <!-- Preload critical fonts -->
  <link
    rel="preload"
    href="/fonts/sarabun-400.woff2"
    as="font"
    type="font/woff2"
    crossorigin
  />
  <link
    rel="preload"
    href="/fonts/sarabun-600.woff2"
    as="font"
    type="font/woff2"
    crossorigin
  />
</head>
```

## Comparison: Google Fonts vs Self-Hosted

### Google Fonts (Current)

**Pros:**
- ✅ Easy to implement and maintain
- ✅ Automatic updates to font files
- ✅ CDN distributed globally
- ✅ Browser caching across sites
- ✅ No bandwidth cost on your server

**Cons:**
- ❌ External dependency
- ❌ Privacy concerns (Google tracking)
- ❌ Extra DNS lookup + connection
- ❌ May be blocked in some networks

### Self-Hosted

**Pros:**
- ✅ No external dependencies
- ✅ Better privacy (GDPR compliant)
- ✅ Full control over font files
- ✅ Single domain (better for CSP)
- ✅ No external tracking

**Cons:**
- ❌ Manual font updates required
- ❌ Uses your bandwidth
- ❌ Larger initial bundle size
- ❌ Need to manage cache headers

## Performance Comparison

### Load Time Impact

**Google Fonts with Preconnect:**
- DNS lookup: ~20-50ms (eliminated with preconnect)
- Font download: ~100-200ms (varies by location)
- Total: ~100-200ms

**Self-Hosted:**
- Font download: ~50-100ms (same domain)
- Total: ~50-100ms

**Self-Hosted with Preload:**
- Font download: ~50-100ms (parallel with HTML)
- Total: ~50-100ms (non-blocking)

### Bundle Size

**Google Fonts:** 0 KB (external)

**Self-Hosted (WOFF2):**
- Sarabun-300: ~12 KB
- Sarabun-400: ~12 KB
- Sarabun-500: ~12 KB
- Sarabun-600: ~12 KB
- Sarabun-700: ~12 KB
- **Total: ~60 KB**

## Recommended Approach

### For Most Projects (Current Setup)
✅ **Use Google Fonts with `font-display: swap` and preconnect**

This is the current configuration and works well for most use cases.

### When to Self-Host

Consider self-hosting if:
- Operating in a region where Google services are blocked
- Strict privacy/GDPR requirements
- Need to work completely offline
- Want to optimize for returning visitors
- Have strict CSP requirements

## Optimization Tips

### 1. Use Subset Fonts

If using only Thai language, create a subset that includes only Thai characters:

```css
@font-face {
  font-family: "Sarabun";
  src: url("/fonts/sarabun-400-thai.woff2") format("woff2");
  unicode-range: U+0E00-0E7F; /* Thai Unicode range */
}
```

### 2. Variable Fonts

Consider using a variable font file (if available) to reduce the number of files:

```css
@font-face {
  font-family: "Sarabun";
  src: url("/fonts/sarabun-variable.woff2") format("woff2-variations");
  font-weight: 300 700; /* Supports all weights */
  font-display: swap;
}
```

### 3. Cache Headers

Configure your server to cache fonts aggressively:

**Nginx:**
```nginx
location /fonts/ {
  expires 1y;
  add_header Cache-Control "public, immutable";
}
```

**Vite (vite.config.ts):**
```typescript
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          if (assetInfo.name?.endsWith('.woff2')) {
            return 'assets/fonts/[name]-[hash][extname]';
          }
          return 'assets/[name]-[hash][extname]';
        },
      },
    },
  },
});
```

### 4. Reduce Font Weights

Only include weights you actually use:

```typescript
// Audit your codebase
grep -r "font-" src/ | grep -E "font-(light|normal|medium|semibold|bold)"
```

Current usage in Big Meter:
- 300 (light): Not used
- 400 (normal): Base text
- 500 (medium): Some UI elements
- 600 (semibold): Headers, buttons
- 700 (bold): Major headings

**Optimization:** Remove 300 and 700 if not needed, saving ~24 KB.

## Testing

### Check Font Loading

1. Open DevTools → Network tab
2. Filter by "Font"
3. Check:
   - Load time
   - File size
   - Whether fonts are blocking render

### Lighthouse Audit

```bash
# Install Lighthouse CLI
npm install -g lighthouse

# Run audit
lighthouse https://your-domain.pwa.co.th --view

# Check:
# - First Contentful Paint (FCP)
# - Largest Contentful Paint (LCP)
# - Cumulative Layout Shift (CLS)
```

### Font-Display Test

Open DevTools → Network → Throttle to "Slow 3G" and reload. Text should appear immediately with fallback font, then swap to Sarabun when loaded.

## Migration Checklist

If migrating to self-hosted fonts:

- [ ] Download Sarabun font files (.woff2)
- [ ] Place files in `public/fonts/`
- [ ] Update `src/styles/index.css` with @font-face rules
- [ ] Remove Google Fonts import
- [ ] Remove preconnect tags from `index.html`
- [ ] Add preload tags for critical fonts (optional)
- [ ] Test font loading in browser
- [ ] Run Lighthouse audit
- [ ] Configure cache headers on server
- [ ] Update CSP headers to remove Google Fonts

## Further Reading

- [Google Fonts Best Practices](https://csswizardry.com/2020/05/the-fastest-google-fonts/)
- [Font Loading Strategies](https://web.dev/font-best-practices/)
- [Self-Hosting Google Fonts](https://google-webfonts-helper.herokuapp.com/)
- [Variable Fonts Guide](https://web.dev/variable-fonts/)

---

**Current Status:** Using Google Fonts with `font-display: swap` + preconnect ✅

**Recommendation:** Keep current setup unless you have specific requirements for self-hosting.

**Last Updated:** 2025-10-03
