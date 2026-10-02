# GcpVpcPeering

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVpcPeeringSpec manages ONE SIDE of a VPC Network Peering: the peering
entry on `network` that points at `peer_network`, and the route exchange
that side allows. VPC peering connects two VPC networks -- in the same
project, different projects, or different organizations -- so their
workloads reach each other over internal IPs with no gateway, no
bandwidth cap beyond the VMs' own, and no egress through the internet.

A peering is two half-entries, one in each network, and goes ACTIVE only
when both exist and point at each other. Declare each side as its own
GcpVpcPeering: two resources when you own both networks (the `state`
output on each side tells you when the pair is up), one when the other
side belongs to someone else (it stays INACTIVE until they create theirs).
The provider serializes the two sides when one chart creates both, so
declaring both at once is safe.

Two forms, selected by whether peer_network is set:

  CREATE (peer_network set) -- this resource creates the peering entry
  and owns its route exchange. Nearly everything is immutable: the name,
  both networks, and the two public-IP subnet-route flags recreate the
  peering when changed. The custom-route flags, stack_type, and
  update_strategy change in place.

  ROUTES-CONFIG (peer_network empty) -- the peering ALREADY EXISTS on
  `network` under `peering_name` and this resource manages only its route
  exchange. This is the shape for peerings Google creates on your behalf:
  `servicenetworking-googleapis-com` for Cloud SQL private IP, Memorystore,
  and other private-services-access products. Exporting custom routes to
  that peering is how an on-premises network reached over VPN or
  Interconnect learns the path to a Cloud SQL instance. Destroying this
  form is a no-op in GCP: the peering and its current route flags stay as
  they are.

Routing rules Google enforces, taught here because the API reports them
only at apply time: the two networks' subnet ranges must not overlap;
peering is NOT transitive (A-B and B-C do not connect A to C); a network
can have at most 25 active peerings (raiseable by quota); and dynamic
routes learned over VPN/Interconnect cross a peering only when the
exporting side sets export_custom_routes AND the importing side sets
import_custom_routes.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcPeering
metadata:
  name: planton-oss-e2e-gcppeer
  id: planton-oss-e2e-gcppeer
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcpvpcpeering
  annotations:
    planton.dev/e2e: "true"
    # Side A under test needs a second network and the side-B peering on it
    # so the pair goes ACTIVE; both are extra fixture instances (path entries
    # deploy after the registry chain, in this order).
    planton.dev/e2e-prerequisites: "catalog/gcp/gcpvpcpeering/e2e/prerequisites/gcpvpcnetwork-peer.yaml,catalog/gcp/gcpvpcpeering/e2e/prerequisites/gcpvpcpeering-side-b.yaml"
  tags:
    - planton-e2e
spec:
  # The peering entry's name on this side; defaults to metadata.name.
  peeringName: planton-oss-e2e-gcppeer-side-a

  # This side's network: the registry's prerequisite GcpVpcNetwork.
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: planton-oss-e2e-gcpvpcnetwork-prereq
      fieldPath: status.outputs.network_self_link

  # The other network: the second fixture network. Set, so this side CREATES
  # the peering.
  peerNetwork:
    valueFrom:
      kind: GcpVpcNetwork
      name: planton-oss-e2e-gcppeer-peer-vpc
      fieldPath: status.outputs.network_self_link

  # Offer this side's custom routes to the peer (side B imports them).
  exportCustomRoutes: true
  importCustomRoutes: false

  # DELETE (default), PREVENT, ABANDON -- create form only.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.peeringName` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.peerNetwork` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.exportCustomRoutes` | `bool` |  |  |  |
| `spec.importCustomRoutes` | `bool` |  |  |  |
| `spec.exportSubnetRoutesWithPublicIp` | `bool` |  | `true` |  |
| `spec.importSubnetRoutesWithPublicIp` | `bool` |  | `false` |  |
| `spec.stackType` | `string` |  |  |  |
| `spec.updateStrategy` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.peeringName

`string`

The name of the peering entry on `network`: what `gcloud compute
networks peerings list` shows, and the name the other side does NOT
need to match (each side names its own entry). Defaults to
metadata.name when empty. 1-63 characters, lowercase letters, digits,
and hyphens, starting with a letter. Immutable in the CREATE form.

In the ROUTES-CONFIG form this is the name of the EXISTING peering
whose routes are managed -- for Google-managed private services access
it is `servicenetworking-googleapis-com`.

- rule: peering_name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.network

`string | valueFrom` · required

The VPC network this side of the peering lives in: a reference to a
GcpVpcNetwork (its network_self_link output) or the network's self link
as a literal. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.peerNetwork

`string | valueFrom`

The OTHER network -- the one this side peers with: a reference to a
GcpVpcNetwork or its self link as a literal
(`projects/{project}/global/networks/{name}` or the full
`https://www.googleapis.com/compute/v1/...` form). May live in another
project or another organization. Immutable.

Set it and this resource CREATES the peering. Leave it EMPTY and this
resource manages the route exchange of a peering that already exists
on `network` under `peering_name` (the routes-config form above).

The peer network is where the traffic goes, not where this resource
lives, so it is an access edge on diagrams, not a containment edge.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.exportCustomRoutes

`bool`

Export this network's CUSTOM routes (static routes and dynamic routes
learned over Cloud VPN / Interconnect / Router Appliance) to the peer.
Default false: only subnet routes cross a peering unless both sides opt
in -- this side exports, the other side imports. The one flag that
makes an on-premises network reachable from the peer (or, on a
service-networking peering, makes Cloud SQL reachable from
on-premises). Mutable. Always sent; in the routes-config form the
provider requires it.

### spec.importCustomRoutes

`bool`

Import the peer network's custom routes into this network. Default
false. Takes effect only when the peer's side exports them. Mutable.
Always sent; in the routes-config form the provider requires it.

### spec.exportSubnetRoutesWithPublicIp

`bool` · optional (explicit presence)

Export subnet routes whose ranges are PUBLIC IPs (privately used public
IP ranges) to the peer. Default TRUE -- Google exports them unless told
not to. Always sent on both forms (the default when unset, so the
manifest states the exchange in full). Immutable in the CREATE form
(changing it recreates the peering); changes in place in the
routes-config form.

- default: `true`

### spec.importSubnetRoutesWithPublicIp

`bool` · optional (explicit presence)

Import the peer's public-IP subnet routes into this network. Default
FALSE -- a network does not learn a peer's privately-used public ranges
unless it asks to. Always sent on both forms (the default when unset).
Immutable in the CREATE form; changes in place in the routes-config
form.

- default: `false`

### spec.stackType

`string`

Which IP versions the peering carries:
  ""          -- same as IPV4_ONLY (provider default)
  "IPV4_ONLY" -- only IPv4 subnet ranges are exchanged
  "IPV4_IPV6" -- IPv4 and IPv6 ranges; both networks must be dual-stack
                 (an internal IPv6 range on each GcpVpcNetwork)
CREATE form only. Mutable.

- rule: stack_type must be IPV4_ONLY or IPV4_IPV6

### spec.updateStrategy

`string`

How a change to this peering's settings is applied:
  ""            -- same as INDEPENDENT (provider default)
  "INDEPENDENT" -- this side's route-exchange flags take effect on
                   their own, as soon as applied
  "CONSENSUS"   -- a changed setting takes effect only once BOTH sides
                   agree on it; until then the peering keeps the old
                   behavior. Use it when the two sides are owned by
                   different teams and a one-sided flip must never
                   change what traffic flows.
CREATE form only. Mutable.

- rule: update_strategy must be INDEPENDENT or CONSENSUS

### spec.deletionPolicy

`string`

What destroying this resource does to the peering in GCP (CREATE form
only -- destroying the routes-config form never touches the peering):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- this side's peering entry is removed; the pair goes
               INACTIVE and traffic stops. The other side's entry stays
               until it is removed too.
  "PREVENT" -- destroy FAILS; the guard for the peering production
               traffic depends on
  "ABANDON" -- the resource leaves management but the peering entry
               stays live

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `stack_type_needs_create_form`: stack_type applies only when this resource creates the peering (peer_network set); the routes-config form leaves it empty
- `update_strategy_needs_create_form`: update_strategy applies only when this resource creates the peering (peer_network set); the routes-config form leaves it empty
- `deletion_policy_needs_create_form`: deletion_policy applies only when this resource creates the peering (peer_network set); destroying the routes-config form is a no-op in GCP

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVpcPeering, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.peering_name` | `string` | The peering entry's name on this side's network -- the resolved value (metadata.name when the spec left peering_name empty). |
| `status.outputs.network` | `string` | This side's network, as a self link. |
| `status.outputs.state` | `string` | The peering's state as GCP reports it on this side: `ACTIVE` when both sides exist and point at each other, `INACTIVE` while the other side is missing or mismatched. Populated only in the CREATE form; empty for the routes-config form, which does not own the peering entry. |
| `status.outputs.state_details` | `string` | GCP's explanation of the state -- why a peering is INACTIVE (`Peering is not active because the peer network does not have a matching peering`), or the timestamp it went active. Populated only in the CREATE form. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.peerNetwork` | GcpVpcNetwork | `status.outputs.network_self_link` |

## See Also

- [Overview](../README.md)
