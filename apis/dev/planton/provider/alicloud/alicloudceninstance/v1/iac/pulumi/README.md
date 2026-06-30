# Pulumi Module to Deploy AliCloudCenInstance

This Pulumi program provisions an Alibaba Cloud CEN (Cloud Enterprise Network)
instance with bundled child-instance attachments. CEN provides private
connectivity between VPCs, VBRs, and CCNs across any Alibaba Cloud region.

## Resources Created

- `cen.Instance` — the CEN hub instance
- `cen.InstanceAttachment` × N — one per entry in `spec.attachments[]`

## CLI Usage (Planton Pulumi)

```bash
# Preview
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Module Structure

| File | Purpose |
|------|---------|
| `main.go` | Pulumi program entrypoint; loads stack input and delegates to module |
| `Pulumi.yaml` | Project configuration |
| `module/main.go` | CEN instance creation + attachment loop; exports outputs |
| `module/locals.go` | Tag computation from metadata |
| `module/outputs.go` | Output constant definitions (`cen_id`, `cen_instance_name`) |

## How It Works

1. The entrypoint (`main.go`) loads the `AliCloudCenInstanceStackInput` from Pulumi config
2. `locals.go` computes tags from metadata fields (name, id, org, env, resource_kind)
3. `main.go` creates the AliCloud provider scoped to `spec.region` (used for API routing only)
4. The CEN instance is created with name, description, optional protection level, and tags
5. Each attachment in `spec.attachments[]` creates a `cen.InstanceAttachment` as a child of the CEN instance
6. Stack outputs are exported: `cen_id` and `cen_instance_name`

For more details on module architecture, see [`overview.md`](./overview.md).
