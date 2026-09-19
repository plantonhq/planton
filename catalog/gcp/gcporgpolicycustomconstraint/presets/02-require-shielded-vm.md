# Require Shielded VM

An `ALLOW` constraint over Compute instances: only instances with Secure
Boot on are acceptable; everything else is implicitly denied. The
allow-list form of a custom constraint.

## What it configures

- `resourceTypes: [compute.googleapis.com/Instance]`.
- `condition` — describes the ONLY acceptable shape.
- `actionType: ALLOW` — permit when true, deny otherwise.

## Adjust before deploying

- **`organizationId`** — yours.
- **`condition`** — tighten to `&& resource.shieldedInstanceConfig.enableVtpm == true`
  for the full Shielded VM posture.

## When to choose something else

A rule that names the forbidden shape takes the **Deny Public GKE Nodes**
preset; most guardrails are `DENY`.
