# DigitalOceanReservedIp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanReservedIpSpec models a DigitalOcean reserved IP address --
one product concept covering the provider's four resources
(digitalocean_reserved_ip / digitalocean_reserved_ipv6 and their
assignment resources): a static public IP reserved in a region, IPv4 or
IPv6, optionally assigned to a droplet. The v4/v6 API asymmetries
(argument names, inline-assign vs assignment-resource, read-back and
delete behavior) are absorbed by the IaC modules, never surfaced here.

BILLING: an UNASSIGNED reserved IPv4 accrues a monthly charge
(prorated hourly) precisely because it holds capacity without a droplet;
an assigned one is free. Reserved IPv6 addresses are free either way.
Destroy the resource when the address is no longer needed -- an orphaned
reservation bills forever.

## Example

```yaml
# Reference manifests for DigitalOceanReservedIp -- protovalidate-valid,
# embedded as the reference page's Example block, and the documents the
# offline tofu plans render. Three documents covering the module's three
# shapes: unassigned IPv4, droplet-assigned IPv4 (inline, mutable), and
# droplet-assigned IPv6 (via the separate assignment resource).
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanReservedIp
metadata:
  name: ingress-standby-ip
spec:
  region: nyc3
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanReservedIp
metadata:
  name: web-frontdoor-ip
spec:
  ipVersion: ipv4
  region: nyc3
  # Literal numeric droplet id; use valueFrom to reference a
  # DigitalOceanDroplet resource instead.
  droplet:
    value: "123456789"
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanReservedIp
metadata:
  name: web-frontdoor-ipv6
spec:
  ipVersion: ipv6
  region: nyc3
  droplet:
    value: "123456789"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.ipVersion` | `string` |  | `ipv4` |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.droplet` | `string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |

## Field Details

### spec.ipVersion

`string`

(Optional) IP version to reserve: ipv4 or ipv6. When unset, ipv4 is
reserved. Cannot be changed after creation -- switching versions
replaces the reservation with a new address.

- default: `ipv4`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","ipv6"]}}

### spec.region

`enum` · required

The DigitalOcean region to reserve the address in. A reserved IP can
only be assigned to droplets in the same region. Changing it replaces
the reservation with a new address.

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

### spec.droplet

`string | valueFrom`

(Optional) The droplet to assign the reserved IP to. Use a literal
numeric droplet ID or a reference to a DigitalOceanDroplet resource.
Assigning, re-pointing, and unassigning all apply in place -- the
reserved address itself never changes. When unset, the address stays
reserved but unassigned (which is exactly the state that BILLS for
IPv4 -- see the message comment).

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanReservedIp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.reserved_ip_address` | `string` | The reserved IP address (IPv4 or IPv6 per the spec's ip_version). This IS the resource identity: imports, lookups, and assignments all address the reservation by this value. |
| `status.outputs.urn` | `string` | The uniform resource name of the reservation (e.g. "do:reservedip:203.0.113.10" or "do:reservedipv6:2001:db8::1"), as used by DigitalOcean project membership. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.droplet` | DigitalOceanDroplet | `status.outputs.droplet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanProject | `spec.resources` | `status.outputs.urn` |

## See Also

- [Overview](../README.md)
