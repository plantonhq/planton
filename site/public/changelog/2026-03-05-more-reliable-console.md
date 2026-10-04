---
title: "More Reliable Console"
date: 2026-03-05
category: fix
tags:
  - console
excerpt: "Fixed several crash-causing bugs and added a recovery screen so the console never shows a blank white page again."
author:
  - name: Swarup Donepudi
    title: Founder
---

Over the past few days, we identified and fixed several issues that were causing the web console to crash or show a blank white page. We also made structural improvements to how the console handles unexpected errors, so that even when something does go wrong, you'll see a helpful recovery screen instead of losing all context.

## What Was Happening

Three separate issues were causing console crashes:

**Background services breaking page loads.** The console's internal tracing system had a bug where, when tracing was disabled (which is the default configuration), every data request would silently return a dummy object instead of the actual response. This meant every page that loaded any data — the dashboard, settings, connections, environments — would crash with a confusing error that appeared to come from the page itself, not the tracing layer. This is now fixed: the tracing system correctly passes through all data requests regardless of whether tracing is enabled.

**Configuration pages failing for organization members.** Three pages in the Config Manager section — Environments > Secrets, Settings > Secrets, and Settings > Secret Backends — were showing a full-screen "Something Went Wrong" dialog when they loaded. The underlying cause was that these pages were using an internal API that required platform-operator permissions, which regular organization members don't have. These pages now use a purpose-built API that works correctly for all organization members, and data loading errors are shown as inline messages with a retry button rather than a disruptive full-screen dialog.

**A startup crash preventing the console from loading at all.** A code dependency issue caused the console to crash during its initial JavaScript loading phase — before any UI could render. This was a particularly tricky bug because it happened at import time, meaning none of the console's error handling could catch it. The fix ensures that the problematic code is evaluated lazily (when it's actually needed) rather than eagerly during startup.

## Better Error Recovery

Beyond fixing the specific crashes, we made a structural improvement to how the console handles unexpected errors at every level:

**The console now has a three-layer safety net.** If a page component crashes, the page-level error boundary catches it and shows a recovery screen within the normal console layout (sidebar, header still visible). If something in the layout itself crashes (sidebar, header, navigation), a wider boundary catches it and shows a full-screen recovery screen. And if something truly catastrophic happens during the initial page load, a self-contained fallback page renders with recovery options — no dependencies on any console infrastructure.

**The recovery screen is actionable.** Instead of a blank white page or a generic "Application error" message, you'll see clear options: retry the current page, ask for help on Discord, or report an issue on GitHub. Technical details are available in an expandable section for debugging.

**Dashboard components are more resilient.** Several dashboard widgets were crashing when API responses were incomplete or still loading. These now handle missing data gracefully, showing loading states instead of crashing.

## Catalog Page Fixes

The Infra Catalog had two visual issues:

- **Provider icons** on catalog detail pages were broken across multiple providers because they pointed to outdated asset URLs. A new fallback system tries three sources in order — built-in icons, known CDN URLs, and stored URLs — so icons display correctly for all 14 providers.
- **DigitalOcean and OpenFGA kinds** (18 total) were showing empty pages because a path resolution issue prevented their content from loading. This is fixed, and all 283 catalog kinds now display their full content including schemas, IaC modules, and presets.
