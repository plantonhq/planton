# Fix Login/Signup URLs and React Hydration Errors

**Date**: January 9, 2026
**Type**: Bug Fix
**Components**: Navigation, UI Components, Investor Deck

## Summary

Fixed critical issues affecting user navigation and React hydration that were causing console errors and incorrect redirects. Login button now properly directs to `/login`, Join Beta signup goes to `/signup`, and all React hydration mismatches have been resolved. These fixes ensure users land on the correct authentication pages and eliminate console errors that could impact user experience and SEO.

## Problem Statement / Motivation

Multiple issues were discovered affecting the user experience and technical health of the website:

### Pain Points

- **Incorrect Login URL**: Login button in header directed users to `https://planton.ai/` instead of the login page at `/login`
- **Incorrect Signup URL**: "Sign Up & Join Beta" button in the Join Beta modal directed to the home page instead of `/signup`
- **Hydration Error in Investor Deck**: Server/client mismatch when initializing slide index from URL hash caused React to throw hydration warnings
- **Invalid HTML Nesting**: Typography components (`<p>` tags) containing Box components (`<div>` tags) violated HTML nesting rules, causing additional hydration errors
- **Console Errors**: Multiple hydration warnings cluttering browser console and potentially impacting performance

## Solution / What's New

Fixed all authentication URLs and eliminated React hydration errors through targeted code changes:

### Authentication URL Fixes

**Header Login Button** (Desktop & Mobile)
- Updated both desktop and mobile login buttons to use correct URL
- File: `src/components/layout/header/header.tsx`

```typescript
// Before
href="https://planton.ai/"

// After
href="https://planton.ai/login"
```

**Join Beta Signup Button**
- Updated GetStartedBtn component to direct to signup page
- File: `src/components/common/typography.tsx`
- Affects all "Get Started" CTAs throughout the site

```typescript
// Before
href="https://planton.ai"

// After
href="https://planton.ai/signup"
```

### Hydration Error Fixes

**InvestorDeckV2 Component**
- File: `src/components/invest/v2/InvestorDeckV2.tsx`
- **Problem**: `useState(getInitialSlideIndex)` returned different values on server (always 0) vs client (potentially different based on URL hash)
- **Solution**: Always initialize with `0`, then update slide index in `useEffect` after hydration completes

```typescript
// Before - caused hydration mismatch
function getInitialSlideIndex(): number {
  if (typeof window === 'undefined') return 0;  // Server: 0
  const hash = window.location.hash.slice(1);
  if (hash) {
    const index = slideConfig.findIndex(s => s.id === hash);
    if (index !== -1) return index;  // Client: possibly different
  }
  return 0;
}
const [currentSlideIndex, setCurrentSlideIndex] = useState(getInitialSlideIndex);

// After - consistent server/client render
const [currentSlideIndex, setCurrentSlideIndex] = useState(0);

useEffect(() => {
  // Update slide index after hydration
  const hash = window.location.hash.slice(1);
  if (hash) {
    const index = slideConfig.findIndex(s => s.id === hash);
    if (index !== -1) {
      setCurrentSlideIndex(index);
    }
  }
}, []);
```

**ServiceHub Components - Invalid HTML Nesting**
- Files:
  - `src/components/landing-page/v3-2026-01-02-1000/ServiceHub.tsx`
  - `src/components/landing-page/v2-2025-12-31-0900/ServiceHub.tsx`
- **Problem**: Typography components (render as `<p>`) contained Box components (render as `<div>`), violating HTML nesting rules
- **Solution**: Added `component="div"` prop to Typography, making it render as `<div>` instead of `<p>`

```typescript
// Before - invalid HTML: <p><div>...</div></p>
<Typography className="text-lg font-semibold text-white mb-6 flex items-center gap-3">
  <Box className="w-8 h-8 rounded-lg bg-[#10b981]/20 flex items-center justify-center">
    <CheckIcon />
  </Box>
  What You Get
</Typography>

// After - valid HTML: <div><div>...</div></div>
<Typography component="div" className="text-lg font-semibold text-white mb-6 flex items-center gap-3">
  <Box className="w-8 h-8 rounded-lg bg-[#10b981]/20 flex items-center justify-center">
    <CheckIcon />
  </Box>
  What You Get
</Typography>
```

## Implementation Details

### Files Modified

1. **src/components/layout/header/header.tsx**
   - Line ~346: Mobile header login button URL
   - Line ~369: Desktop header login button URL

2. **src/components/common/typography.tsx**
   - Line 70: GetStartedBtn href updated to `/signup`

3. **src/components/invest/v2/InvestorDeckV2.tsx**
   - Removed `getInitialSlideIndex()` helper function
   - Changed `useState` initialization to always use `0`
   - Added `useEffect` to handle slide index from URL hash after hydration

4. **src/components/landing-page/v3-2026-01-02-1000/ServiceHub.tsx**
   - Line 43: Added `component="div"` to Typography

5. **src/components/landing-page/v2-2025-12-31-0900/ServiceHub.tsx**
   - Line 69: Added `component="div"` to "No More" Typography
   - Line 89: Added `component="div"` to "You Get" Typography

### Technical Approach

**Hydration Fix Pattern**: The key insight is that server-rendered HTML must match the initial client render. Any differences (like reading `window.location.hash`) must happen in `useEffect` after hydration completes, not during initial render.

**HTML Nesting Fix Pattern**: When Material-UI Typography needs to contain block-level elements (like Box), explicitly set `component="div"` to avoid invalid `<p>` → `<div>` nesting.

## Benefits

✅ **Correct User Navigation**: Users clicking "Login" now land on the login page, not the console home
✅ **Improved Conversion**: "Sign Up & Join Beta" button directs to signup, reducing friction in onboarding
✅ **Clean Console**: No more React hydration warnings in browser console
✅ **SEO Health**: Hydration errors can negatively impact search rankings; these are now resolved
✅ **Better UX**: Investor deck with URL hashes now works correctly without hydration flashes
✅ **Code Quality**: Follows React best practices for server-side rendering and hydration

## Impact

**Users**: No more confusion about authentication flow; correct pages load on first click
**Developers**: Clean console makes debugging easier; proper hydration patterns set good example
**Marketing**: Improved conversion funnel with correct signup flow
**SEO**: Eliminated hydration warnings that could signal poor page quality to search engines

## Related Work

- Authentication flow improvements would be complemented by session handling work
- ServiceHub component fixes align with broader landing page optimization efforts
- Investor deck improvements support the `/invest` page experience

---

**Status**: ✅ Live
**Files Changed**: 5 files
**Lines Modified**: ~15 line changes across all files
