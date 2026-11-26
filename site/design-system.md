# The Website's Design System

The visual law of planton.ai: the site in `site/` and its shell package, `packages/website-shell`. The independent design reviewer grades every page of the site against this file. The console and the CLI have their own law, in the private repository's build skill (`ux/design-system/`); this file governs the website alone.

Every rule here is true in the code today, and every pattern in the code traces back here. When they disagree, the code that a founder-approved page shipped wins and this file is corrected the same day.

---

## The Idea

The website is a calm, light page that tells one story and shows the product working. The product's own artifacts (a setup prompt, a trust policy, a permission list) do the persuading, so the page around them stays quiet.

- **Light, one appearance.** Every public page renders on the same light canvas. There is no dark mode and no browser-preference switch on the website (`MarketingShell`, `HomepageAppearance`).
- **Monochrome chrome.** Header, footer, buttons, links, tables and code panels are ink on a warm off-white. Color appears only where it carries meaning: a status, a provider's logo, the dark product diagram in Home's hero.
- **Hierarchy by size and ink, not weight.** Headings are large and regular (400), never bold. Weight is reserved for small labels.
- **Borders, not backgrounds.** Sections are separated by one hairline (`--hp-line`). Panels sit one step darker than the canvas, with a border.
- **The product is the illustration.** Real artifacts are shown flat, at full size, in a code panel or a table, from a demo company, never a customer. A drawn diagram is allowed only when it is labelled as an explanation, as Home's team diagram is.

---

## Color

### The palette

One record defines the website's colors: `src/theme/homepage.ts` (`homepageLightTokens`). `MarketingShell` publishes it as CSS variables to every page (`--marketing-*`, read by the Tailwind roles `bg-canvas`, `text-fg`, `border-edge`; `--website-*`, read by the shell), and `HomepageAppearance` publishes the `--hp-*` variables the story pages' CSS modules read.

| Role | Hex | `--hp-*` | Use |
|---|---|---|---|
| Canvas | `#f6f6f3` | `--hp-canvas` | Page, header and footer background |
| Panel | `#eeeeeb` | `--hp-panel` | Code panels, tables, cards |
| Raised | `#e3e3e0` | `--hp-raised` | Hover, inline code, tags |
| Ink | `#171717` | `--hp-ink` | Headings, links, table keys, the primary button |
| Body | `#454545` | `--hp-body` | Paragraphs, table cells, code |
| Secondary | `#595959` | `--hp-secondary` | Ledes, captions, labels, table headers |
| Muted | `#616161` | | Fine print |
| Line | `#c7c7c4` | `--hp-line` | Every border and section rule |
| Line (hover) | `#999996` | | Hovered borders |
| Button | `#171717` on `#ffffff` text | `--hp-button`, `--hp-button-text` | The one primary door (Start Free) |

### Meaning, never decoration

| Meaning | Hex |
|---|---|
| Success | `#167044` |
| Danger | `#b42318` |
| Warning | `#875200` |

Info is ink or secondary gray, never blue. The docs' inline links (`#2457a6`, `src/theme/docs.ts`) are the one blue on the site: in long prose a link must be findable without hovering.

### The shell's own appearance

`packages/website-shell/src/theme/tokens.ts` holds a dark palette. It is the shell's default for a host that passes no theme of its own (the console's signed-out pages, which carry the shell from npm). The website always passes its light theme (`src/theme/light-website.ts`), so the dark values never paint planton.ai; they are fallbacks in `tailwind.config.ts`. One design system, two appearances: the website's light one is the only one this file governs.

### Provider logos

Provider logos (AWS, Google Cloud, Azure, Cloudflare, DigitalOcean, GitHub) keep their brand colors. They identify, so they are never turned monochrome.

---

## Type

### One typeface

**Inter**, loaded by `next/font/google` in `src/app/layout.tsx` at weights 300 to 700, exposed as `--font-inter`. Code, commands and policies use the system monospace stack (`ui-monospace, SFMono-Regular, Menlo, monospace`).

### Weights

| Weight | Use |
|---|---|
| 400 | Every heading (the page title, section headings), body text, table cells |
| 500 | Buttons, table keys and headers, the header's links |
| 600 | Short emphasis inside a block (a list's title, the bold words of the setup prompt) |

300 and 700 are loaded but used nowhere on the story pages. Pricing, Desktop and the 404 still render the older shared hero (`components/marketing/page-hero.tsx`), whose title is bold; they take this scale when they are next rebuilt.

### The scale (Home and Security)

From `components/marketing/homepage.module.css`, which the Security page composes rather than copies:

| Element | Size | Line height | Tracking |
|---|---|---|---|
| Page title (h1) | `clamp(44px, 4.4vw, 65px)` | 1.09 | -0.055em |
| Section heading (h2) | `clamp(32px, 3.2vw, 46px)` (Security's answers: `clamp(28px, 2.6vw, 38px)`) | 1.18 | -0.045em |
| Lede | 17px | 1.8 | |
| Body | 16px | 1.6 to 1.75 | |
| Caption, label, table header | 13px | 1.6 | |
| Code | 13px (11.5px on a phone) | 1.62 | |

Headings use `text-wrap: balance`; paragraphs use `text-wrap: pretty`.

---

## Layout

- **One container.** `max-width: 1280px` plus a gutter of `clamp(20px, 4vw, 64px)` on each side (`.container`).
- **Two columns, one line.** A story section puts its words left and its artifact right on `1fr 1.12fr` with a 38px gap, so every right-hand column starts on one vertical line down the page. Below 1100px the columns stack. A section whose words come in two parts, the claim with what was asked and the proof under it (`.lead`, `.proof` in `.split`), keeps the artifact beside both; on one column the artifact follows the claim it answers, then the proof.
- **Sections** open with a hairline rule and breathe with 56 to 80px of block padding (44 to 60px on a phone).

---

## Components

| Pattern | How it looks | Where |
|---|---|---|
| Primary door | Ink fill, white text, 8px radius, 52px tall, a `↗` | `ExperienceLink primary` |
| Secondary door | Ink text, no box, underline on hover, a `↗` | `ExperienceLink`, `DemoLink` |
| Code panel | Panel fill, line border, 12px radius, 20 × 22px padding, body-colored monospace; lines and indentation kept, a longer line scrolls inside the panel, or, for a file a person reads (`wrap`), wraps where the line's own text starts (past a `# `, `$ `, a CLI's `✓ `, `✔ ` or `ℹ `, `- ` or `key: `). A long command keeps the shell's own `\` continuations, never a wrap that splits a flag. A CLI's status line wraps only between words: an identifier such as `keycloak-operator` or an arrow `->` never breaks at its hyphen. Deeply nested config (a value several keys in) scrolls rather than hanging far right | `.prompt` (Home), composed as `.code` in `Artifact` |
| Table | Panel fill, line border, 12px radius, header row in secondary at 500, the first column in ink at 500 | `Artifact` (a table record) |
| Artifact label | 13px secondary caption above the artifact, saying what it is and where it comes from | `Artifact`'s `figcaption` |
| A check | 14px secondary line: either a link in ink, underlined, or a sentence over a command in monospace (kept whole on one line; a long one scrolls); one check per line | `Check` |
| Cloud picker | One row of segments in a line border with the controls' 8px radius, 14px at 500 in secondary; the chosen one in ink on the panel fill; native radios, a visible focus ring | `CloudPicker`, first in the first section that differs by cloud, so the cloud is chosen before any artifact; every section below reads the same choice |
| Quoted ask | What a person typed to their agent: a 13px secondary line naming who, then their words at 15px in ink, behind one left hairline; words, never a second code panel | `.ask` (Home) |

`Artifact` and `Check` (`components/marketing/Artifact.tsx`, records typed in `src/data/artifact.ts`) are how every story page shows a real thing beside its words; a page never draws its own panel, table or check.

---

## The docs

Docs pages share the site's light canvas, text and border colors through the class strings in `src/theme/docs.ts`: links, tags, callouts, inline code, code blocks, tables and the sidebar. Every docs component imports from that file and types no color of its own. Code blocks and Mermaid diagrams are contained visual treatments: they may keep muted syntax hues so tokens stay distinguishable.

---

## Rules for contributors

**Always:**
1. Read every color from the palette's roles (`--hp-*`, `--marketing-*`, the Tailwind roles, or `docs.ts`); a component never types a hex.
2. Put every sentence in `src/data/`, naming its chapter; components render records and state nothing.
3. Show the real artifact, flat and readable, with a label saying what it is.
4. Keep the two-column line: an answer without an artifact shows its command where the artifact would be.
5. Compose a shared style from `homepage.module.css` with `composes:`; never copy its rules.

**Never:**
1. A bold heading on a story page.
2. A gradient, a tinted card, a colored header or footer, or a hue that carries no meaning.
3. A dark section on a light page, except the product diagram Home's hero shows.
4. A tilted, blurred or redrawn product screenshot.
5. A second palette file for the website.

---

## Key files

| File | What it holds |
|---|---|
| `src/theme/homepage.ts` | The website's palette, and the CSS variables `MarketingShell` publishes |
| `src/theme/light-website.ts` | The MUI theme the shell renders with on the website |
| `src/theme/docs.ts` | The docs' class strings |
| `src/components/marketing/MarketingShell.tsx` | Every public page's light appearance, header and footer |
| `src/components/marketing/HomepageAppearance.tsx` | The `--hp-*` variables for the story pages |
| `src/components/marketing/homepage.module.css` | Home's type scale, container, code panel and doors |
| `src/components/security/security.module.css` | The Security page, composed from Home's module |
| `packages/website-shell/src/theme/tokens.ts` | The shell's default (dark) appearance, and the Tailwind roles |
| `packages/website-shell/src/data/navigation.ts` | The header's and footer's links |
| `src/app/layout.tsx` | Inter, the root metadata |
