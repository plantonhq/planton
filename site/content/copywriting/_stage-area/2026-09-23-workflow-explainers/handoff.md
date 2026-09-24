# Homepage workflow explainers — implementation handoff

## Current scope

Five illustrative customer architectures cycle AWS → GCP → Azure → Cloudflare → DigitalOcean. Delivery and Coding Agent are independent six-stage explainers. All seven stories last 28 seconds: six four-second phases plus a four-second completed hold. This supersedes the earlier AWS study and internal hosted-cluster example recorded in draft-1.md.

Public copy and inventories live in `src/data/architecture-stories.ts` and `workflow-explainers.ts`. Customer labels lead; technology names follow. Infrastructure arrows describe deployment dependencies, not runtime traffic. Document ingestion, application code, vLLM configuration, and VM bootstrap remain customer inputs.

## Rendering

Pure shared SVG scenes consume story and time. Browser playback and the export-only Remotion adapter supply time. Explicit keyed paths keep ports and reserved lanes separate from resource relationships. Rounded paths share arc-length sampling; blue packets travel at 90 scene units per second. Sequence diagrams use aligned three-column rows and one outside return.

The icon registry embeds selected canonical public-repo SVG files. ResourceIcon frames them in quiet dark category tints, preserving glyph colors. No private-repo assets or runtime catalog requests are introduced. The five-provider selector is centered on desktop and scrollable on narrow screens.

Only the selected provider advances. Completion drives selection after the final hold. Manual selection, phase inspection, disclosure, and keyboard focus disable cycling; Pause survives selection. Offscreen and hidden-page time suspend playback. Following preview feedback, hover does not pause: visitors use Pause/Play or phase inspection. Reduced motion completes diagrams statically; no JavaScript exposes all five architectures and both sequence transcripts. Outgoing panels are inert and cleaned up independently of interrupted CSS events.

## Source grounding

All displayed kinds exist in this checkout's catalog. Inventories distinguish typed references from explicit ordering and application configuration. DigitalOcean uses tagged autoscale members, managed PostgreSQL, public Spaces assets, and a CDN. Certificate/domain setup, scoped data access, and firewalls are visible in its inventory. No cloud infrastructure was deployed during website implementation.

## Verification and delivery

Run the site build with the private platform docs root configured. Timeline checks validate phases, gates, catalog identity, inventories, and route/card intersections. Browser checks cover seven diagrams, provider completion and wraparound, interruption, pause/replay, focus, hover, hidden/offscreen time, reduced motion, static fallback, and responsive screenshots. Homepage/booking tests mock outbound submissions.

Desktop captures must show closed components below navigation at 1366×768, 1440×900, and 1920×1080. Mobile layouts retain readable type and natural scrolling. Inspect initial, transferring, and completed states and provider inventories. Export all seven 1080-square H.264 videos; generated media stays outside Git and R2.

User authorized PR creation and merge without waiting for GitHub checks. Finish local checks and show refreshed preview first, then create/attach/merge the PR and verify deployment. No independent usability study or conversion uplift is claimed.
