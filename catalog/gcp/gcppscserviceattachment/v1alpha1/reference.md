# GcpPscServiceAttachment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPscServiceAttachmentSpec publishes a service through Private Service
Connect (PSC): the PRODUCER half. A service attachment sits in front of a
producer's internal load balancer -- the regional forwarding rule of an
internal passthrough Network Load Balancer or an internal Application
Load Balancer -- and lets consumers in other VPC networks (other
projects, other organizations) reach it through a PSC endpoint in their
own network, over private IPs, with no VPC peering and no overlapping-range
constraints. The consumer half is a regional GcpGlobalForwardingRule with
an empty load_balancing_scheme whose target is this attachment's
self_link.

Consumer traffic is NATed into the producer network through the
nat_subnets: dedicated subnets of purpose PRIVATE_SERVICE_CONNECT in the
producer VPC and region, sized for the number of consumer endpoints
(each connected endpoint consumes NAT addresses). connection_preference
decides who may connect: ACCEPT_AUTOMATIC lets any consumer connect,
ACCEPT_MANUAL admits only the projects, networks, or endpoints in
consumer_accept_lists (and refuses consumer_reject_lists), each with a
connection limit.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPscServiceAttachment
metadata:
  name: orders-db-psc
spec:
  projectId:
    value: my-gcp-project
  attachmentName: orders-db-psc
  description: Publishes the orders database's internal passthrough load balancer to consumer VPCs over Private Service Connect
  region: us-central1
  # The producer's internal load balancer: a regional GcpGlobalForwardingRule
  # (scheme INTERNAL) in the same region.
  targetService:
    valueFrom:
      kind: GcpGlobalForwardingRule
      name: orders-db-ilb
      fieldPath: status.outputs.self_link
  # PSC NAT subnets (purpose PRIVATE_SERVICE_CONNECT) in the producer VPC
  # and region; consumer traffic is translated into them.
  natSubnets:
    - valueFrom:
        kind: GcpSubnetwork
        name: orders-psc-nat
        fieldPath: status.outputs.subnetwork_self_link
  # Only the listed consumers connect; everyone else stays PENDING.
  connectionPreference: ACCEPT_MANUAL
  consumerAcceptLists:
    - projectId:
        value: consumer-analytics-prod
      connectionLimit: 10
    - network:
        value: https://www.googleapis.com/compute/v1/projects/partner-project/global/networks/partner-vpc
      connectionLimit: 2
  consumerRejectLists:
    - value: legacy-sandbox
  # A consumer removed from the accept list is disconnected, not left
  # connected until it reconnects.
  reconcileConnections: true
  enableProxyProtocol: false
  domainNames:
    - orders.internal.example.
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.attachmentName` | `string` |  |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.targetService` | `string \| valueFrom` | yes |  | GcpGlobalForwardingRule (`status.outputs.self_link`) |
| `spec.natSubnets` | `[]string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.connectionPreference` | `string` | yes |  |  |
| `spec.consumerAcceptLists` | `[]GcpPscServiceAttachmentConsumer` |  |  |  |
| `spec.consumerAcceptLists[].projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.consumerAcceptLists[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.consumerAcceptLists[].endpointUrl` | `string` |  |  |  |
| `spec.consumerAcceptLists[].connectionLimit` | `int32` | yes |  |  |
| `spec.consumerRejectLists` | `[]string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.reconcileConnections` | `bool` |  |  |  |
| `spec.enableProxyProtocol` | `bool` |  |  |  |
| `spec.domainNames` | `[]string` |  |  |  |
| `spec.propagatedConnectionLimit` | `int32` |  |  |  |
| `spec.showNatIps` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the service attachment (the producer
project). A literal project ID or a reference to a GcpProject. If
omitted, the provider's default project is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.attachmentName

`string`

Name of the service attachment in GCP. 1-63 characters, lowercase
letters, digits, and hyphens, starting with a letter and not ending
with a hyphen. Defaults to metadata.name. Immutable.

- rule: attachment_name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.region

`string` · required

The region of the attachment -- the same region as the target
forwarding rule and the NAT subnets. Required. Immutable.

- rule: region must be a valid GCP region name such as us-central1
- rule: {"required":true}

### spec.description

`string`

Human-readable description of the attachment.

- rule: {"string":{"maxLen":"2048"}}

### spec.targetService

`string | valueFrom` · required

The producer's load balancer: the regional forwarding rule (a
GcpGlobalForwardingRule with `region` set) of the internal passthrough
Network Load Balancer or internal Application Load Balancer that
serves the published service, in this attachment's region. Required;
changing it moves the published service in place.

- references: GcpGlobalForwardingRule (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalForwardingRule, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.natSubnets

`[]string | valueFrom` · required

The PSC NAT subnets consumer traffic is translated into: one or more
GcpSubnetwork of purpose PRIVATE_SERVICE_CONNECT in the producer VPC
and this region. Required, at least one. Size them for the consumer
endpoint count; add subnets to grow capacity. The provider treats the
list as a set, so order never diffs.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.connectionPreference

`string` · required

Who may connect:
- ACCEPT_AUTOMATIC: any consumer in any project connects without
  approval (subject to the reject list)
- ACCEPT_MANUAL: only consumers in consumer_accept_lists connect; every
  other connection request stays PENDING until accepted
Required. Mutable.

- rule: connection_preference must be ACCEPT_AUTOMATIC or ACCEPT_MANUAL
- rule: {"required":true}

### spec.consumerAcceptLists

`[]GcpPscServiceAttachmentConsumer`

Consumers admitted under ACCEPT_MANUAL, each with its connection limit.
The provider compares the list as a set.

- rule: exactly one of project_id, network, or endpoint_url names the consumer

### spec.consumerAcceptLists[].projectId

`string | valueFrom`

The consumer project allowed to connect: a project ID or number, or a
reference to a GcpProject. Every network in the project may connect.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.consumerAcceptLists[].network

`string | valueFrom`

The consumer VPC network allowed to connect (a network self-link or a
reference to a GcpVpcNetwork); narrower than a project.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.consumerAcceptLists[].endpointUrl

`string`

The specific consumer endpoint (a consumer forwarding rule's URL)
allowed to connect; the narrowest form.

### spec.consumerAcceptLists[].connectionLimit

`int32` · required

How many Private Service Connect endpoints this consumer may connect to
the attachment at once. Required, at least 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.consumerRejectLists

`[]string | valueFrom`

Consumer projects (IDs or numbers, or GcpProject references) refused
even under ACCEPT_AUTOMATIC. Compared as a set.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.reconcileConnections

`bool` · optional (explicit presence)

Whether a change to the accept or reject lists reconciles EXISTING
connections: false (Google's default) only affects PENDING endpoints,
so an accepted consumer stays connected after being removed from the
accept list; true moves existing ACCEPTED endpoints to REJECTED when
their project lands on the reject list, and vice versa. Unset lets
Google apply its default.

### spec.enableProxyProtocol

`bool`

Enable the PROXY protocol on connections through this attachment, so
the producer's backends receive the consumer's original TCP/IP address
data in the connection header. Required by the API (state it either
way); the backends must speak the protocol when true.

### spec.domainNames

`[]string`

Domain name registered with Cloud DNS for the connected endpoints, e.g.
"p.mycompany.com." (with the trailing dot). At most one. Immutable.

- rule: {"repeated":{"maxItems":"1","items":{"string":{"pattern":"^([a-z0-9-]+\\.)+$"}}}}

### spec.propagatedConnectionLimit

`int32` · optional (explicit presence)

How many consumer spokes a connected PSC endpoint may be propagated to
through Network Connectivity Center; per accept-list entry under
ACCEPT_MANUAL, per consumer project under ACCEPT_AUTOMATIC. Unset lets
Google apply its default (250); an explicit 0 is sent as 0.

- rule: {"int32":{"gte":0}}

### spec.showNatIps

`bool`

Show the NAT IP addresses of every connected endpoint in the
attachment's connected-endpoints listing. Google's API currently
ignores the flag (the provider records the value it sends); modeled so
the manifest states the intent it will honor once the API does.

### spec.deletionPolicy

`string`

What destroy does to the attachment:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the attachment is deleted; every consumer endpoint
               connected through it loses the service
  "PREVENT" -- destroy FAILS; protects a published service consumers
               depend on
  "ABANDON" -- the attachment leaves management but keeps serving

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `accept_lists_need_manual_preference`: consumer_accept_lists applies only with connection_preference ACCEPT_MANUAL; ACCEPT_AUTOMATIC admits every consumer

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPscServiceAttachment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the service attachment: the value a consumer forwarding rule names as its target. Format: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/serviceAttachments/{name} |
| `status.outputs.attachment_name` | `string` | Name of the attachment as it exists in GCP -- the resolved value (metadata.name when the spec left attachment_name empty). |
| `status.outputs.region` | `string` | Region of the attachment. |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint for optimistic concurrency control. |
| `status.outputs.connected_endpoints_count` | `string` | Number of consumer endpoints connected to the attachment at provisioning time (in any status); "0" for a freshly published service. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.targetService` | GcpGlobalForwardingRule | `status.outputs.self_link` |
| `spec.natSubnets` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.consumerAcceptLists[].projectId` | GcpProject | `status.outputs.project_id` |
| `spec.consumerAcceptLists[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.consumerRejectLists` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
