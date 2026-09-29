# Private VPC Worker

This preset deploys the standard runner appliance: an always-on worker on
two private subnets that receives deploy operations from the control
plane and executes them from inside the VPC. The 30-second decision for
making a private-endpoint Kubernetes cluster (or any in-network target)
deployable.

## When to Use

- A Kubernetes cluster with a private API endpoint needs to be deployed
  to and operated
- Any deployment target reachable only from inside the VPC (private
  databases, internal services)
- You want zero inbound network exposure: the runner only dials out

## Key Configuration Choices

- **Token as a managed-secret reference** -- the token authorizes the
  runner to JOIN, nothing more: on first boot the runner registers
  itself with the control plane and receives its own individually
  revocable identity, and revoking the token never touches runners it
  already admitted. On Planton the platform mints the token and writes
  it at exactly this reference before the infrastructure applies.
- **No mode or replica knobs** -- everything beyond the join (work
  queue, tunnel, API endpoints) arrives in the join response, so the
  runner self-configures on arrival; and it runs as exactly one instance
  by design -- more capacity means more runners, not more copies
- **Two private subnets in different AZs** -- the runner reschedules
  across an AZ event; private subnets need a NAT route for the runner's
  outbound traffic (image pulls, control-plane dial-out)
- **Default sizing (0.5 vCPU / 1 GiB)** -- comfortable for typical IaC
  operations; see the high-capacity preset when stacks grow large

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<runner-name>` | Name for the runner appliance | Any name you choose |
| `<aws-region>` | AWS region code | The region hosting the private targets |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the private AwsSubnet resources | Your subnet manifests' `metadata.name` |

The `runner-token` secret slug is yours to choose -- on Planton the
platform writes the token there automatically; elsewhere, create a token
with `planton runner token create` and store it under that slug.

## Related Presets

- `02-public-subnet` -- same worker on public subnets, for VPCs without NAT
- `03-high-capacity` -- pinned version and larger sizing for heavy IaC workloads
