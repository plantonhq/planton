# GcpVpcNetwork

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVpcNetworkSpec defines configuration for a Google Cloud VPC (Virtual Private Cloud).

Private services access (Cloud SQL / AlloyDB / Memorystore private IP) is NOT
bundled here — compose GcpGlobalAddress (VPC_PEERING range) +
GcpServiceNetworkingConnection instead.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcNetwork
metadata:
  name: my-sample-vpc
spec:
  projectId:
    value: my-gcp-project-123
  networkName: app-vpc
  autoCreateSubnetworks: false
  routingMode: REGIONAL
  description: Application VPC with custom subnets

  # BGP best-path selection: STANDARD unlocks MED comparison across
  # neighbor ASNs and inter-region cost biasing — the multi-site levers.
  bgpBestPathSelection:
    mode: STANDARD
    alwaysCompareMed: true
    interRegionCost: ADD_COST_TO_MED

  # DELETE keeps destroys real; PREVENT belongs on the network everything
  # else depends on.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.autoCreateSubnetworks` | `bool` |  |  |  |
| `spec.routingMode` | `enum` |  | `REGIONAL` |  |
| `spec.networkName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.mtu` | `int32` |  |  |  |
| `spec.enableUlaInternalIpv6` | `bool` |  |  |  |
| `spec.internalIpv6Range` | `string` |  |  |  |
| `spec.networkFirewallPolicyEnforcementOrder` | `string` |  |  |  |
| `spec.networkProfile` | `string` |  |  |  |
| `spec.bgpBestPathSelection` | `GcpVpcNetworkBgpBestPathSelection` |  |  |  |
| `spec.bgpBestPathSelection.mode` | `string` |  |  |  |
| `spec.bgpBestPathSelection.alwaysCompareMed` | `bool` |  |  |  |
| `spec.bgpBestPathSelection.interRegionCost` | `string` |  |  |  |
| `spec.deleteDefaultRoutesOnCreate` | `bool` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns this VPC network.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the network.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.autoCreateSubnetworks

`bool`

Whether to use auto subnet mode (true) or custom subnet mode (false).
**Default:** false (custom mode). Auto mode is not recommended for production.

### spec.routingMode

`enum` · optional (explicit presence)

Dynamic routing mode for the VPC's Cloud Routers: REGIONAL or GLOBAL.
**Default:** REGIONAL (Cloud Router advertises routes only in one region).
Use GLOBAL only for multi-region routing needs.

- default: `REGIONAL`

Allowed values (use exactly as shown):

- `REGIONAL`
- `GLOBAL`

### spec.networkName

`string` · required

Name of the VPC network to create in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a lowercase letter or number.
Example: "my-vpc-network", "prod-network"

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.description

`string`

Human-readable description of the network. Immutable: changing it
destroys and recreates the network.

- rule: {"string":{"maxLen":"2048"}}

### spec.mtu

`int32` · optional (explicit presence)

Maximum Transmission Unit in bytes (1300–8896). Default 1460.
Jumbo frames (up to 8896) apply within the VPC; traffic to the internet
or other VPCs may still be subject to lower effective MTUs.

- rule: {"int32":{"lte":8896,"gte":1300}}

### spec.enableUlaInternalIpv6

`bool`

Enable ULA internal IPv6 on this network. When true, GCP assigns a /48
from the fd20::/20 ULA prefix (or uses internal_ipv6_range when set).

### spec.internalIpv6Range

`string`

When enabling ULA internal IPv6, optionally specify the /48 range from
fd20::/20. Immutable. If omitted, GCP allocates one automatically.

### spec.networkFirewallPolicyEnforcementOrder

`string`

Order in which firewall policies and classic firewall rules are evaluated.
Default AFTER_CLASSIC_FIREWALL.

- rule: network_firewall_policy_enforcement_order must be BEFORE_CLASSIC_FIREWALL or AFTER_CLASSIC_FIREWALL

### spec.networkProfile

`string`

Full or partial URL of a network profile to apply at creation time.
Immutable.

### spec.bgpBestPathSelection

`GcpVpcNetworkBgpBestPathSelection`

BGP best-path selection settings for the network.

### spec.bgpBestPathSelection.mode

`string`

The BGP best selection algorithm: LEGACY (default) or STANDARD.

- rule: mode must be LEGACY or STANDARD

### spec.bgpBestPathSelection.alwaysCompareMed

`bool`

When mode is STANDARD, enables comparison of MED across routes with
different neighbor ASNs.

### spec.bgpBestPathSelection.interRegionCost

`string`

When mode is STANDARD, controls inter-regional cost behavior in the BPS
algorithm: DEFAULT or ADD_COST_TO_MED.

- rule: inter_region_cost must be DEFAULT or ADD_COST_TO_MED

### spec.deleteDefaultRoutesOnCreate

`bool`

When true, default routes (0.0.0.0/0) are not created automatically.
Immutable.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the network for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only: changing them later replaces the network.

### spec.deletionPolicy

`string`

Deletion policy for the VPC network — what happens when this resource
is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the network is deleted (GCP refuses while subnets,
               peerings, or attached resources remain in it)
  "PREVENT" -- destroy FAILS; protects the network every subnet,
               route, and peering in it depends on
  "ABANDON" -- the network is removed from management but left
               serving in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVpcNetwork, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.network_self_link` | `string` | Full self-link URL of the created network (useful for connecting subnets or other resources to this VPC). |
| `status.outputs.network_name` | `string` | Name of the VPC network (e.g., "my-vpc-network"). Referenced by GcpCloudRun.spec.vpc_access.network FK. |
| `status.outputs.network_id` | `string` | GCP self-link of the VPC network (e.g., "projects/PROJECT/global/networks/NAME"). Used by GcpCloudSql.spec.network.private_network for private-IP configuration. |
| `status.outputs.gateway_ipv4` | `string` | IPv4 address of the default internet gateway for this network (if present). |
| `status.outputs.internal_ipv6_range` | `string` | ULA internal IPv6 range assigned to this network when ULA IPv6 is enabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAddress | `spec.network` | `status.outputs.network_self_link` |
| GcpAlloydbCluster | `spec.network` | `status.outputs.network_id` |
| GcpCloudComposerEnvironment | `spec.nodeConfig.network` | `status.outputs.network_self_link` |
| GcpCloudFunction | `spec.serviceConfig.directVpcNetworkInterface.network` | `status.outputs.network_name` |
| GcpCloudRun | `spec.vpcAccess.networkInterfaces[].network` | `status.outputs.network_name` |
| GcpCloudRunJob | `spec.template.vpcAccess.networkInterfaces[].network` | `status.outputs.network_name` |
| GcpCloudSql | `spec.network.privateNetwork` | `status.outputs.network_id` |
| GcpComputeInstance | `spec.networkInterfaces[].network` | `status.outputs.network_self_link` |
| GcpComputeMig | `spec.template.networkInterfaces[].network` | `status.outputs.network_self_link` |
| GcpDataprocCluster | `spec.clusterConfig.gceConfig.network` | `status.outputs.network_self_link` |
| GcpDnsRecord | `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `status.outputs.network_self_link` |
| GcpDnsRecord | `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `status.outputs.network_self_link` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].networkUrl` | `status.outputs.network_self_link` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].networkUrl` | `status.outputs.network_self_link` |
| GcpDnsZone | `spec.privateVisibilityConfig.networks[].networkUrl` | `status.outputs.network_self_link` |
| GcpDnsZone | `spec.peeringConfig.targetNetwork` | `status.outputs.network_self_link` |
| GcpFilestoreInstance | `spec.fileShare.nfsExportOptions[].network` | `status.outputs.network_name` |
| GcpFilestoreInstance | `spec.networkConfig.network` | `status.outputs.network_name` |
| GcpFirewallRule | `spec.network` | `status.outputs.network_self_link` |
| GcpGcsBucket | `spec.ipFilter.vpcNetworkSources[].network` | `status.outputs.network_id` |
| GcpGkeCluster | `spec.network` | `status.outputs.network_self_link` |
| GcpGkeNodePool | `spec.networkConfig.additionalNodeNetworks[].network` | `status.outputs.network_self_link` |
| GcpGlobalAddress | `spec.network` | `status.outputs.network_self_link` |
| GcpGlobalForwardingRule | `spec.network` | `status.outputs.network_self_link` |
| GcpHaVpnGateway | `spec.network` | `status.outputs.network_self_link` |
| GcpMemorystoreInstance | `spec.pscAutoConnections[].network` | `status.outputs.network_id` |
| GcpPlantonRunner | `spec.vpcAccess.network` | `status.outputs.network_name` |
| GcpRedisInstance | `spec.authorizedNetwork` | `status.outputs.network_self_link` |
| GcpRegionNetworkEndpointGroup | `spec.network` | `status.outputs.network_self_link` |
| GcpRouterNat | `spec.vpcSelfLink` | `status.outputs.network_self_link` |
| GcpServerlessVpcConnector | `spec.network` | `status.outputs.network_name` |
| GcpServiceConnectionPolicy | `spec.network` | `status.outputs.network_id` |
| GcpServiceNetworkingConnection | `spec.network` | `status.outputs.network_self_link` |
| GcpSubnetwork | `spec.vpcSelfLink` | `status.outputs.network_self_link` |
| GcpVertexAiEndpoint | `spec.network` | `status.outputs.network_self_link` |
| GcpVertexAiEndpoint | `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | `status.outputs.network_self_link` |
| GcpVertexAiIndexEndpoint | `spec.network` | `status.outputs.network_self_link` |
| GcpVertexAiIndexEndpoint | `spec.privateServiceConnectConfig.pscAutomationConfigs[].network` | `status.outputs.network_self_link` |
| GcpVertexAiNotebook | `spec.networkInterface.network` | `status.outputs.network_self_link` |
| GcpVpcPeering | `spec.network` | `status.outputs.network_self_link` |
| GcpVpcPeering | `spec.peerNetwork` | `status.outputs.network_self_link` |

## See Also

- [Overview](../README.md)
