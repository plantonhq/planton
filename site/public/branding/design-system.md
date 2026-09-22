---
title: "Planton Design System"
description: "Canonical visual language reference for the Planton platform — website, console, and CLI."
hideCopyMarkdown: false
hideViewMarkdown: false
---

# Planton Design System

**The Planton platform uses a monochrome visual language across every surface — marketing website, web console, micro-apps, and CLI — where color is reserved exclusively for semantic status and provider brand identity. Dark mode is the default. Hierarchy comes from luminance and borders, never from decorative hue.**

---

## Purpose and Audience

This document is the canonical reference for the Planton visual language. It covers every user-facing surface of the platform and establishes the rules that all implementations must follow.

**Who this is for:**

- **Contributors** — Engineers and AI agents building on the Planton codebase. Every pattern here must exist in code, and every pattern in code must trace back here.
- **Partners and integrators** — Anyone building experiences that sit alongside or within Planton's UI.
- **LLM agents** — This document is available as raw markdown at [/branding/design-system.md](/branding/design-system.md) for programmatic consumption.

Planton is a 61+ package monorepo developed primarily by AI agents. The code is the documentation — there are no separate Figma files or design handoffs. If a pattern exists in this document, it must exist in the codebase, and vice versa.

---

## Design Philosophy

### Monochrome, Data-Dense, Dark-First

The visual language is inspired by GitHub Dark Default and the density of tools like Cursor. The rationale:

- **No colored chrome.** Navigation, headers, sidebars, buttons, tabs, and interactive elements use neutral grays. Color is reserved for semantic status (success, error, warning) and provider brand logos.
- **Dark as default.** New users start in dark mode across all surfaces. Light mode is available and persists via cookie/localStorage.
- **Hierarchy by luminance.** Visual priority is established through brightness steps in the gray ramp, not through competing hues. A reader should understand information hierarchy without learning a color code.
- **Borders, not backgrounds.** Cards and sections are separated by borders, not by contrasting background colors. This is the GitHub pattern (unified canvas) rather than the VS Code pattern (dark frame).
- **Honest semantics.** Green, red, and amber appear only where they carry meaning — deployment status, validation errors, capacity warnings. They are never decorative.

### What Was Removed

The monochrome system was established by removing: purple/sky/cyan gradient text and buttons, tinted cards, colored navigation dropdowns, rainbow compliance badges, gradient orbs, and decorative accent colors. Legacy hex values like `#7c3aed` (purple), `#0ea5e9` (sky), `#0099FF` (pricing blue), `#FDA935` (gold), and `#6665D2` (Discord purple) were all replaced with the neutral palette.

---

## The Dual-Surface Model

Planton has two primary web surfaces that share the same visual language but use different implementation stacks:

| Property | Marketing Website | Web Console |
|----------|-------------------|-------------|
| **Background** | `#0a0a0a` | `#0d1117` |
| **Styling system** | Tailwind CSS | MUI v6 styled-components |
| **Theme tokens** | TypeScript class strings (`src/theme/docs.ts`) | Token ramp pipeline (`dark-colors.ts` → `dark.tsx` → `createTheme`) |
| **Framework** | Next.js (static export) | Next.js (SSR) |
| **Color palette** | Tailwind config overrides | MUI palette + custom grey ramp |

### Why Different Background Shades

The website uses `#0a0a0a` (near-black) for a marketing-appropriate depth. The console uses `#0d1117` (GitHub's canvas color) because it provides better contrast ratios for data-dense UI — tables, code viewers, graphs, and nested panels — where WCAG compliance matters at every text size.

Both surfaces share the same design principles: monochrome chrome, semantic-only color, border-based separation, and Inter typography. The implementation differs because each surface optimizes for its use case.

---

## Color System

### Website Palette

The website palette is defined once, in the website-shell package's `tokens.ts`, and projected into both the MUI theme and Tailwind. Components use the role-named Tailwind classes below and never type a hex; the hex column is for the reader, not for code.

| Role | Hex | Tailwind class |
|------|-----|----------------|
| Canvas (page background) | `#0a0a0a` | `bg-canvas` |
| Panel (header, footer, panels one step up) | `#111111` | `bg-panel` |
| Card (marketing cards) | `#151515` | `bg-card` |
| Card under the pointer | `#1f1f1f` | `bg-card-hover` |
| Raised (inputs, pills, tooltips, code blocks; sits on a card) | `#1a1a1a` | `bg-raised` |
| Primary text | `#ededed` | `text-fg` (also `text-white`, overridden in Tailwind config) |
| Body text (inside cards) | `#b0b0b0` | `text-fg-body` |
| Secondary text | `#a0a0a0` | `text-fg-secondary` |
| Muted text | `#666666` | `text-fg-muted` |
| Faint text (fine print) | `#555555` | `text-fg-faint` |
| Border | `#2a2a2a` | `border-edge` |
| Border hover | `#3a3a3a` | `border-edge-hover` |
| Success (functional only) | `#10b981` | `text-ok`, `bg-ok/10` |
| Failure (functional only) | `#ef4444` | `text-danger` |
| Warning (functional only) | `#f59e0b` | `text-warn` |
| CTA surface | `#ffffff` | `bg-cta` (true white, not the `#ededed` override) |
| CTA text | `#000000` | `text-cta-text` |

Two card surfaces exist on purpose. Marketing cards are `#151515`, one step above the canvas, separated from it by a border rather than a fill; the lighter `#1a1a1a` is the surface that sits on a card (an input, a pill, a tooltip, a code block) so it reads as raised. An earlier revision of this table listed cards as `#1a1a1a` while every card on the site painted `#151515`; the palette now records what the pixels do.

The Tailwind `white` token is overridden to `#ededed` for body text. Call-to-action buttons use `bg-cta` to retain true white contrast.

### Console Palette (Dark Mode)

The console uses a GitHub Dark Default-aligned token ramp.

| Role | Token | Hex |
|------|-------|-----|
| Canvas | `background.default` | `#0d1117` |
| Subtle surface | `grey[20]` | `#161b22` |
| Muted surface | `grey[70]` | `#21262d` |
| Deep surface | `grey[90]` | `#010409` |
| Primary text | `text.primary` | `#e6edf3` |
| Secondary text | `text.secondary` | `#7d8590` |
| Disabled text | `text.disabled` | `#484f58` |
| Border / divider | `divider` | `#30363d` |
| Primary accent | `primary.main` | `#e6edf3` (near-white) |

### Console Palette (Light Mode)

| Role | Token | Hex |
|------|-------|-----|
| Canvas | `background.default` | `#ffffff` |
| Subtle surface | `grey[20]` | `#f6f8fa` |
| Muted surface | `grey[70]` | `#eaeef2` |
| Deep surface | `grey[90]` | `#f6f8fa` |
| Primary text | `text.primary` | `#1f2328` |
| Secondary text | `text.secondary` | `#656d76` |
| Disabled text | `text.disabled` | `#8b949e` |
| Border / divider | `divider` | `#d0d7de` |
| Primary accent | `primary.main` | `#1f2328` (near-black) |

### Semantic Color Budget

Status colors are identical in spirit across both surfaces. They are reserved exclusively for functional severity — never decoration.

| Severity | Dark | Light | Usage |
|----------|------|-------|-------|
| Success | `#3fb950` | `#1a7f37` | Deployment succeeded, health OK, pipeline passed |
| Error | `#f85149` | `#cf222e` | Failed, validation error, destructive action |
| Warning | `#d29922` | `#9a6700` | Attention needed, degraded, approaching limits |
| Info | `#7d8590` | `#656d76` | Informational severity only (gray, not blue) |

**Info is gray, not blue.** This is intentional. In a monochrome system, blue would be the only hue-based accent in the chrome, breaking the visual language. Info severity uses neutral gray; decorative uses of `info.main` are replaced with `text.secondary` or `divider`.

### Provider Brand Colors

Provider logos (AWS, GCP, Azure, Kubernetes, GitHub, GitLab) always render in their original brand colors. They are never converted to monochrome. In data contexts — org graphs, connection panels, resource lists — provider color serves a functional identification purpose.

Providers with poor contrast on dark backgrounds (AWS, GitHub) have dedicated dark-mode SVG variants.

---

## Typography

### Font Stack

**Inter** is the sole typeface across all Planton surfaces — website, console, micro-apps, and documentation.

```
font-family: 'Inter', system-ui, -apple-system, sans-serif;
```

Loaded via `next/font/google` with weights 300, 400, 500, 600, and 700. The CSS variable `--font-inter` is set on the root element.

### Weight Scale

| Weight | Name | Usage |
|--------|------|-------|
| 300 | Light | Rarely used; large display text only |
| 400 | Regular | Body text, form labels, table cells |
| 500 | Medium | Navigation items, badges, metadata |
| 600 | Semibold | Section titles, card headers, button labels |
| 700 | Bold | Page titles, hero headings |

### Heading Hierarchy (Website)

Content pages (docs, tutorials, blog) use a calmer brightness tier than marketing pages. All headings, section titles, sidebar labels, table headers, and contributor names use `#b0b0b0` (`secondary.80`) — ~10% brighter than body text (`#a0a0a0`). Size and weight carry the structural hierarchy; the subtle brightness lift aids scanning without creating visual fatigue in dense documentation. Marketing page headings (features, pricing, solutions, landing) continue to use `#ededed`.

| Element | Size (mobile → desktop) | Weight | Color | Tracking |
|---------|------------------------|--------|-------|----------|
| Page title | 24px → 36–48px | Bold | `#b0b0b0` | Tight |
| H1 (in-body) | 20px → 24–30px | Semibold | `#b0b0b0` | Tight |
| H2 | 18px → 20–24px | Semibold | `#b0b0b0` | Tight |
| H3 | 16px → 18–20px | Semibold | `#b0b0b0` | Tight |
| H4 | 16px → 18px | Semibold | `#b0b0b0` | Tight |

Body text is 16px `#a0a0a0` with `tracking-tight` in prose contexts. Heading class constants are centralized in `src/theme/docs.ts`.

The `#b0b0b0` treatment extends beyond headings to all content-area chrome: markdown table headers, sidebar section titles ("Planton Documentation", "Tutorials", "On this page", "Contributors"), tutorial list entry titles, sort controls, and author names. Interactive states (hover, active/selected) remain `text-white` for feedback contrast.

### Console Typography

The console inherits Inter through the MUI theme. MUI's default typography scale applies, with `textTransform: 'none'` and `letterSpacing: '-0.011em'` on buttons to match the website's density.

---

## Surface Hierarchy

Both surfaces use a **unified canvas** approach where the header, sidebar, and content area share the same base background. Visual hierarchy comes from borders and subtle surface-level differentiation, not from competing background shades.

### Website Surfaces

| Surface | Hex | Usage |
|---------|-----|-------|
| Canvas | `#0a0a0a` | Page background, header, sidebar, footer |
| Panel | `#111111` | Secondary sections, FAQ containers, inline panels |
| Card / Overlay | `#1a1a1a` | Feature cards, pricing cards, navigation dropdowns |
| Border | `#2a2a2a` | All structural separation |
| Border (hover / overlay) | `#3a3a3a` | Hover states, navigation dropdown borders |

### Console Surfaces (Dark)

| Surface | Token | Hex | Usage |
|---------|-------|-----|-------|
| Canvas | `background.default` | `#0d1117` | Page, header, sidebar, content area |
| Subtle | `grey[20]` | `#161b22` | Tabs, accordion headers, popover backgrounds |
| Muted | `grey[70]` | `#21262d` | Input fields, hover states, selected tabs |
| Deep | `grey[90]` | `#010409` | Terminal backgrounds, table headers |
| Border | `divider` | `#30363d` | Card borders, dividers, input outlines |

### Console Surfaces (Light)

| Surface | Token | Hex | Usage |
|---------|-------|-----|-------|
| Canvas | `background.default` | `#ffffff` | Page, header, sidebar, content area |
| Subtle | `grey[20]` | `#f6f8fa` | Tabs, accordion headers |
| Muted | `grey[70]` | `#eaeef2` | Input fields, hover states |
| Deep | `grey[90]` | `#f6f8fa` | Table headers |
| Border | `divider` | `#d0d7de` | Card borders, dividers |

### Card Separation

Cards are separated by borders, not background color differences. The console applies a global `MuiCard` override with `border: 1px solid ${theme.palette.divider}`. The website uses `border-[#2a2a2a]` on card elements. Individual components do not declare their own borders — the system provides them.

MUI's dark-mode elevation overlay (`backgroundImage: linear-gradient(rgba(255,255,255,...))`) is disabled globally via `backgroundImage: 'none'` on `MuiPaper.root` and `MuiAppBar.root`. All surfaces render at their exact `backgroundColor` with no hidden gradients.

---

## Button Hierarchy

### Website Buttons

| Variant | Appearance | Usage |
|---------|------------|-------|
| Primary CTA | `bg-[#fff] text-black` | Sign up, Get started, Deploy |
| Secondary | `hover:border-white hover:bg-white/5` | Secondary actions, navigation |
| Ghost | No background, text only | Low-emphasis actions |

Primary CTAs use true `#ffffff` (not the `#ededed` text override) to maintain maximum contrast. The black text on white surface inverts the page's dark-on-light relationship, drawing the eye.

### Console Buttons

| Variant | Dark Mode | Light Mode | Usage |
|---------|-----------|------------|-------|
| `containedPrimary` | Near-white bg, dark text | Near-black bg, white text | Create, Deploy, Save |
| `outlinedPrimary` | Transparent, divider border | Transparent, divider border | Back, Retry, Refresh |
| `containedSecondary` | Dark bg, divider border | White bg, divider border | Tertiary actions |
| `textSecondary` | No bg, secondary text | No bg, secondary text | Low-emphasis |

The CTA inversion pattern: `containedPrimary` button text uses `background.default` (the page background color). This creates dark-on-light in dark mode and white-on-dark in light mode, and automatically adapts if the page background changes.

---

## Icon Rules

| Category | Treatment | Examples |
|----------|-----------|----------|
| Navigation chrome | Monochrome (`currentColor` or CSS filter) | Discord, Registry, Library, Docs |
| Provider logos | Original brand colors | AWS, GCP, Azure, Kubernetes, GitHub |
| Provider logos (dark variant) | Dark-mode SVG variant | AWS (white text), GitHub (white octocat) |
| App-internal icons | `currentColor` | Action icons, status indicators, MUI icons |

Monochrome overrides for navigation icons are applied at the consumption site (e.g., in header actions), not in the icon library. Colored variants remain available for data contexts where color serves a functional purpose.

Provider icons with poor contrast on dark backgrounds have dark-mode SVG variants in `public/connectors/`. A utility function handles selection based on the current theme mode. For library packages that cannot import console utilities, an inline dark icon map constant is used.

---

## Dark and Light Mode

### Dark as Default

All Planton surfaces default to dark mode. The resolution chain for determining the active theme is:

1. Cookie (`planton_theme`)
2. localStorage
3. Fallback: `'dark'`

System `prefers-color-scheme` is not used. Dark is the brand default; users who prefer light mode persist their choice via the toggle.

### Flash Prevention

Both the website and console use an inline `<script>` in the root layout that executes before React hydration:

1. Reads theme from cookie → localStorage → defaults to `'dark'`
2. Sets `document.documentElement.style.backgroundColor` to the appropriate canvas color
3. Sets `document.documentElement.style.colorScheme` for native browser element styling (scrollbars, autofill)

This prevents a white flash on dark-mode pages during the hydration gap.

The console's error boundary (`global-error.tsx`) has its own independent theme detection using the same resolution chain, because it renders without ThemeProvider access.

---

## Console Token Pipeline

The console's color system flows through a strict five-layer pipeline. There are no alternative styling systems.

```
Token ramps           →  Palette config        →  createTheme()     →  ThemeProvider      →  Components
(dark-colors.ts)         (dark.tsx)                (theme.ts)           (appContext.tsx)       (styled.ts)
(light-colors.ts)        (light.tsx)
```

### Layer 1: Token Ramps

Files `dark-colors.ts` and `light-colors.ts` each export 9 color ramps (objects with numeric keys 0–100). These are the single source of truth for every hex code in the console.

| Ramp | Purpose |
|------|---------|
| `grey*` | Surfaces, text, borders, disabled states |
| `primary*` | Primary action color (CTA buttons, focus rings) |
| `secondary*` | Secondary UI elements (steppers, labels, timestamps) |
| `error*` | Error severity ramp |
| `warning*` | Warning severity ramp |
| `success*` | Success severity ramp |
| `info*` | Info severity ramp (neutral gray, not blue) |
| `exceptions*` | One-off UI chrome slots (stack job logs, triggers, banners) |
| `crimson*` | Code syntax highlighting (editor-specific) |

### Layer 2: Palette Config

Files `dark.tsx` and `light.tsx` map raw ramp slots to MUI semantic roles (`palette.background.default`, `palette.text.primary`, `palette.divider`, etc.) and define component overrides. The palette config uses direct ramp imports. Component overrides use `theme.palette.*` exclusively — never direct ramp imports.

### Layer 3: Theme Creation

File `theme.ts` exports `appTheme(type, font)` which calls `createTheme()` with the appropriate `ThemeOptions`. Also exports shared helpers consumed by both dark and light overrides.

### Layer 4: ThemeProvider

File `appContext.tsx` manages theme state with the 3-step resolution chain (cookie → localStorage → dark) and wraps the app in MUI's `ThemeProvider`.

### Layer 5: Cross-Package Theme Access

The console's 61+ library packages access the theme through two mechanisms:
- **`useTheme()`** — MUI's built-in hook for `theme.palette.*` in styled-components.
- **`PlantonThemeContext`** — A lightweight context providing `{ mode: 'light' | 'dark' }` for cases where a library component needs the current mode explicitly (e.g., selecting a dark variant of a provider icon).

### Graph and Visualization Colors

Centralized in `graph-colors.ts`, these use desaturated tinted grays (~25–40% saturation) that harmonize with the monochrome chrome while preserving hue families for functional differentiation.

| Export | Purpose |
|--------|---------|
| `GRAPH_CATEGORY` | Node types in org graph, DAG, K8s graph |
| `GRAPH_EDGE` | Edge stroke colors for relationship types |
| `CHART` | Chart line and gradient colors |
| `TERMINAL_SYNTAX` | Shell ANSI highlight colors |
| `IAC_SYNTAX` | Pulumi/Terraform diff syntax colors |

---

## Code Syntax Highlighting

Code blocks are content islands where muted hue differentiation serves a functional purpose — readability — similar to how provider logos keep brand colors and graph nodes use desaturated tints. The principle is the same as the console's `graph-colors.ts`: preserve hue families at low saturation so tokens remain distinguishable without breaking the monochrome chrome.

### Philosophy

Syntax tokens use ~25% saturation. Each hue family is preserved as a tinted gray rather than a bright color. The overall impression should be subtly warm or cool grays, not a rainbow. Diff markers (deletion/addition) stay slightly more saturated because they carry semantic meaning.

This approach matches how Cursor, Vercel, and Linear handle code in monochrome contexts.

### Muted Syntax Palette

On background `#1a1a1a` (bgTertiary), base text `#c8c8c8`:

| Token | Hex | Description |
|-------|-----|-------------|
| Keywords / operators | `#8b9bb5` | Cool gray-blue |
| Functions / params | `#c5c0a0` | Warm muted |
| Comments | `#666666` | Pure neutral (matches textMuted) |
| Strings / built-ins | `#b5a08f` | Warm gray |
| Numbers / literals | `#a0ad98` | Sage gray |
| Variables / template | `#9cb5c5` | Cool gray |
| Types / names / titles | `#85ada3` | Muted teal |
| Tags / selectors | `#8b9bb5` | Same as keywords |
| Attributes | `#9cb5c5` | Same as variables |
| Deletion | `#c07070` | Muted red (semantic) |
| Addition | `#90a888` | Muted green (semantic) |
| Meta | `#8b9bb5` | Same as keywords |

### Code Block Chrome

| Element | Value | Notes |
|---------|-------|-------|
| Fenced block background | `#1a1a1a` | bgTertiary |
| Fenced block border | `#2a2a2a` | Standard border |
| Inline code background | `#2a2a2a` | One step lighter for contrast against prose |
| Copy button (default) | `bg-[#2a2a2a]` text `#666666` | Monochrome |
| Copy button (copied) | `bg-[#10b981]/10` text `#10b981` | Semantic green for success |

### Mermaid Diagrams

Mermaid diagrams use the surface palette for all chrome — node fills, borders, edge colors, text. Git graph and pie chart retain functional colors for branch/segment differentiation, consistent with the console's `graph-colors.ts` approach.

| Element | Value |
|---------|-------|
| SVG background | `#1a1a1a` |
| Node fill | `#2a2a2a` |
| Node stroke | `#ededed` |
| Edge / line | `#666666` |
| Edge label bg | `#2a2a2a` |
| Cluster fill | `#1a1a1a` |
| Cluster stroke | `#3a3a3a` |
| Text | `#ededed` |
| Font | Inter |

### Callouts (Blockquotes)

Markdown blockquotes (`> **Tip:** ...`, `> **Note:** ...`, `> **Warning:** ...`) render as callouts. The blockquote renderer inspects the HAST node for a bold type prefix and applies the appropriate variant.

| Variant | Left Border | Background | Usage |
|---------|-------------|------------|-------|
| Default / Tip / Note | `#3a3a3a` (borderHover) | `#111111` (bgSecondary) | Informational callouts |
| Warning / Caution | `#ef4444` at 40% opacity | `#111111` (bgSecondary) | Destructive or cautionary notes |

Design rationale:
- **Background is `#111` (bgSecondary)**, distinct from both page canvas (`#0a0a0a`) and code blocks (`#1a1a1a`). Each surface level is visually separable.
- **Border is `border-l-2`** (2px) rather than 4px — thinner but brighter (`#3a3a3a`) creates a cleaner, more refined signal.
- **No italic.** The bold type prefix ("**Tip:**") and the distinct background surface are sufficient callout signals. Italic reduces readability, especially when mixed with inline code.
- **Warning uses semantic red** at low opacity — consistent with the "honest semantics" principle.

### Console Follow-Up

The console uses `react-syntax-highlighter` with bundled themes (`darcula`, `vsc-dark-plus`, `material-light`) — a different rendering pipeline from the website's CSS-based hljs. Aligning the console to the same muted philosophy requires building custom react-syntax-highlighter theme objects from the token pipeline. This is flagged as a separate effort in the console design system document.

---

## Website Theme Tokens

The website centralizes its design tokens in `src/theme/docs.ts` as Tailwind class strings. This is intentional — Tailwind class strings in TypeScript, not CSS variables — for consistency with the rest of the codebase.

| Token | Purpose |
|-------|---------|
| `LINK_CLASSES` | Inline links (`text-white/80 hover:text-white underline ...`) |
| `TAG_CLASSES` | Tags and badges (`bg-white/10 text-white/70 ...`) |
| `BLOCKQUOTE_CLASSES` | Callouts — default/tip/note (`border-[#3a3a3a] bg-[#111] ...`) |
| `BLOCKQUOTE_WARNING_CLASSES` | Callouts — warning/caution (`border-[#ef4444]/40 bg-[#111] ...`) |
| `INLINE_CODE_CLASSES` | Inline code (`bg-[#2a2a2a] text-white ...`) |
| `PARAGRAPH_CLASSES` | Body text (`text-[#a0a0a0] ...`) |
| `LIST_CLASSES` | Lists (`text-[#a0a0a0] ...`) |
| `CODE_BLOCK_CLASSES` | Fenced code block container (`bg-[#1a1a1a] border-[#2a2a2a] ...`) |
| `CODE_BLOCK_COPY_CLASSES` | Copy button default state |
| `CODE_BLOCK_COPY_ACTIVE_CLASSES` | Copy button copied state (semantic green) |
| `TABLE_CLASSES` | Table container (`bg-[#1a1a1a] border-[#2a2a2a] ...`) |
| `TABLE_HEAD_CLASSES` | Table header row (`bg-[#111]`) |
| `TABLE_ROW_CLASSES` | Table body rows (`border-[#2a2a2a]`) |
| `TABLE_HEADER_CLASSES` | Table header cells |
| `TABLE_CELL_CLASSES` | Table body cells (`text-[#a0a0a0]`) |
| `MERMAID_CONTAINER_CLASSES` | Mermaid diagram container |
| `HR_CLASSES` | Horizontal rules (`border-[#2a2a2a]`) |
| `SIDEBAR_ACTIVE_CLASSES` | Active sidebar items (`bg-white/10 text-white`) |
| `SIDEBAR_ITEM_CLASSES` | Default sidebar items (`text-[#a0a0a0]`) |

Sidebar badge colors preserve semantic meaning: `Popular` uses green (`#10b981`), `Deprecated` uses red (`#ef4444`). All other badges use white with opacity variations.

---

## Rules for Contributors

These rules apply to all Planton surfaces. Every component, every styled file, every new feature must follow them.

### Always Do

1. **Use the theme system for all colors.** On the website, use Tailwind classes from the established palette. In the console, use `theme.palette.*` for every color in styled-components and `sx` props.

2. **Use semantic tokens first.** Prefer `text.primary`, `background.paper`, `divider` (console) or `text-white`, `bg-[#0a0a0a]`, `border-[#2a2a2a]` (website) over raw values.

3. **Use `alpha()` from `@mui/material` for opacity variations** in the console. Never write raw `rgba()` with theme colors.

4. **Use theme callbacks in MUI component overrides.** All overrides use `({ theme }) => ({...})`.

5. **Use `background.default` for CTA inversion** in the console. When text sits on a `primary.main` background, use `background.default` for the text color — not `common.white`.

6. **Preserve `backgroundImage: 'none'` on Paper overrides** in the console. Without this, MUI's dark-mode elevation overlay reactivates.

7. **Import graph/chart colors from `graph-colors.ts`.** Never define inline hex values for visualization colors.

8. **Use true `#fff` for website CTAs.** The Tailwind `white` override (`#ededed`) is for body text. CTAs must use `bg-[#fff]` explicitly.

### Never Do

1. **Never hardcode hex values in component files.** No `'#0d1117'`, `'#e6edf3'`, `'#30363d'` in console components. No arbitrary hex outside the established website palette.

2. **Never use CSS named colors.** No `'white'`, `'black'`, `'red'`. Use `'common.white'`, `'common.black'`, `theme.palette.error.main` (console) or the corresponding Tailwind classes (website).

3. **Never import from token ramp files in console component code.** `dark-colors.ts` and `light-colors.ts` are consumed only by `dark.tsx` and `light.tsx`.

4. **Never use `info.main` for decoration.** Info is reserved for informational severity status only.

5. **Never use `if (mode === 'dark')` conditionals in console styling.** The theme system handles mode resolution. Styled-components access `theme.palette.*` which resolves correctly for both modes.

6. **Never add elevation-based background differentiation.** Cards are separated by borders. If a Card needs visual separation, it inherits the global border override.

7. **Never use SCSS or CSS Modules for theming in the console.** The legacy SCSS layer has been eliminated. All styling uses MUI styled-components and `sx` props.

8. **Never add new design system primitives without flagging.** If a styling need is not covered by existing tokens, flag it as a gap — do not invent a new token.

9. **Never add decorative color to chrome elements.** No colored backgrounds on headers, sidebars, or navigation. If you are reaching for a hue, ask whether it carries semantic meaning.

---

## Key File Inventory

### Website (`planton.ai` repo)

| File | Purpose |
|------|---------|
| `src/theme/docs.ts` | Centralized Tailwind class string tokens |
| `tailwind.config.ts` | Color overrides (`white: '#ededed'`), Inter font |
| `src/app/layout.tsx` | Root layout, Inter loading, flash prevention |
| `src/app/globals.css` | Global styles |

### Console (`planton` monorepo, `client-apps/web/console/`)

| File | Purpose |
|------|---------|
| `src/themes/dark-colors.ts` | Dark token ramps (source of truth for dark hex values) |
| `src/themes/light-colors.ts` | Light token ramps (source of truth for light hex values) |
| `src/themes/dark.tsx` | Dark MUI ThemeOptions + component overrides |
| `src/themes/light.tsx` | Light MUI ThemeOptions + component overrides |
| `src/themes/theme.ts` | `createTheme` + shared override helpers |
| `src/themes/graph-colors.ts` | Centralized graph/chart/syntax color constants |
| `src/contexts/appContext.tsx` | Theme state management, dark-first fallback, ThemeProvider |
| `src/components/providers/index.tsx` | Context bridges including PlantonThemeBridge |
| `src/app/layout.tsx` | Flash prevention inline script |
| `src/app/global-error.tsx` | Standalone error boundary with own palette |
| `theme.d.ts` | TypeScript augmentations for custom palette extensions |
| `src/styles/utils.scss` | Layout utility classes only (zero colors) |
