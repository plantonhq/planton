# GcpPscServiceAttachment — Terraform Implementation

This directory contains the Terraform implementation for publishing a
service through Private Service Connect from the Planton spec: one
`google_compute_service_attachment` in front of the producer's internal load
balancer.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, attachment name (spec or metadata.name fallback), the accept list with its three optional arms, the tri-state propagated limit and its send-if-zero twin |
| `main.tf` | `google_compute_service_attachment` |
| `outputs.tf` | `self_link`, `attachment_name`, `region`, `fingerprint`, `connected_endpoints_count` |

## Send Posture

- **`target_service`, `nat_subnets`** -- the flattened references (a
  forwarding rule self link, subnet self links); the provider stores
  `nat_subnets` as a set, so order never diffs.
- **`consumer_accept_lists`** -- each entry names exactly one of
  `project_id_or_num`, `network_url`, `endpoint_url` (the spec's
  `projectId` / `network` / `endpointUrl`); the unset arms are sent as
  null, never as the provider's `""` default.
- **`reconcile_connections`** -- Optional+Computed on the provider; sent
  only when the spec sets it, so Google's default is never fought.
- **`propagated_connection_limit`** -- tri-state: null lets Google apply
  its default of 250; an explicit 0 is sent together with
  `send_propagated_connection_limit_if_zero = true`, the provider's own
  mechanism for transmitting a zero -- derived in `locals.tf`, never a spec
  field (PARITY with the Pulumi module).
- **`enable_proxy_protocol`** -- required by the API; the proto default
  `false` is a real answer and is always sent.
- **`show_nat_ips`** -- sent only when true; Google's API currently ignores
  the flag.
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON decides
  what a destroy does to a published service.
