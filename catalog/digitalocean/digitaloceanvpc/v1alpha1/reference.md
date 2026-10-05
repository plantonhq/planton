# DigitalOceanVpc

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanVpcSpec defines the specification required to deploy a DigitalOcean Virtual
Private Cloud (VPC) -- a private, isolated network for Droplets and other resources within one
region. The VPC's name comes from the resource's metadata.name; the spec models the provider's
remaining argument surface in full.

## Example

```yaml
# Example DigitalOceanVpc manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- the smallest real VPC: region only. DigitalOcean assigns a
# non-conflicting IP range and reports it through the ip_range output.
#
# Document 2 -- a production-shaped VPC: described, with an explicit /16 for
# deliberate IP planning. The range is immutable -- changing it later
# replaces the VPC.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpc
metadata:
  name: example-dovpc-minimal
spec:
  region: nyc3
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpc
metadata:
  name: example-dovpc-full
spec:
  description: Production network for the web tier
  region: nyc3
  ipRangeCidr: "10.10.0.0/16"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.description` | `string` |  |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.ipRangeCidr` | `string` |  |  |  |

## Field Details

### spec.description

`string`

(Optional) A human-readable description for the VPC (up to 255 characters).

- rule: {"string":{"maxLen":"255"}}

### spec.region

`enum` · required

The DigitalOcean region where the VPC will be created. Cannot be changed after creation.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.ipRangeCidr

`string`

(Optional) The IP range for the VPC in CIDR notation. DigitalOcean accepts prefix lengths
from /16 through /24, and the range must not overlap any other network in the account.
Example: "10.10.0.0/16"

When omitted, DigitalOcean auto-generates a non-conflicting range and reports it back
through the `ip_range` output. The range is immutable: changing it after creation
REPLACES the VPC.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/(1[6-9]|2[0-4])$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanVpc, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpc_id` | `string` | The unique identifier (UUID) of the created DigitalOcean VPC. Other kinds' `vpc` references resolve against this output. |
| `status.outputs.ip_range` | `string` | The VPC's IP range in CIDR notation. Reported by DigitalOcean, which also covers the case where ip_range_cidr was left unset and DigitalOcean auto-assigned a range. |
| `status.outputs.urn` | `string` | The uniform resource name (URN) of the VPC, e.g. "do:vpc:<uuid>". |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanApp | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanDatabaseCluster | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanDatabaseReplica | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanDroplet | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanDropletAutoscalePool | `spec.dropletTemplate.vpc` | `status.outputs.vpc_id` |
| DigitalOceanKubernetesCluster | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanLoadBalancer | `spec.vpc` | `status.outputs.vpc_id` |
| DigitalOceanProject | `spec.resources` | `status.outputs.urn` |
| DigitalOceanVpcPeering | `spec.vpc1` | `status.outputs.vpc_id` |
| DigitalOceanVpcPeering | `spec.vpc2` | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)
