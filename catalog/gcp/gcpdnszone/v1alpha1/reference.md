# GcpDnsZone

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpDnsZoneSpec defines a Google Cloud DNS managed zone
(`google_dns_managed_zone`). DNS records belong in the separate GcpDnsRecord
kind — this resource owns the zone shell only.

## Example

```yaml
# Public managed zone for an internet-facing domain.
# DNS records are composed separately via GcpDnsRecord.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsZone
metadata:
  name: example.com
spec:
  projectId:
    value: my-gcp-project-123
  visibility: public
  description: Public authoritative zone for example.com

  # Ephemeral test fixture: delete the zone shell on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.dnsName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.visibility` | `string` |  | `public` |  |
| `spec.privateVisibilityConfig` | `GcpDnsZonePrivateVisibilityConfig` |  |  |  |
| `spec.privateVisibilityConfig.networks` | `[]GcpDnsZonePrivateVisibilityNetwork` |  |  |  |
| `spec.privateVisibilityConfig.networks[].networkUrl` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.privateVisibilityConfig.gkeClusters` | `[]GcpDnsZonePrivateVisibilityGkeCluster` |  |  |  |
| `spec.privateVisibilityConfig.gkeClusters[].gkeClusterName` | `string \| valueFrom` | yes |  | GcpGkeCluster (`status.outputs.cluster_id`) |
| `spec.dnssecConfig` | `GcpDnsZoneDnssecConfig` |  |  |  |
| `spec.dnssecConfig.state` | `string` |  | `off` |  |
| `spec.dnssecConfig.defaultKeySpecs` | `[]GcpDnsZoneDnssecKeySpec` |  |  |  |
| `spec.dnssecConfig.defaultKeySpecs[].algorithm` | `string` |  |  |  |
| `spec.dnssecConfig.defaultKeySpecs[].keyLength` | `int32` |  |  |  |
| `spec.dnssecConfig.defaultKeySpecs[].keyType` | `string` |  |  |  |
| `spec.dnssecConfig.nonExistence` | `string` |  |  |  |
| `spec.forwardingConfig` | `GcpDnsZoneForwardingConfig` |  |  |  |
| `spec.forwardingConfig.targetNameServers` | `[]GcpDnsZoneForwardingTargetNameServer` | yes |  |  |
| `spec.forwardingConfig.targetNameServers[].ipv4Address` | `string` |  |  |  |
| `spec.forwardingConfig.targetNameServers[].domainName` | `string` |  |  |  |
| `spec.forwardingConfig.targetNameServers[].forwardingPath` | `string` |  |  |  |
| `spec.forwardingConfig.targetNameServers[].ipv6Address` | `string` |  |  |  |
| `spec.peeringConfig` | `GcpDnsZonePeeringConfig` |  |  |  |
| `spec.peeringConfig.targetNetwork` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.cloudLoggingConfig` | `GcpDnsZoneCloudLoggingConfig` |  |  |  |
| `spec.cloudLoggingConfig.enableLogging` | `bool` |  |  |  |
| `spec.forceDestroy` | `bool` |  | `false` |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the managed zone. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.dnsName

`string`

Authoritative DNS domain for the zone (e.g. "example.com."). Must end with
a trailing dot. When omitted, modules derive dns_name from metadata.name
plus "." — preserving the legacy default for public zones.

- rule: dns_name must be a fully qualified domain name ending with a trailing dot (e.g., example.com.)

### spec.description

`string`

Human-readable description shown in the GCP console.

### spec.visibility

`string` · optional (explicit presence)

Zone visibility: public (internet-facing) or private (VPC/GKE only).

- default: `public`
- rule: visibility must be public or private

### spec.privateVisibilityConfig

`GcpDnsZonePrivateVisibilityConfig`

VPC/GKE visibility targets for standard private zones. Not used with
forwarding or peering zone types.

- rule: private_visibility_config requires at least one network or gke_cluster

### spec.privateVisibilityConfig.networks

`[]GcpDnsZonePrivateVisibilityNetwork`

### spec.privateVisibilityConfig.networks[].networkUrl

`string | valueFrom` · required

VPC network self-link or resource name. Reference a GcpVpcNetwork or
supply the full URL (projects/{project}/global/networks/{network}).

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.privateVisibilityConfig.gkeClusters

`[]GcpDnsZonePrivateVisibilityGkeCluster`

### spec.privateVisibilityConfig.gkeClusters[].gkeClusterName

`string | valueFrom` · required

GKE cluster resource path (projects/{p}/locations/{loc}/clusters/{name}).
Reference a GcpGkeCluster or supply the path directly.

- references: GcpGkeCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.dnssecConfig

`GcpDnsZoneDnssecConfig`

DNSSEC signing configuration (public zones).

### spec.dnssecConfig.state

`string` · optional (explicit presence)

DNSSEC mode: off, on, or transfer (import existing signed zone).

- default: `off`
- rule: state must be off, on, or transfer

### spec.dnssecConfig.defaultKeySpecs

`[]GcpDnsZoneDnssecKeySpec`

Optional custom initial key specs. Updatable only while state is off.

### spec.dnssecConfig.defaultKeySpecs[].algorithm

`string`

DNSSEC algorithm. Common values: ecdsap256sha256, rsasha256.

- rule: algorithm must be ecdsap256sha256, ecdsap384sha384, rsasha1, rsasha256, or rsasha512

### spec.dnssecConfig.defaultKeySpecs[].keyLength

`int32`

Key length in bits (e.g. 256 for ECDSA, 2048 for RSA).

### spec.dnssecConfig.defaultKeySpecs[].keyType

`string`

keySigning (KSK) or zoneSigning (ZSK).

- rule: key_type must be keySigning or zoneSigning

### spec.dnssecConfig.nonExistence

`string`

Authenticated denial-of-existence mechanism: nsec or nsec3.

- rule: non_existence must be nsec or nsec3

### spec.forwardingConfig

`GcpDnsZoneForwardingConfig`

Outbound forwarding targets (private forwarding zones).

### spec.forwardingConfig.targetNameServers

`[]GcpDnsZoneForwardingTargetNameServer` · required

- rule: {"repeated":{"minItems":"1"}}
- rule: each target_name_server requires ipv4_address, ipv6_address, or domain_name
- rule: a target_name_server accepts ipv4_address or ipv6_address, never both

### spec.forwardingConfig.targetNameServers[].ipv4Address

`string`

IPv4 address of the upstream resolver.

### spec.forwardingConfig.targetNameServers[].domainName

`string`

Fully qualified domain name of the forwarding target.

### spec.forwardingConfig.targetNameServers[].forwardingPath

`string`

Query path: default (RFC1918 via VPC, else internet) or private (always VPC).

- rule: forwarding_path must be default or private

### spec.forwardingConfig.targetNameServers[].ipv6Address

`string`

IPv6 address of the upstream resolver. A target carries one address
family — the provider rejects a target with both IPv4 and IPv6 set.

### spec.peeringConfig

`GcpDnsZonePeeringConfig`

DNS peering target network (private peering zones).

### spec.peeringConfig.targetNetwork

`string | valueFrom` · required

VPC network to peer with. Reference a GcpVpcNetwork or supply the self-link.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.cloudLoggingConfig

`GcpDnsZoneCloudLoggingConfig`

Cloud DNS query logging.

### spec.cloudLoggingConfig.enableLogging

`bool`

When true, log every query received by this managed zone. Off by default.

### spec.forceDestroy

`bool` · optional (explicit presence)

When true, delete all record sets in the zone on destroy. Default false.

- default: `false`

### spec.labels

`map<string, string>`

Additional GCP labels merged with platform labels.

### spec.deletionPolicy

`string`

Deletion policy for the managed zone — the second of this kind's two
destroy levers (force_destroy empties the records first; this decides
the zone shell itself):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the zone is deleted (GCP refuses while non-default
               record sets remain unless force_destroy is true); its
               delegated name servers stop answering
  "PREVENT" -- destroy FAILS; protects a zone that registrars and
               parent zones delegate to
  "ABANDON" -- the zone is removed from management but keeps serving
               DNS in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `forwarding_requires_private`: forwarding_config requires visibility private
- `peering_requires_private`: peering_config requires visibility private
- `private_visibility_requires_private`: private_visibility_config requires visibility private
- `forwarding_peering_mutually_exclusive`: forwarding_config and peering_config are mutually exclusive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDnsZone, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | Numeric identifier the Cloud DNS API assigns to the managed zone. |
| `status.outputs.zone_name` | `string` | Name of the managed zone resource in GCP — the handle GcpDnsRecord references when creating record sets in this zone. |
| `status.outputs.nameservers` | `[]string` | Nameservers Cloud DNS assigned to the zone. For public zones, configure these at the domain registrar to delegate the domain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.privateVisibilityConfig.networks[].networkUrl` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.privateVisibilityConfig.gkeClusters[].gkeClusterName` | GcpGkeCluster | `status.outputs.cluster_id` |
| `spec.peeringConfig.targetNetwork` | GcpVpcNetwork | `status.outputs.network_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDnsRecord | `spec.managedZone` | `status.outputs.zone_name` |
| KubernetesExternalDns | `spec.googleCloudDns.zoneIdFilters` | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
