# Meets Micro-App: Prospect Presentation Framework

**Date**: January 21, 2026
**Type**: Feature
**Components**: Meets Micro-App, Interactive Demo, UI Components, Design System

## Summary

Built a production-grade, reusable presentation framework for prospect meetings at `planton.ai/meets/<prospect>`. The foundation includes a slide-per-file architecture, shared UI primitives, presenter notes, keyboard/touch navigation, and a prospect registry system. First implementation: a 24-slide SEP demo deck with complete content, animations, and presenter notes.

## Problem Statement / Motivation

Planton needed a professional, reusable system for delivering tailored presentations to prospects during sales meetings. The requirements:

1. **Prospect-specific customization**: Each prospect gets a personalized deck with their company context
2. **Version control**: Multiple presentations per prospect (by date) with history
3. **Presenter support**: Built-in notes, keyboard navigation, direct slide linking
4. **Reusability**: Foundation components usable across all future prospect decks
5. **Web-native**: No PowerPoint exports—live web presentations with URLs for follow-up

### Pain Points

- Manual slide deck creation in Google Slides/PowerPoint lacked version control
- No easy way to track which deck version was shown to which prospect
- Slide content couldn't be programmatically generated or kept in sync with product
- Sharing decks required exporting PDFs or granting edit access
- No presenter notes visible during live presentations

## Solution / What's New

### Architecture

A modular presentation system following the existing `invest` micro-app pattern:

```
src/
├── app/(micro-apps)/meets/
│   ├── layout.tsx              # Root layout with metadata
│   ├── meets.css               # Shared styles
│   ├── page.tsx                # /meets → redirects to /meets/sep
│   └── [prospect]/
│       ├── page.tsx            # /meets/sep → latest presentation
│       └── [date]/
│           ├── page.tsx        # /meets/sep/2026-01-23
│           └── MeetsDeckClient.tsx
│
└── components/meets/
    ├── MeetsDeck.tsx           # Presentation engine
    ├── shared/
    │   ├── primitives.tsx      # 20+ reusable UI components
    │   ├── navigation.tsx      # Progress bar, controls, keyboard
    │   └── presenter-notes.tsx # Toggle-able notes panel
    └── prospects/
        ├── index.ts            # Prospect registry
        └── sep/
            ├── config.ts       # Slide array + metadata
            └── slides/         # 24 individual slide components
```

### URL Structure

- `/meets` → Redirects to default prospect (SEP)
- `/meets/sep` → Latest SEP presentation (direct render, no redirect)
- `/meets/sep/2026-01-23` → Specific dated presentation
- Hash navigation: `/meets/sep#12` → Direct link to slide 12

### Shared Primitives

Created a comprehensive design system for presentations:

| Component | Purpose |
|-----------|---------|
| `Slide` | Base wrapper with consistent padding, animations |
| `SlideHeader` | Section tag + title combo |
| `SlideTitle`, `SlideSubtitle` | Responsive typography |
| `SectionTag` | Colored category labels |
| `Card`, `CardTitle` | Content containers |
| `HighlightCard` | Emphasized content blocks |
| `NumberedCard` | Sequential items with large numbers |
| `QuoteBox` | Testimonials and quotes |
| `Comparison` | Before/after side-by-side |
| `FlowDiagram` | Step-by-step process flows |
| `Checklist` | Feature lists with checkmarks |
| `IconList` | Items with emoji/icon prefixes |
| `Metric` | Large number + label combinations |
| `StatsGrid` | 3-5 column statistics display |
| `Badge`, `DemoBadge` | Status and live demo indicators |
| `Grid`, `TwoCol` | Layout utilities |
| `Callout` | Highlighted messages with variants |

### Presentation Engine (`MeetsDeck`)

Core features:
- **Slide transitions**: Framer Motion animations (fade + slide)
- **Keyboard navigation**: Arrow keys, spacebar, Home/End
- **Touch support**: Swipe gestures for mobile/tablet
- **Hash-based URLs**: Browser back/forward works, direct linking
- **Progress indicator**: Visual progress bar + dot navigation
- **Presenter notes**: Toggle with 'N' key, yellow overlay panel

### SEP Deck Content

24 slides covering the full sales narrative:

1. **S01-S06**: Problem framing (SEP's challenges, costs, ideal solution)
2. **S07-S09**: Planton introduction (Infra Hub, Service Hub)
3. **S10**: Why Planton fits SEP specifically
4. **S11-S15**: Live demo walkthrough with `DemoBadge` indicators
5. **S16-S18**: Trust signals (open source, security, compliance)
6. **S19-S20**: Business case (ROI, customer success stories)
7. **S21-S24**: Partnership, next steps, Q&A, thank you

## Implementation Details

### Next.js 15 Compatibility

The dynamic route required careful handling of Next.js 15's async params:

```typescript
// page.tsx - Server component with generateStaticParams
interface MeetsPageProps {
  params: Promise<{ prospect: string; date: string }>;
}

export function generateStaticParams() {
  const prospects = listProspects();
  return prospects.map((key) => {
    const [prospect, date] = key.split('/');
    return { prospect, date };
  });
}

export default async function MeetsPage({ params }: MeetsPageProps) {
  const { prospect, date } = await params;
  // ...
}
```

The `MeetsDeckClient` wrapper separates client-side interactivity from server-side static generation.

### Prospect Registry

Central lookup system for presentations:

```typescript
// prospects/index.ts
const prospectRegistry: Record<string, ProspectConfig> = {
  'sep/2026-01-23': sepConfig,
};

export function getProspectConfig(prospect: string, date: string) {
  return prospectRegistry[`${prospect}/${date}`];
}

export function getLatestProspectConfig(prospect: string) {
  // Returns most recent presentation for prospect
}

export function listProspects() {
  return Object.keys(prospectRegistry);
}
```

### Customer Data Integration

Slide content pulls from actual customer data in `planton/_business/sales/customers/`:
- **TynyBay**: IT consulting firm, manages 8+ client projects with 1 DevOps
- **Odwen**: Online warehousing platform on GCP
- **iorta TechNext**: BFSI SalesVerse platform, 7 developers, junior DevOps achieving senior outcomes

## Benefits

### For Sales Team
- **Consistent branding**: Every presentation uses the same polished design system
- **Quick customization**: Copy `sep/` folder, update content, new prospect ready
- **Shareable links**: Send `/meets/prospect#slide` for follow-up discussions
- **Version history**: Git tracks every presentation version

### For Presenters
- **Presenter notes**: Hit 'N' to see talking points without audience seeing
- **Keyboard control**: Natural navigation without fumbling with mouse
- **Direct linking**: Jump to specific slides during Q&A

### For Engineering
- **Slide-per-file**: Easy to find, edit, or reorder individual slides
- **Reusable primitives**: 20+ UI components accelerate future deck creation
- **Type-safe**: Full TypeScript throughout with proper interfaces
- **Static export**: Works with `output: 'export'` for CDN deployment

## Impact

### Files Created

| Category | Count |
|----------|-------|
| Slide components | 24 |
| Shared primitives | 3 files, 20+ UI components |
| Route handlers | 4 |
| Configuration | 2 |
| **Total** | **33 files** |

### URL Routes Added

- `/meets` (redirect)
- `/meets/[prospect]` (latest presentation)
- `/meets/[prospect]/[date]` (specific presentation)

### Design System Expansion

The primitives in `shared/primitives.tsx` can be reused for:
- Future prospect decks
- Internal training materials
- Product walkthroughs
- Conference presentations

## Related Work

- **Invest Micro-App**: Followed similar architecture pattern (`src/app/(micro-apps)/invest/`)
- **Customer Data**: Content sourced from `planton/_business/sales/customers/` documentation

## Future Enhancements

- **PDF export**: Generate static PDFs for offline sharing
- **Analytics**: Track which slides prospects spend time on
- **Template system**: More starting templates beyond SEP
- **Collaborative editing**: Allow non-engineers to edit content via CMS

---

**Status**: ✅ Live
**Timeline**: Single session implementation
**URL**: https://planton.ai/meets/sep
