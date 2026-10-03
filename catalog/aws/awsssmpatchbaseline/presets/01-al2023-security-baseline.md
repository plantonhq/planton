# AL2023 Security Baseline

This preset creates the sensible Linux default: Critical and Important
security patches auto-approve seven days after release, governing one
patch group.

## When to Use

- Amazon Linux 2023 fleets that should take security patches
  automatically — after a soak period long enough for bad vendor
  patches to be pulled
- The first custom baseline in an account (AWS's predefined defaults
  approve with no soak)

## What You Get

- A days-based approval rule scoped to security Critical/Important,
  reporting missing patches as CRITICAL compliance
- The named patch group bound to this baseline (tag nodes
  `Patch Group: <name>` to opt them in)

## Customize

- `approveAfterDays: 0` takes patches on release day; `2` is a common
  fast lane for actively exploited CVEs
- Add `rejectedPatches` + `rejectedPatchesAction: BLOCK` to pin known-bad
  packages
- `setAsDefaultBaseline: true` makes this the OS default for nodes in
  no governed group (delete restores AWS's own default)
- Schedule the actual scan/install with
  [AWS SSM Association](/infra-catalog/aws-ssm-association) and
  [AWS SSM Maintenance Window](/infra-catalog/aws-ssm-maintenance-window)
