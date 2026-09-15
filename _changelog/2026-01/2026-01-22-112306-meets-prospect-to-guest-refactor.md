# Meets Micro-App: Prospect to Guest Terminology Refactor

**Date**: January 22, 2026
**Type**: Refactoring
**Components**: Meets Micro-App, URL Routing, Guest Registry

## Summary

Refactored the Meets micro-app to rename "prospect" terminology to "guest" throughout the codebase, and updated the date format to include timestamps. This makes the terminology more inclusive (supporting companies, investors, partners, advisors, and individuals) and enables more precise meeting organization.

## Problem Statement / Motivation

The original implementation used "prospect" terminology which was too sales-focused and limiting. The Meets micro-app serves various meeting types beyond sales:

### Pain Points

- "Prospect" implies only sales/business development meetings
- Investors, partners, and advisors are not prospects
- Individual meetings (advisory calls, networking) don't fit the prospect framing
- The date format `YYYY-MM-DD` lacked precision for scheduling
- Multiple meetings on the same day couldn't be distinguished

## Solution / What's New

### Terminology Changes

| Before | After |
|--------|-------|
| `[prospect]` route | `[guest]` route |
| `prospects/` folder | `guests/` folder |
| `ProspectConfig` type | `GuestConfig` type |
| `prospectRegistry` | `guestRegistry` |
| `getProspectConfig()` | `getGuestConfig()` |
| `getLatestProspectConfig()` | `getLatestGuestConfig()` |
| `listProspects()` | `listGuests()` |

### Date Format Update

Changed from `YYYY-MM-DD` to `yyyy-mm-dd-hhmm`:

```
Before: sep/2026-01-23
After:  sep/2026-01-23-1400
```

This allows:
- Multiple meetings per day to be distinguished
- Clear scheduling context in the URL
- Sortable format for chronological ordering

### URL Structure (Unchanged)

The URLs remain memorable and easy to type:
- `/meets/[guest]` - Always shows the upcoming/latest meeting
- `/meets/[guest]/[date]` - Specific dated meeting

## Implementation Details

### Route Changes

```
src/app/(micro-apps)/meets/
├── [prospect]/           → [guest]/
│   ├── page.tsx          (params: guest instead of prospect)
│   └── [date]/
│       ├── page.tsx      (params: guest, date)
│       └── MeetsDeckClient.tsx (unchanged)
```

### Component Changes

```
src/components/meets/
├── prospects/            → guests/
│   ├── index.ts          (GuestConfig, guestRegistry, getGuestConfig, etc.)
│   └── sep/
│       ├── config.ts     (sepConfig with guest field)
│       └── slides/       (unchanged)
└── MeetsDeck.tsx         (MeetsDeckProps.guest instead of .prospect)
```

### Registry Entry

```typescript
// Before
const prospectRegistry: Record<string, ProspectConfig> = {
  'sep/2026-01-23': sepConfig,
};

// After
const guestRegistry: Record<string, GuestConfig> = {
  'sep/2026-01-23-1400': sepConfig,
};
```

## Benefits

### For Users
- **Inclusive terminology**: "Guest" applies to all meeting types
- **Shareable URLs**: `planton.ai/meets/acme` is easy to remember and share
- **Meeting history**: All meetings with a guest organized under their slug

### For Development
- **Type safety**: `GuestConfig` clearly communicates intent
- **Scalability**: Date format supports multiple daily meetings
- **Consistency**: Single terminology throughout codebase

## Impact

### Files Changed

| Category | Count |
|----------|-------|
| Route files | 2 modified, 2 renamed |
| Component files | 2 modified |
| Slide files | 24 renamed (path only) |
| **Total** | 28 files |

### URLs Affected

- `/meets/sep` - Now routes through `[guest]` parameter
- `/meets/sep/2026-01-23-1400` - Updated date format

### No Breaking Changes

The SEP presentation remains accessible at the same primary URL (`/meets/sep`). The dated URL changed format but this was preproduction.

## Related Work

- **Meets Micro-App Foundation**: `2026-01-21-224947-meets-micro-app-presentation-framework.md`
- **Meeting Deck Rules**: See companion changelog for the rules system

---

**Status**: ✅ Live
**Timeline**: Single session refactoring
**URL**: https://planton.ai/meets/sep
