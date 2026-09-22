# GcpPscServiceAttachment — Pulumi Implementation

This directory contains the Pulumi implementation for publishing a service
through Private Service Connect from the Planton spec: one
`gcp.compute.ServiceAttachment` in front of the producer's internal load
balancer.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `service_attachment` |
| `module/locals.go` | Ambient project fallback, attachment name (spec or metadata.name fallback) |
| `module/service_attachment.go` | Maps spec to `gcp.compute.ServiceAttachment`; exports the outputs |
| `module/outputs.go` | Output key constants (`self_link`, `attachment_name`, `region`, `fingerprint`, `connected_endpoints_count`) |

## Send Posture (parity with Terraform)

- **`TargetService`, `NatSubnets`** -- the resolved reference values (a
  forwarding rule self link, subnet self links).
- **`ConsumerAcceptLists`** -- each entry names exactly one of
  `ProjectIdOrNum`, `NetworkUrl`, `EndpointUrl` (the spec's `projectId` /
  `network` / `endpointUrl`); unset arms are left nil.
- **`ReconcileConnections`** -- sent only when the spec sets it
  (Optional+Computed on the provider).
- **`PropagatedConnectionLimit`** -- tri-state: nil lets Google apply its
  default of 250; an explicit 0 is sent together with
  `SendPropagatedConnectionLimitIfZero = true`, derived from the spec value
  exactly as the Terraform module does.
- **`EnableProxyProtocol`** -- required by the API; always sent.
- **`ShowNatIps`** -- sent only when true.
- **`DeletionPolicy`** -- DELETE (default), PREVENT, or ABANDON; sent only
  when set on both engines.
- **`connected_endpoints_count`** -- derived from the resource's
  `ConnectedEndpoints` output as a string, the shape the Terraform module
  exports.
