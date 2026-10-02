# GcpSccMuteConfig Guide

The judgment this guide protects: a mute rule hides findings from everyone downstream. Keep rules narrow, explain them, and prefer the dynamic kind you can take back.

## Dynamic versus static

- **`DYNAMIC`** (Google's recommendation) mutes every existing and future finding that matches. When the rule is changed or deleted, or a finding stops matching, the mute is lifted (unless another rule still matches).
- **`STATIC`** sets a permanent mute on future matching findings only. Changing or deleting the rule later does not unmute them.

Google treats the type as immutable after creation; change it by replacing the rule (a new `muteConfigId`).

## Writing the filter

Supported fields: `severity`, `category`, `resource.name`, `resource.project_name`, `resource.project_display_name`, `resource.folders.resource_folder`, `resource.parent_name`, `resource.parent_display_name`, `resource.type`, `findingClass`, `indicator.ip_addresses`, `indicator.domains` -- each with `=` or `:` (substring), combined with `AND` and `OR`. Write it for the scope: a rule scoped to one project only ever sees that project's findings.

## Downstream effect

Muted findings stay queryable. Notification configs and BigQuery exports include them unless their filter says `NOT mute = "MUTED"` -- the presets of those kinds do.
