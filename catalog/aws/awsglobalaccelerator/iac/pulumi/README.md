# Pulumi Module: AWS Global Accelerator

Provisions an AWS Global Accelerator — the anycast traffic director — using
Pulumi (Go), as one bundled family: the accelerator, its listeners, and
each listener's endpoint groups.

## Resources Created

- `globalaccelerator.Accelerator` — the accelerator with its two static
  anycast IPs (AWS-allocated or BYOIP), IPv4 or dual-stack, and the
  always-materialized attributes block carrying the flow-logs switch.
- `globalaccelerator.Listener` — one per named listener in the spec:
  TCP/UDP port ranges with client affinity.
- `globalaccelerator.EndpointGroup` — one per endpoint group under each
  listener: per-region traffic dials, endpoint weights, client-IP
  preservation, health-check tuning, and port overrides.

## How It Works

The module receives an `AwsGlobalAcceleratorIacInput` (the manifest plus
provider credentials), builds the AWS provider through the shared builder,
and renders the family from the spec. Send conditions match the Terraform
module argument-for-argument: presence-honest optionals pass through only
when set (provider defaults: enabled=true, IPV4), and the accelerator
attributes block is always sent with an explicit `flow_logs_enabled` so
flow logs can be turned off once on.

The attributes block uses the value form of `AcceleratorAttributesArgs`
(not the generated `...Ptr(...)` constructor): the pointer form compiles
clean but trips a runtime marshaling assertion at deploy time.

Listeners and endpoint groups are keyed by their spec names, and their
ARNs are exported as maps keyed the same way (`listener_arns`,
`endpoint_group_arns` keyed `listener_name/group_name`).

## Outputs

| Name | Description |
|------|-------------|
| `accelerator_arn` | ARN of the accelerator |
| `accelerator_dns_name` | DNS name fronting the static IPv4 pair |
| `accelerator_dual_stack_dns_name` | DNS name for dual-stack accelerators |
| `accelerator_hosted_zone_id` | Route 53 alias hosted-zone ID |
| `accelerator_ip_addresses` | The static anycast IP addresses |
| `listener_arns` | Map of listener name → listener ARN |
| `endpoint_group_arns` | Map of `listener_name/group_name` → endpoint group ARN |
