# DigitalOcean VPC

A DigitalOcean Virtual Private Cloud described once in a Planton manifest: a private, isolated network for Droplets, Kubernetes clusters, load balancers, and databases within one region. The VPC's name comes from the resource's `metadata.name`; the spec carries the region, an optional description, and an optional IP range — omit the range and DigitalOcean assigns a non-conflicting one.

## What this component models

The spec maps onto DigitalOcean's `digitalocean_vpc` in full:

| Spec field | What it controls |
|---|---|
| `region` | Where the VPC lives; create-only |
| `description` | Free-form description, up to 255 characters; editable |
| `ipRangeCidr` | The private range in CIDR notation, prefix /16 through /24; omit to let DigitalOcean choose. Immutable — changing it replaces the VPC |

Whether a VPC is the region's DEFAULT is computed by DigitalOcean and cannot be set through infrastructure-as-code -- and in a region with no VPC yet, the first one created becomes the undeletable default (see the GUIDE before declaring your first VPC in a new region).

## Quick start

The smallest real VPC:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpc
metadata:
  name: my-vpc
spec:
  region: nyc3
```

A production VPC with deliberate IP planning:

```yaml
spec:
  description: Production network for the web tier
  region: nyc3
  ipRangeCidr: "10.10.0.0/16"
```

Other kinds join the network by referencing this VPC:

```yaml
# on a DigitalOceanDroplet / KubernetesCluster / LoadBalancer / DatabaseCluster spec
vpc:
  valueFrom:
    kind: DigitalOceanVpc
    name: my-vpc
    fieldPath: status.outputs.vpc_id
```

## Behavior worth knowing

- **The range is immutable** — an edit to a set `ipRangeCidr` surfaces as a full VPC replacement at plan time; members must be moved first.
- **DigitalOcean accepts /16 through /24** — and the range must not overlap any other network in the account (including DigitalOcean's auto-assigned 10.x blocks).
- **Omitted range, honest read-back** — DigitalOcean's assigned range lands in the `ip_range` output without producing a plan diff.
- **Deletion waits for members** — a VPC with resources still inside refuses deletion; the module retries briefly, but the real fix is destroying the members first.
- **VPCs import by UUID** — the `vpc_id` output is the resource identity.

## Outputs

| Output | Meaning |
|---|---|
| `vpc_id` | The VPC's UUID — what other kinds' `vpc` references consume (`status.outputs.vpc_id`) |
| `ip_range` | The VPC's CIDR range as DigitalOcean reports it (covers the auto-assigned case) |
| `urn` | The uniform resource name (`do:vpc:<uuid>`) |

## See also

- `GUIDE.md` — operational judgment calls (range planning, the default-VPC trap, deletion order)
- `presets/` — standard starting point
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
