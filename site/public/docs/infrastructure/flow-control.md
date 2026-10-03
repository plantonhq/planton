---
title: "Flow Control"
description: "Governance policies that control how Infra Jobs execute — manual approvals, preview requirements, and deployment pauses."
icon: guide
order: 60
tags:
  - Infrastructure
  - Flow Control
  - Governance
---

# Flow Control

A Flow Control policy defines the rules for how [Infra Jobs](/docs/infrastructure/infra-jobs) execute. It controls whether jobs require manual approval, whether a preview step runs before every apply, and whether the system pauses between preview and apply to give you a chance to review the plan. Different environments need different levels of rigor — development should be fast, production should be careful — and Flow Control makes that distinction automatic.

## Why Flow Control Exists

Without governance, every infrastructure change follows the same path: submit configuration, run the IaC operations, apply the result. That works for development, but production deserves more care. You want to see the plan before it runs. You want someone to approve destructive changes. You might want to skip the refresh step for resources where drift is not a concern and speed matters.

Flow Control policies encode these decisions into the platform. Define the rules once, attach them to the right scope, and every Infra Job in that scope automatically follows them. No per-deployment configuration, no manual steps to remember, no governance gaps.

## The Five Controls

Every Flow Control policy consists of five settings. Each is a boolean toggle that modifies how Infra Jobs execute.

### Manual Approval Required

When enabled, every Infra Job created in this scope pauses immediately after creation and waits for manual approval before any operations begin. Nothing executes until a human explicitly approves it.

This differs from "Pause After Preview" (below), which allows the preview to run automatically and only pauses before the apply step. Manual Approval Required pauses before any operation starts.

### Lifecycle Events Disabled

When enabled, Infra Jobs are not automatically created when an Infra Component is created, updated, or deleted. You can still create Infra Jobs manually through the CLI or web console, but the automatic trigger from lifecycle events is suppressed.

This is useful for shared infrastructure that multiple teams depend on, where accidental changes could have wide impact. Only explicit, intentional Infra Jobs run.

### Skip Refresh

When enabled, the refresh step is skipped during Infra Job execution. Refresh synchronizes the IaC state file with the actual cloud state, catching drift from manual changes. Skipping it saves time but means the job operates on potentially stale state.

Enable this only for resources managed exclusively through Planton with no manual cloud console access, where drift is not a concern.

### Preview Before Apply

When enabled, a preview step always runs before update or destroy operations. The preview shows exactly what will change — resources to create, modify, or delete — before anything happens. This is enabled by default.

Disabling it removes the preview step, making Infra Jobs faster but removing the safety net of seeing the plan first. Only disable this for well-understood, low-risk resources in non-production environments.

### Pause After Preview

When enabled (and Preview Before Apply is also enabled), the Infra Job pauses after the preview completes but before the apply or destroy runs. A team member must explicitly approve the previewed plan before the job continues.

This is the most common production governance pattern: let the preview run automatically so you can see the plan, then require a human to approve it before any changes are made to real infrastructure.

<!-- SCREENSHOT: Flow Control settings in Infra Job detail
  Page: /resource/infra-hub/infra-job/[infraJobId] (Config tab)
  Action: Show the Flow Control settings display with boolean indicators
  Focus: The flow control settings section with check/cross icons
  Alt: Infra Job configuration tab showing resolved flow control settings with enabled/disabled indicators for each control
-->

## Policy Resolution Hierarchy

Flow Control policies are attached to a scope using a selector. The platform resolves which policy applies to a given Infra Job by checking four levels, from most specific to least:

1. **Infra Component** — A policy targeting a specific Infra Component by ID
2. **Environment** — A policy targeting all resources in a specific environment
3. **Organization** — A policy targeting all resources in the organization
4. **Platform** — A default policy for the entire platform

The first match wins. Policies are not merged across levels — if an environment-level policy exists, the organization and platform policies are ignored entirely for resources in that environment.

If no policy matches at any level, the platform default applies.

## Common Patterns

### Development: Fast Iteration

Optimize for speed. Let changes apply immediately without manual steps:

- Manual Approval Required: **off**
- Lifecycle Events Disabled: **off**
- Skip Refresh: **off** (keep drift detection)
- Preview Before Apply: **on** (the plan still appears in logs)
- Pause After Preview: **off** (no manual approval step)

### Production: Full Control

Maximize safety. Require preview and manual approval before any change takes effect:

- Manual Approval Required: **off** (let the job start and preview automatically)
- Lifecycle Events Disabled: **off**
- Skip Refresh: **off**
- Preview Before Apply: **on**
- Pause After Preview: **on** (require human approval of the plan)

### Shared Infrastructure: Explicit Only

For resources that multiple teams depend on, disable automatic triggers entirely:

- Manual Approval Required: **on**
- Lifecycle Events Disabled: **on** (no automatic Infra Jobs)
- Skip Refresh: **off**
- Preview Before Apply: **on**
- Pause After Preview: **on**

## Using the Web Console

Flow Control settings are visible on the Infra Job detail page in the configuration tab. The display shows each of the five controls with an enabled/disabled indicator, reflecting the resolved policy for that specific job.

To create or update Flow Control policies, use the CLI (below). Policy management through the web console is planned.

## Using the CLI

Flow Control policies are managed through the standard resource commands:

```bash
# Create or update a Flow Control policy from a YAML manifest
planton apply -f flow-control-policy.yaml

# Get a Flow Control policy by ID
planton get flow-control-policy <policy-id>

# Delete a Flow Control policy
planton delete flow-control-policy <policy-id>
```

To check which Flow Control policy applies to a given resource type and environment, use the Infra Job preflight check:

```bash
planton infra-job preflight-checks --catalog-kind <kind>
```

The preflight report includes the resolved policy and its settings.

## Related Documentation

- [Infra Jobs](/docs/infrastructure/infra-jobs) — The execution units that Flow Control governs
- [Infra Components](/docs/infrastructure/infra-components) — The resources whose deployments are controlled
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — Pipeline-level approval gates for multi-resource deployments
