# Public Subnet Worker

This preset places the runner on public subnets with a public IPv4
address -- the right posture for VPCs that have an internet gateway but
no NAT gateway. The public IP exists purely so the runner can pull its
image and dial out to the control plane; the security group the
deployment creates refuses all ingress, so nothing on the internet can
reach the runner.

## When to Use

- The VPC has no NAT gateway and you don't want to pay for one just to
  host a runner (NAT is a per-hour, per-GB cost)
- Sandbox or simple VPCs where all subnets route to an internet gateway
- The in-network targets are reachable from the public subnets (same
  VPC, routing permitting)

## Key Configuration Choices

- **`assignPublicIp: true`** -- required on public subnets without NAT;
  without it the runner cannot pull its image or reach the control plane
  and never starts
- **Still zero inbound exposure** -- a public IP is not an inbound door:
  the runner's security group allows no ingress at all, and the runner
  only ever dials out
- **Token as a managed-secret reference** -- the token authorizes
  joining only; the runner registers itself on first boot and receives
  its own individually revocable identity, and revoking the token never
  touches runners it already admitted
- **Two public subnets in different AZs** -- the runner reschedules
  across an AZ event

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<runner-name>` | Name for the runner appliance | Any name you choose |
| `<aws-region>` | AWS region code | The region hosting the targets |
| `<public-subnet-a-resource-name>` / `<public-subnet-b-resource-name>` | Names of the public AwsSubnet resources | Your subnet manifests' `metadata.name` |

## Related Presets

- `01-private-vpc-worker` -- the recommended posture when a NAT route exists
- `03-high-capacity` -- pinned version and larger sizing for heavy IaC workloads
