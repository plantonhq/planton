# Terraform Module to Deploy AwsGlobalAccelerator

This module provisions an AWS Global Accelerator — the anycast traffic
director — as one bundled family: the accelerator, its listeners, and each
listener's endpoint groups.

## Features

- **Anycast front door**: two static anycast IPs (AWS-allocated or BYOIP
  via `ip_addresses`), IPv4 or dual-stack
- **Listeners**: TCP/UDP port ranges with client affinity, keyed by name
- **Endpoint groups**: per-region groups with traffic dials, endpoint
  weights, client-IP preservation, health-check tuning, and port
  overrides
- **Flow logs**: the accelerator `attributes` block is ALWAYS materialized
  with an explicit `flow_logs_enabled` value — flow-log settings live on a
  separate accelerator-attributes API and the provider diff-suppresses a
  missing block, so the explicit send is what lets flow logs be turned
  off once on

Global Accelerator is a global (non-regional) service: the provider homes
all API calls in us-west-2 regardless of the configured region, and every
create/update waits for the accelerator to return to DEPLOYED, so applies
take minutes. On destroy the provider first disables the accelerator (an
AWS requirement), waits, then deletes.

Generated `variables.tf` reflects the proto schema for
`AwsGlobalAccelerator`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the
manifest `spec`.

For a working example, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).
