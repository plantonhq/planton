---
title: "Audit Trails"
description: "How Planton tracks every resource change with immutable version records, full-state snapshots, unified diffs, and authorization-gated access"
icon: history
order: 30
tags:
  - Audit
  - Versions
  - Change Tracking
  - Compliance
---

# Audit Trails

Every change to a resource in Planton — an Infra Component deployment, a Service configuration update, a connection modification — is captured as an immutable version record. These records form a complete, append-only history of every resource in your organization.

This is not a simple log of "who clicked what." Each version record captures the full state of the resource before and after the change, a unified diff showing exactly what changed, who made the change, and whether the resulting deployment succeeded.

## What Gets Captured

Each version record contains:

| Field | What It Captures |
|-------|-----------------|
| **New state** | Complete YAML snapshot of the resource after the change |
| **Original state** | Complete YAML snapshot before the change (empty for the first version) |
| **Unified diff** | Git-style diff showing additions, removals, and modifications with surrounding context |
| **Event type** | Whether the resource was **created**, **updated**, **deleted**, or **restored** |
| **Who** | The identity account that initiated the change |
| **When** | Timestamp of when the version was recorded |
| **Infra Job link** | For infrastructure deployments, the Infra Job that executed the change |
| **Infra Job status** | Whether the deployment is queued, running, succeeded, or failed |
| **Previous version** | Link to the preceding version, forming a chronological chain |

### Infrastructure Manifest Diffs

For Infra Components, version records include an additional focused diff that isolates the infrastructure manifest from the surrounding platform metadata. This means you can see exactly what changed in the deployment definition (the VPC configuration, the database parameters, the cluster settings) without wading through platform bookkeeping fields.

The web console provides a tabbed view: **Manifest** shows only the infrastructure changes, **Wrapper** shows the full resource including platform metadata.

## Immutability

Version records are append-only. Once a version is created, it cannot be modified or deleted. The only field that updates after creation is the Infra Job progress status — which tracks whether the deployment triggered by the change has succeeded or failed. The version content itself (states, diffs, identity, timestamps) is immutable.

This means the audit trail cannot be retroactively altered. If a resource was modified at 3:42 PM on Tuesday by user Alice, that fact is permanently recorded regardless of what happens afterward.

## Viewing Audit Data

### Versions List

Navigate to any resource's detail page in the web console and open the **Versions** tab. The versions list shows all recorded changes, newest first, with:

- Version name (clickable — opens the diff view)
- Who made the change (avatar and name)
- Event type and timestamp
- Infra Job progress icon (queued, running, succeeded, failed)
- Version ID

<!-- SCREENSHOT: Resource versions list
  Page: /orgs/{org}/infra-component/{env}/{kind}/{name}/versions
  Action: Show the versions tab for an infra component with at least 3 version entries
  Focus: The versions list showing version names, users, event types, and status icons
  Alt: Resource versions list showing chronological change history with user avatars, event types, and deployment status indicators
-->

### Version Detail

Click any version to see the complete YAML representation of the resource at that point in time. This is the full snapshot — not a diff, but the entire resource state. Useful for understanding what a resource looked like at a specific moment.

<!-- SCREENSHOT: Version detail view
  Page: /orgs/{org}/infra-component/{env}/{kind}/{name}/versions/{versionId}
  Action: Show the full version detail with YAML content
  Focus: The YAML viewer showing the complete resource state
  Alt: Version detail view showing the full YAML resource state at a specific point in time
-->

### Diff View

The diff view is the primary tool for understanding what changed between versions. It displays a unified diff showing additions (green), removals (red), and context lines.

Two display modes are available:

- **Line-by-line** — Traditional unified diff format, one column
- **Side-by-side** — Before and after shown in parallel columns

For Infra Components, vertical tabs switch between the **Manifest** diff (infrastructure definition only) and the **Wrapper** diff (full resource including platform metadata). The Manifest tab focuses your attention on the infrastructure changes that matter.

<!-- SCREENSHOT: Diff view with side-by-side comparison
  Page: /orgs/{org}/infra-component/{env}/{kind}/{name}/versions/{versionId}/diff
  Action: Show the diff view in side-by-side mode with visible additions and removals
  Focus: The diff display showing changed lines highlighted in green and red
  Alt: Version diff view showing side-by-side comparison of resource changes with additions highlighted in green and removals in red
-->

## Access Control

Audit data is authorization-gated. You can only view version records for resources you have permission to see. If you have viewer access to an environment, you can view versions of all resources in that environment. If you only have access to a specific service, you can only view that service's version history.

This means the audit trail respects the same permission model as the rest of the platform. See [Authentication and Authorization](/docs/security/authentication-and-authorization) for how permissions are evaluated.

## How Versions Are Created

Version records are created automatically. When a resource changes — through the web console, CLI, or API — the platform publishes an event. The audit system processes that event and creates the version record, including computing the diff and capturing the full state.

There is no manual step to "enable auditing" or "create a version." Every change to a versioned resource is captured. This ensures the audit trail is complete without depending on users or automation to remember to log changes.

## Practical Uses

### Debugging Failed Deployments

When an infrastructure deployment fails, the version record shows exactly what configuration change was attempted. The Infra Job link takes you to the deployment logs. Together, these answer the two critical debugging questions: "what changed?" and "what happened when we tried to apply it?"

### Change Review

Before a deployment runs, the diff shows what will change. After it runs, the version record confirms what actually changed. This provides a complete audit chain for compliance reviews — every change has a recorded intention and result.

### Rollback Investigation

Version records do not provide a rollback mechanism directly, but they provide the data needed to rollback manually. By viewing a previous version's full state, you can see exactly what the resource looked like before a problematic change and use that information to create a corrective update.

### Compliance Reporting

The combination of immutable records, identity tracking, and timestamp logging provides the raw material for compliance audits. Every resource change is attributable to a specific user at a specific time, with full before-and-after state capture.

## Related Documentation

- [Security Overview](/docs/security) — Platform security model and architecture
- [Authentication and Authorization](/docs/security/authentication-and-authorization) — How permissions control access to audit data
- [Infra Jobs](/docs/infrastructure/infra-jobs) — Infrastructure deployment execution linked from version records
- [Teams and Access](/docs/teams-and-access) — Role-based access that gates audit data visibility
