# High Capacity (Production Hardened)

This preset is the production posture for heavy workloads: a sized-up
runner (1 vCPU / 4 GiB) with a pinned version, a first-class runtime IAM
role, and extended log retention. Everything is deliberate -- nothing
tracks `latest`, and the runner's AWS identity is a role you compose and
audit.

## When to Use

- Large stacks or high operation concurrency (memory pressure shows up
  as failed IaC operations mid-apply)
- Production environments with change control: versions are pinned and
  bumped deliberately
- Keyless cloud access through the runner: operations use the runtime
  role instead of injected keys

## Key Configuration Choices

- **`memory: 4096` before more cpu** -- IaC operations are memory-hungry
  (provider plugins, plans held in memory); memory is the dimension that
  fails first
- **`runnerVersion: v0.4.0`** -- a pinned tag (replace with the current
  release); rollback is re-pinning the previous tag
- **`taskRole` referencing your own `AwsIamRole`** -- trust policy for
  `ecs-tasks.amazonaws.com`, exactly the permissions the runner's
  workloads need, auditable in one place
- **`logRetentionDays: 90`** -- the runner's logs are the audit trail of
  every infrastructure change it executed
- **Token as a managed-secret reference** -- the token authorizes
  joining only; the runner registers itself on first boot and receives
  its own individually revocable identity, and task replacement re-joins
  with the same token

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<runner-name>` | Name for the runner appliance | Any name you choose |
| `<aws-region>` | AWS region code | The region hosting the private targets |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the private AwsSubnet resources | Your subnet manifests' `metadata.name` |
| `<runtime-role-resource-name>` | Name of the AwsIamRole resource for the runner's runtime identity | Your role manifest's `metadata.name` |

## Related Presets

- `01-private-vpc-worker` -- the default-sized starting point
- `02-public-subnet` -- same worker on public subnets, for VPCs without NAT
