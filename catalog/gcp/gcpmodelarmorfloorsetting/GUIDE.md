# GcpModelArmorFloorSetting Guide

The judgment this guide protects: a floor outlives this block -- Google cannot delete one -- and a floor that blocks before it has been tuned breaks every AI call in its scope. Roll a floor out inspect-only, raise it deliberately, and relax it by applying, never by destroying.

## What a floor does

A floor setting belongs to a project, a folder, or the organization, and does two things:

1. **It governs templates.** `filterConfig` is the weakest filter set any Model Armor template in the scope may carry; once `enableFloorSettingEnforcement` is on, a template below it is flagged or refused.
2. **It screens Google's AI services directly.** With `AI_PLATFORM` in `integratedServices`, every Vertex AI model call in scope goes through the floor's filters; `GOOGLE_MCP_SERVER` does the same for Google-hosted MCP servers. Each service's setting chooses `INSPECT_ONLY` (record the verdict, let the call through) or `INSPECT_AND_BLOCK`, and whether verdicts go to Cloud Logging.

This is the way to put an entire project's model traffic behind Model Armor without touching application code.

## The lifecycle is unusual

Google keeps exactly one floor per project, folder, and organization. Applying this block overwrites that floor (create and update are the same call). Destroying it does not delete anything: the provider stops managing the floor and the last applied settings stay in force. So:

- To relax or retire a floor, apply it with `enableFloorSettingEnforcement: false` (and, if needed, weaker filters), then destroy.
- Never declare two blocks for the same scope -- they overwrite each other on every apply.
- The live test leaves a floor behind in the test project on purpose; it is inspect-only and looser than every test template, so it blocks nothing.

## Rolling out

1. Apply the floor with `enableFloorSettingEnforcement: true`, the integrated services you want, `INSPECT_ONLY`, and `enableCloudLogging: true`.
2. Watch the logged verdicts: what would have been blocked, and which templates fall below the floor.
3. Raise templates that fall below; tune the floor's thresholds.
4. Switch the services to `INSPECT_AND_BLOCK`.

## Scope and inheritance

A folder floor is inherited by every project and folder beneath it, an organization floor by everything. Put the baseline at the organization and tighten per folder or project. `scope` takes one of `projectId` (a `GcpProject` reference), `folderId` (a `GcpFolder` reference), or the numeric `organizationId`; empty means the connection's project. Folder and organization floors need the floor-setting role at that level.

## Filters

`filterConfig` has the same shape as a template's: prompt injection and jailbreak, Responsible AI categories, Sensitive Data Protection (basic or your own templates), and malicious URLs. Confidence levels here are the loosest a template may use -- `HIGH` is the most permissive floor, `LOW_AND_ABOVE` the strictest.
