# AwsEksCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEksClusterSpec defines an EKS cluster CONTROL PLANE: the managed
Kubernetes API server, etcd, and the cluster-level posture that everything
else composes onto -- networking exposure, authentication mode, secrets
encryption, control-plane logging, upgrade policy, and (optionally) EKS
Auto Mode.

The cluster is deliberately only the control plane. Compute attaches as
separate, composable nodes: AwsEksNodeGroup references this cluster's
status.outputs.name, and IAM trust for workloads flows through the
cluster's status.outputs.oidc_issuer_url (point an AwsIamOidcProvider at
it to enable IRSA with a single reference).

The cluster name comes from metadata.name (AWS limit: 100 characters,
alphanumeric start, then alphanumerics/underscores/hyphens). The name and
the cluster IAM role are create-only in AWS; the immutable networking
fields are called out on each field below.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksCluster
metadata:
  name: awsekscluster-demo
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  clusterRoleArn:
    value: arn:aws:iam::123456789012:role/EksClusterServiceRole
  version: "1.31"
  endpointPrivateAccess: true
  accessConfig:
    authenticationMode: API
  enabledClusterLogTypes:
    - audit
    - authenticator
  upgradeSupportType: STANDARD
  # Provisioned control-plane capacity for large/bursty clusters (billed
  # hourly on top of the cluster fee; "standard" is the free default).
  controlPlaneScalingTier: tier-xl
  # EKS Hybrid Nodes: the on-premises node/pod ranges allowed to join over
  # VPN/Direct Connect. Declaring ranges is free.
  remoteNetworks:
    nodeCidrs:
      - 10.80.0.0/16
    podCidrs:
      - 10.85.0.0/16
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.clusterRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.version` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.endpointPublicAccess` | `bool` |  |  |  |
| `spec.endpointPrivateAccess` | `bool` |  |  |  |
| `spec.publicAccessCidrs` | `[]string` |  |  |  |
| `spec.controlPlaneEgressMode` | `string` |  |  |  |
| `spec.ipFamily` | `string` |  |  |  |
| `spec.serviceIpv4Cidr` | `string` |  |  |  |
| `spec.enabledClusterLogTypes` | `[]string` |  |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.accessConfig` | `AwsEksClusterAccessConfig` |  |  |  |
| `spec.accessConfig.authenticationMode` | `string` |  |  |  |
| `spec.accessConfig.bootstrapClusterCreatorAdminPermissions` | `bool` |  |  |  |
| `spec.autoMode` | `AwsEksClusterAutoMode` |  |  |  |
| `spec.autoMode.enabled` | `bool` |  |  |  |
| `spec.autoMode.nodePools` | `[]string` |  |  |  |
| `spec.autoMode.nodeRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.upgradeSupportType` | `string` |  |  |  |
| `spec.zonalShiftEnabled` | `bool` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.bootstrapSelfManagedAddons` | `bool` |  |  |  |
| `spec.forceUpdateVersion` | `bool` |  |  |  |
| `spec.controlPlaneScalingTier` | `string` |  |  |  |
| `spec.remoteNetworks` | `AwsEksClusterRemoteNetworks` |  |  |  |
| `spec.remoteNetworks.nodeCidrs` | `[]string` |  |  |  |
| `spec.remoteNetworks.podCidrs` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the cluster control plane runs in. Must match the region
of the subnets it attaches to. Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom` · required

The subnets (at least two, in distinct availability zones) where EKS
places the control plane's elastic network interfaces. These decide
which zones the API server is reachable from inside the VPC; worker
subnets are chosen separately on each node group. Reference AwsSubnet
subnet_id outputs or pass literal subnet IDs. Create-only in AWS
(changing the set updates in place, but the VPC itself cannot change).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.clusterRoleArn

`string | valueFrom` · required

The IAM role the EKS control plane assumes to manage AWS resources on
your behalf (ENIs, security groups, logs). The role must trust
eks.amazonaws.com and carry the AmazonEKSClusterPolicy managed policy --
attach it on the AwsIamRole itself (managed_policy_arns); this component
never modifies a role it merely references. Reference an AwsIamRole's
role_arn output or pass a literal ARN. Create-only in AWS.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.version

`string`

The Kubernetes minor version of the control plane, e.g. "1.31". Leave
empty to let AWS pick the current default version. EKS only ever
upgrades one minor at a time and never downgrades -- lowering this value
is rejected by AWS. Node groups pin their own version, so upgrade the
control plane first, then roll the node groups.

### spec.securityGroupIds

`[]string | valueFrom`

Additional security groups attached to the control plane's network
interfaces, on top of the cluster security group EKS always creates
(exported as cluster_security_group_id). Most clusters need none --
reach for this only for legacy rules that must ride along. Reference
AwsSecurityGroup security_group_id outputs or pass literal IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.endpointPublicAccess

`bool` · optional (explicit presence)

Whether the Kubernetes API server is reachable from the internet. AWS
defaults this to true; set false for a fully private cluster (pair with
endpoint_private_access = true or the API server becomes unreachable).
Scope a public endpoint down with public_access_cidrs.

### spec.endpointPrivateAccess

`bool`

Whether the Kubernetes API server is reachable from within the VPC
through private ENIs. AWS defaults this to false, which is almost never
what a production cluster wants: without it, in-VPC clients (nodes,
CI runners) reach the API server over the public endpoint.

### spec.publicAccessCidrs

`[]string`

IPv4 CIDR blocks allowed to reach the PUBLIC API endpoint. Empty means
AWS's default of 0.0.0.0/0 (open to the internet -- rely on IAM/RBAC
only). Restricting this to office/VPN egress ranges is the single
cheapest hardening step for a public-endpoint cluster.

- rule: {"repeated":{"items":{"string":{"pattern":"^(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(?:\\.(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}/(?:[0-9]|[12]\\d|3[0-2])$"}}}}

### spec.controlPlaneEgressMode

`string`

How control-plane-initiated traffic egresses the VPC:
- "AWS_MANAGED" (AWS default): EKS routes control-plane egress itself.
- "CUSTOMER_ROUTED": egress follows your VPC route tables (inspection/
  egress-firewall architectures).
- "CUSTOMER_ISOLATED": no control-plane egress through your VPC.
Reverting CUSTOMER_ROUTED back to AWS_MANAGED is not supported in place
-- AWS forces cluster replacement.

### spec.ipFamily

`string`

The IP address family for pod and service networking: "ipv4" (AWS
default) or "ipv6". Create-only: changing it replaces the cluster. IPv6
clusters assign pod addresses from the VPC's IPv6 CIDR and require IPv6-
enabled subnets.

### spec.serviceIpv4Cidr

`string`

The CIDR block Kubernetes assigns SERVICE addresses from (ipv4 clusters
only). Must be a /12 to /24 inside the private ranges (10/8,
172.16/12, 192.168/16, 100.64/10) and must not overlap the VPC or any
peered/connected network -- overlap breaks routing in ways that only
surface later. Create-only. Empty keeps the AWS default
(10.100.0.0/16 or 172.20.0.0/16, whichever avoids the VPC).

### spec.enabledClusterLogTypes

`[]string`

Control-plane log types streamed to CloudWatch Logs: "api", "audit",
"authenticator", "controllerManager", "scheduler". Empty disables
control-plane logging. "audit" and "authenticator" are the two most
valuable in practice (who did what, and who got in); enabling all five
on a busy cluster carries real CloudWatch ingestion cost.

- rule: {"repeated":{"unique":true}}

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key for ENVELOPE ENCRYPTION of Kubernetes secrets
-- secrets in etcd are encrypted with a data key that this KMS key
wraps. Unset uses AWS-owned encryption only. One-way door: once enabled
it cannot be disabled or re-keyed on a live cluster (AWS forces
replacement). Reference an AwsKmsKey's key_arn output or pass a literal
ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.accessConfig

`AwsEksClusterAccessConfig`

How identities are granted cluster access (access entries vs the legacy
aws-auth ConfigMap) and whether the creator gets admin. Unset keeps
AWS defaults (API_AND_CONFIG_MAP + creator-admin).

- rule: authentication_mode must be 'API', 'API_AND_CONFIG_MAP', or 'CONFIG_MAP' when set

### spec.accessConfig.authenticationMode

`string`

The source of truth for cluster access:
- "API": EKS access entries only -- the modern model; IAM principals
  are granted access as first-class EKS resources.
- "API_AND_CONFIG_MAP" (AWS default): access entries plus the legacy
  aws-auth ConfigMap, for migration.
- "CONFIG_MAP": legacy aws-auth ConfigMap only.
Moving toward "API" is one-way on a live cluster: AWS allows
CONFIG_MAP -> API_AND_CONFIG_MAP -> API but never back.

### spec.accessConfig.bootstrapClusterCreatorAdminPermissions

`bool` · optional (explicit presence)

Whether the identity creating the cluster is automatically granted
cluster-admin. AWS defaults this to true. Set false for clusters
whose admins are managed explicitly through access entries -- with no
other admin configured, false can lock everyone out of a fresh
cluster. Create-only.

### spec.autoMode

`AwsEksClusterAutoMode`

EKS Auto Mode: AWS provisions and manages compute, block storage, and
load balancing for the cluster itself -- no node groups to operate.
The alternative to (not a companion of) explicit AwsEksNodeGroup
compute; in practice a cluster uses one model or the other.

- rule: each node pool must be 'general-purpose' or 'system'
- rule: node_pools only applies when auto mode is enabled
- rule: node_role_arn is required when node_pools is non-empty

### spec.autoMode.enabled

`bool`

Turn Auto Mode on. AWS then provisions and scales EC2 capacity,
provisions EBS volumes, and manages load balancers for the cluster's
workloads without any node group.

### spec.autoMode.nodePools

`[]string`

The built-in node pools Auto Mode may launch capacity from:
"general-purpose" (workloads) and/or "system" (cluster-critical pods
on dedicated capacity). Empty enables Auto Mode with no built-in
pools -- capacity then comes only from custom NodePool resources
defined in-cluster.

- rule: {"repeated":{"unique":true}}

### spec.autoMode.nodeRoleArn

`string | valueFrom`

The IAM role Auto Mode nodes assume (the node identity for launched
capacity). Required by AWS when node_pools is non-empty. Reference an
AwsIamRole's role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.upgradeSupportType

`string`

The upgrade support tier once this cluster's Kubernetes version leaves
standard support: "STANDARD" (upgrade on schedule, no surcharge) or
"EXTENDED" (AWS default -- stay on the version up to ~26 months at a
significant hourly surcharge). Choose STANDARD when you upgrade
promptly and want the surcharge risk gone.

### spec.zonalShiftEnabled

`bool`

Allows Amazon Application Recovery Controller to shift in-cluster
east-west traffic away from an impaired availability zone.

### spec.deletionProtection

`bool`

Blocks cluster deletion at the EKS API until explicitly disabled --
the guard rail for shared/production control planes.

### spec.bootstrapSelfManagedAddons

`bool` · optional (explicit presence)

Whether EKS installs the default self-managed add-ons (vpc-cni,
kube-proxy, CoreDNS) at creation. AWS defaults this to true; set false
for a "bring your own add-ons" cluster that manages every add-on
explicitly (the GitOps-friendly posture). Create-only.

### spec.forceUpdateVersion

`bool`

Force the Kubernetes version update even if pods cannot be safely
drained onto the new version (pod disruption budgets that can never be
satisfied). Only consulted while `version` changes.

### spec.controlPlaneScalingTier

`string`

EKS Provisioned Control Plane: pre-provisions control-plane capacity
for very large or bursty clusters instead of relying on EKS's
reactive scaling (which can lag sudden API-server load, e.g. mass
node joins or operator storms). "standard" (AWS default -- reactive
scaling, no surcharge) or a provisioned tier of increasing capacity:
"tier-xl", "tier-2xl", "tier-4xl", "tier-8xl" -- each billed hourly
ON TOP of the cluster fee. Updates in place; empty keeps standard.

### spec.remoteNetworks

`AwsEksClusterRemoteNetworks`

EKS Hybrid Nodes: the on-premises/edge networks whose nodes and pods
join this cluster over your VPN or Direct Connect. Declaring the
ranges is free -- billing starts only when hybrid nodes register.
Updates in place on a live cluster.

- rule: remote_networks requires at least one of node_cidrs or pod_cidrs
- rule: each node CIDR must be within 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, or 100.64.0.0/10
- rule: each pod CIDR must be within 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, or 100.64.0.0/10

### spec.remoteNetworks.nodeCidrs

`[]string`

CIDR blocks of the on-premises network the hybrid NODES have their
addresses in. This is the range the control plane accepts kubelet
registrations from -- without a node's address inside one of these
blocks, it cannot join.

- rule: {"repeated":{"items":{"string":{"pattern":"^(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(?:\\.(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}/(?:[0-9]|[12]\\d|3[0-2])$"}}}}

### spec.remoteNetworks.podCidrs

`[]string`

CIDR blocks the on-premises PODS have their addresses in (the CNI's
pod network on the hybrid nodes). Optional: needed when pods must be
directly routable from the cluster (e.g. webhooks running on hybrid
nodes); node-only setups can omit it.

- rule: {"repeated":{"items":{"string":{"pattern":"^(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(?:\\.(?:25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}/(?:[0-9]|[12]\\d|3[0-2])$"}}}}

## Validation Rules

- `version_format`: version must be a Kubernetes minor version of 1.24 or later, e.g. '1.31'
- `control_plane_egress_mode_valid`: control_plane_egress_mode must be 'AWS_MANAGED', 'CUSTOMER_ROUTED', or 'CUSTOMER_ISOLATED' when set
- `ip_family_valid`: ip_family must be 'ipv4' or 'ipv6' when set
- `service_ipv4_cidr_format`: service_ipv4_cidr must be an IPv4 CIDR with a /12 to /24 prefix
- `service_ipv4_cidr_ipv4_only`: service_ipv4_cidr only applies when ip_family is 'ipv4'
- `log_types_valid`: each log type must be one of: api, audit, authenticator, controllerManager, scheduler
- `upgrade_support_type_valid`: upgrade_support_type must be 'STANDARD' or 'EXTENDED' when set
- `control_plane_scaling_tier_valid`: control_plane_scaling_tier must be 'standard', 'tier-xl', 'tier-2xl', 'tier-4xl', or 'tier-8xl' when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEksCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint` | `string` | endpoint is the URL of the Kubernetes API server for the EKS cluster. |
| `status.outputs.cluster_ca_certificate` | `string` | cluster_ca_certificate is the Base64-encoded certificate authority for the cluster. |
| `status.outputs.cluster_security_group_id` | `string` | cluster_security_group_id is the ID of the security group created by EKS for the cluster control plane. |
| `status.outputs.oidc_issuer_url` | `string` | oidc_issuer_url is the URL of the OpenID Connect issuer for the cluster (used for IAM Roles for Service Accounts). |
| `status.outputs.cluster_arn` | `string` | cluster_arn is the Amazon Resource Name of the EKS cluster. |
| `status.outputs.name` | `string` | name is the EKS cluster name. |
| `status.outputs.platform_version` | `string` | platform_version is the EKS platform version of the control plane (e.g. "eks.12") -- AWS's own revision of the control plane software for the cluster's Kubernetes minor version. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.clusterRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.autoMode.nodeRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBatchComputeEnvironment | `spec.eksConfiguration.eksClusterArn` | `status.outputs.cluster_arn` |
| AwsEksAccessEntry | `spec.clusterName` | `status.outputs.name` |
| AwsEksAddon | `spec.clusterName` | `status.outputs.name` |
| AwsEksFargateProfile | `spec.clusterName` | `status.outputs.name` |
| AwsEksNodeGroup | `spec.clusterName` | `status.outputs.name` |
| AwsIamOidcProvider | `spec.url` | `status.outputs.oidc_issuer_url` |
| AwsManagedPrometheusScraper | `spec.sourceEks.clusterArn` | `status.outputs.cluster_arn` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.cidr` | `spec.service_ipv4_cidr` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.except` | `spec.service_ipv4_cidr` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.cidr` | `spec.service_ipv4_cidr` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.except` | `spec.service_ipv4_cidr` |

## See Also

- [Overview](../README.md)
