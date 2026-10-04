# Feature Pages Structural Overhaul: From Wall of Text to Visual Rhythm

**Date**: March 25, 2026
**Type**: Design | Refactoring
**Components**: Features Pages, UI Components, Design System, Navigation

## Summary

Complete structural redesign of all 7 product feature pages (InfraHub, ServiceHub, Runner, Security, Agent Fleet, CLI, Open Source). Replaced the monotonous repeating `CapabilitySection` pattern — identical across every page — with a varied, visually rich section structure using animated terminals, tabbed code viewers, metrics strips, bento grids, flow step visualizations, and scroll-triggered animations. Built a new shared component toolkit (`src/components/product/shared/`) that serves as the foundation for all product pages. Additionally moved Agent Fleet and CLI pages under the `/features/*` route group so they receive the sticky sub-navigation header.

## Problem Statement / Motivation

The previous copywriting overhaul (same day) updated content to reflect the current product, but the structural presentation remained poor. Every product page rendered identically: Hero, then 6 `CapabilitySection` rows alternating left/right, then CTA. The `CapabilitySection` component was copy-pasted verbatim into 7 separate files. Each row was the same layout: icon + title + paragraph + bullet list + optional `Card + pre` code block.

The homepage (v3) used 18 distinct section types across its 18 sections. The product pages used exactly 1.

### Pain Points

- All 7 product pages looked identical — no visual personality per module
- Monotonous wall of text with no rhythm, no breathing room, no spatial variety
- Code examples rendered as flat `Card + pre` blocks — no terminal chrome, no syntax coloring, no interactivity
- ASCII art diagrams used for the Runner architecture (literally box-drawing characters in a `<pre>` tag)
- Shared design system components (`TerminalWindow`, `Step`, `MetricCard`, `Metric`, `Grid`, `Divider`, `FeatureCard`) exported from `shared.tsx` but not used on any product page
- Framer Motion available in dependencies but not used for any scroll or entrance animations
- Agent Fleet (`/agents`) and CLI (`/cli`) pages lived outside the `/features/*` route group and didn't receive the sticky sub-navigation bar
- The lack of product screenshots made the text-heavy pages feel even more sparse

## Solution / What's New

### New Shared Component Toolkit

Built 7 reusable components in `src/components/product/shared/`:

| Component | Purpose |
|-----------|---------|
| `ScrollReveal` | Framer Motion scroll-triggered fade-up animation with configurable direction, delay, distance |
| `StaggerContainer` / `StaggerItem` | Staggered entrance animations for lists and grids |
| `AnimatedTerminal` | Terminal window with traffic-light chrome, typewriter line-by-line reveal, per-line status coloring, copy button |
| `CodeTabs` | Tabbed code viewer with animated tab switching (e.g., GCP/AWS/Azure YAML manifests) |
| `MetricsStrip` | Full-width horizontal band with count-up animated metric values |
| `BentoGrid` / `BentoItem` | CSS Grid layout with `wide` and `tall` span variants for varied card sizes |
| `FlowSteps` | Horizontal step-flow visualization with connecting arrows, responsive vertical on mobile |
| `ScreenshotPlaceholder` | Browser chrome frame with dot-grid interior for future product screenshots |

### Shared Page Skeleton

All 7 pages now follow a consistent-but-varied skeleton:

1. **Hero** — Badge, headline, pain/solution, CTAs, AnimatedTerminal showing the product in action
2. **Metrics Strip** — 3-4 key numbers with count-up animation
3. **Primary Showcase(s)** — 1-2 flagship features with CodeTabs, architecture diagrams, or FlowSteps
4. **Feature Bento Grid** — 3-4 secondary features as compact, visually varied cards (each with a unique accent)
5. **Deep Dive** — One capability with full AnimatedTerminal demo + 3 equal-height FeatureCards
6. **Screenshot Placeholder** — Optional browser chrome frame for future screenshots
7. **CTA** — Divider + ScrollReveal-wrapped call-to-action

### Per-Page Structural Personality

Each page customizes the skeleton with unique elements:

- **InfraHub**: Multi-provider CodeTabs (GCP/AWS/Azure), Infra Charts FlowSteps (Template → Values → Render → DAG → Deployed), preset tag cloud, Pulumi/Terraform visual toggle
- **ServiceHub**: Git-push-to-deploy AnimatedTerminal, 3-step environment promotion FlowSteps (Dev → Staging → Production), Buildpacks/Dockerfile CodeTabs, deploy target grid
- **Runner**: CSS architecture diagram replacing ASCII art (Control Plane / Encrypted Tunnel / Runner layers), 6-step secure tunnel FlowSteps, IRSA credential resolution terminal
- **Security**: Secret resolution AnimatedTerminal, multi-backend SecretRef CodeTabs (GCP SM / AWS SM / Vault), layered Runner trust model diagram, color-coded audit diff output
- **Agent Fleet**: Agent session AnimatedTerminal with tool calls, skill YAML + agent config CodeTabs, MCP tool call mini-terminal, test results visual
- **CLI**: Install + apply AnimatedTerminal, `planton kubectl` proxied access terminal, command subcommand tags, .env generation terminal
- **Open Source**: `openmcf apply` AnimatedTerminal, protobuf/YAML/CLI CodeTabs, 5-step Forge contribution FlowSteps, "What Planton Adds" 2x4 feature grid

### Route Consolidation

Moved Agent Fleet and CLI under the `/features/*` route group:

| Before | After | Old URL |
|--------|-------|---------|
| `/agents` | `/features/agent-fleet` | Redirects to new path |
| `/cli` | `/features/cli` | Redirects to new path |

Updated all references in: features layout sub-nav, header navigation, product overview module grid, planton-copilot redirect.

## Implementation Details

### Component Architecture

Each product page follows a 3-file pattern (unchanged from before):

```
src/components/product/<module>/
  ├── index.tsx         (barrel export — unchanged)
  ├── hero.tsx          (AnimatedTerminal + inline accent strips)
  ├── capabilities.tsx  (MetricsStrip + Showcases + BentoGrid + DeepDive)
  └── cta.tsx           (Divider + ScrollReveal wrapper)
```

The `capabilities.tsx` file in each module now renders a React Fragment containing multiple distinct `Section` components, each using different shared primitives. This replaced the single `capabilities` array mapped over identical `CapabilitySection` rows.

### Key Technical Decisions

- **Framer Motion for animations**: `useInView` with `once: true` for scroll-triggered reveals. Line-by-line typewriter via `AnimatePresence` with staggered delays. Count-up metrics via `requestAnimationFrame` (not Framer, to avoid the `set-state-in-effect` lint rule).
- **CSS architecture diagrams over SVG**: The Runner page's architecture diagram uses styled `Box` components with borders and chips rather than SVG or ASCII art. Consistent with the design system, theme-aware, responsive.
- **Equal-height cards**: `className="h-full"` on both `StaggerItem` (Framer Motion `motion.div`) and `FeatureCard` to work with CSS Grid's default `align-items: stretch`.
- **No new dependencies**: All animations use the existing `framer-motion@^11.15.0` dependency. No new packages added.

### Lint Fixes

- Fixed `react-hooks/set-state-in-effect` error in `MetricsStrip.tsx` — moved the regex match inside the effect and removed synchronous `setState` from the effect body
- Removed unused `FeatureCard` and `Grid` imports from Open Source capabilities
- Fixed pre-existing `variant` unused variable warning in `shared.tsx` by renaming to `_variant`
- Removed unused `PipelinesIcon` import from InfraHub capabilities

## Benefits

- **Visual variety**: No two consecutive sections use the same layout pattern. The eye encounters a different visual shape every scroll
- **Interactive code demos**: Tabbed multi-provider YAML manifests replace static single-provider examples
- **Animated engagement**: Typewriter terminals, count-up metrics, and scroll-triggered reveals create a living, breathing page
- **Developer authenticity**: AnimatedTerminals showing real CLI output and deployment flows demonstrate the product working, not just describing it
- **Space filling without screenshots**: Bento grids, flow visualizations, architecture diagrams, and metrics strips create visual substance without requiring product screenshots
- **Consistent but unique**: Shared skeleton ensures coherence across pages while per-page customizations give each module its own personality
- **Reusable foundation**: The 7 shared components can be used for any future product page or marketing section

## Impact

### Pages Affected

- **7 product pages** structurally redesigned (InfraHub, ServiceHub, Runner, Security, Agent Fleet, CLI, Open Source)
- **2 pages** moved to new routes with redirects (`/agents` → `/features/agent-fleet`, `/cli` → `/features/cli`)
- **Navigation** updated (header, features sub-nav, product overview)
- **1 redirect** updated (`/features/planton-copilot` → `/features/agent-fleet`)

### Files Changed

~30 files across:
- `src/components/product/shared/` — 8 new files (7 components + barrel)
- `src/components/product/{infra-hub,service-hub,runner,security,agent-fleet,cli,open-source}/` — 3 files each (hero, capabilities, cta)
- `src/app/(root)/features/{agent-fleet,cli}/page.tsx` — 2 new route files
- `src/app/(root)/{agents,cli}/page.tsx` — 2 redirect conversions
- `src/app/(root)/features/layout.tsx` — sub-nav paths
- `src/app/(root)/features/planton-copilot/page.tsx` — redirect target
- `src/components/layout/header/header.tsx` — navigation links
- `src/components/product/overview/modules-grid.tsx` — module card links
- `src/components/landing-page/v3-2026-01-02-1000/shared.tsx` — lint fix

## Related Work

- Website copywriting overhaul (same day) — updated content; this change restructured presentation
- Monochrome black-and-white theme redesign (same day) — visual identity; this change added structural variety
- Homepage v3 (January 2026) — established the 18-section design language that product pages now match in quality

---

**Status**: ✅ Live
**Timeline**: Single session
