# GcpNetworkEndpointGroup — Terraform Implementation

This directory contains the Terraform implementation for provisioning a
network endpoint group from the Planton spec -- zonal when `spec.zone` is
set, global when it is empty -- together with its membership.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, group name (spec or metadata.name fallback), the `is_zonal` scope selector, the endpoint lists shaped for each collection |
| `main.tf` | `google_compute_network_endpoint_group` + bulk `google_compute_network_endpoints` (zonal) or `google_compute_global_network_endpoint_group` + per-member `google_compute_global_network_endpoint` (global); exactly one pair created by count / for_each guards |
| `outputs.tf` | `self_link`, `neg_name`, `neg_id` selected from whichever arm exists; `zone`; `size` |

## One Kind, Two Scopes

`spec.zone` selects the API collection. A zone name builds the zonal group
and writes its membership as ONE set through Google's bulk endpoint
operation (`google_compute_network_endpoints`, created only when the list
is non-empty), so the manifest's list is the group's whole membership and
changes in place. An empty zone builds the global internet group with one
`google_compute_global_network_endpoint` per member, keyed by list
position (`for_each`); every global endpoint argument is immutable, so a
changed member is replaced.

## Send Posture

- **`network_endpoint_type`** -- the zonal resource defaults it to
  `GCE_VM_IP_PORT` itself, so an empty spec value is sent as null; the
  global resource requires it (the spec guarantees it is set there).
- **`default_port`, endpoint `port`** -- tri-state (`optional` in the
  spec): null sends nothing. A global endpoint's port falls back to
  `default_port`, because the global endpoint resource requires one.
- **`instance`, `ip_address`, `fqdn`** -- empty strings mean "not this
  arm" and are sent as null; the spec has already matched each to the
  group's type.
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON, fanned
  out to the endpoint resources so the membership shares the group's fate.
