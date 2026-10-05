# GcpServiceConnectionPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpServiceConnectionPolicySpec defines a service connection policy —
the per-network authorization that lets Google's service connectivity
automation create Private Service Connect endpoints in your VPC on a
producer's behalf.

Newer Google managed services (Memorystore for Valkey, Redis Cluster,
and other PSC-first producers) do not peer with your network the way
private services access does. Instead, when you create an instance,
the producer asks the connectivity automation to place PSC endpoints
inside your subnets — and the automation refuses unless a policy for
that producer's service class already exists on the network in that
region. This resource is that policy: without it, instance creation
fails with a connectivity error; with it, endpoints appear
automatically and their IPs surface on the instance.

One policy exists per (network, service class, region) triple. The
policy names the service class it authorizes (gcp-memorystore for
Memorystore for Valkey, gcp-memorystore-redis for Redis Cluster —
Google publishes the identifier per service; third-party producers
publish their own), the subnets the automation may draw endpoint IPs
from, and an optional connection limit.

Deploy the policy before the first instance of its service class in
that region, and keep it alive as long as any instance depends on it —
deleting the policy strands existing endpoints and blocks new ones.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceConnectionPolicy
metadata:
  name: my-sample-scp
spec:
  # GCP project owning the network and the policy.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Region the policy applies to — service connectivity automation is
  # regional, so each region needs its own policy.
  location: us-central1

  # The consumer VPC network — a resource path (a GcpVpcNetwork reference
  # resolves to it automatically).
  network:
    value: projects/my-gcp-project-123/global/networks/my-vpc

  # The producer's published service class. gcp-memorystore authorizes
  # Memorystore for Valkey instances on this network.
  serviceClass: gcp-memorystore

  # Subnets the automation draws PSC endpoint IPs from, plus an optional
  # connection cap.
  pscConfig:
    subnetworks:
      - value: projects/my-gcp-project-123/regions/us-central1/subnetworks/my-subnet
    limit: 10

  # DELETE keeps destroys real; PREVENT belongs on policies live managed
  # instances depend on.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.policyName` | `string` |  |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.serviceClass` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.pscConfig` | `GcpServiceConnectionPolicyPscConfig` |  |  |  |
| `spec.pscConfig.subnetworks` | `[]string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.pscConfig.limit` | `int32` |  |  |  |
| `spec.pscConfig.producerInstanceLocation` | `string` |  |  |  |
| `spec.pscConfig.allowedGoogleProducersResourceHierarchyLevels` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project that owns the network and the policy. Can be a literal
project ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.policyName

`string`

Name for the policy resource. If not specified, defaults to
metadata.name. Must start with a lowercase letter, contain only
lowercase letters, numbers, and hyphens, and end with a letter or
number. Immutable after creation.

- rule: policy_name must start with a lowercase letter, use only lowercase letters, numbers, and hyphens, and end with a letter or number (max 63 characters)

### spec.location

`string` · required

GCP region the policy applies to (e.g. us-central1). Service
connectivity automation is regional: a producer instance in another
region needs its own policy there. Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.network

`string | valueFrom` · required

The consumer VPC network this policy authorizes connections into.
A reference resolves to the GcpVpcNetwork's resource path
(projects/{project}/global/networks/{name}) — the format the Service
Connectivity API requires; both engines also normalize full self-link
URLs to that path. Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.serviceClass

`string` · required

The service class this policy authorizes — the producer's published
identifier. Google services use a gcp- prefix (gcp-memorystore for
Memorystore for Valkey, gcp-memorystore-redis for Redis Cluster);
third-party producers publish their own class names. One policy
exists per (network, service class, region). Immutable after
creation.

- rule: {"required":true,"string":{"maxLen":"255","pattern":"^[a-z][a-z0-9-]*$"}}

### spec.description

`string`

Free-text description of what this policy is for — which workloads
or teams the authorized service class serves.

- rule: {"string":{"maxLen":"1024"}}

### spec.labels

`map<string, string>`

User-defined labels to organize and track the policy. Merged beneath
Planton's platform attribution labels (platform keys win on
conflict).

### spec.pscConfig

`GcpServiceConnectionPolicyPscConfig`

Private Service Connect configuration: the subnets endpoint IPs come
from, the connection limit, and the optional producer-location
allowlist. Required in practice — the policy authorizes nothing
usable without address space — but optional in the API, so presets
always set it.

- rule: producer_instance_location CUSTOM_RESOURCE_HIERARCHY_LEVELS requires at least one entry in allowed_google_producers_resource_hierarchy_levels
- rule: allowed_google_producers_resource_hierarchy_levels only takes effect when producer_instance_location is CUSTOM_RESOURCE_HIERARCHY_LEVELS

### spec.pscConfig.subnetworks

`[]string | valueFrom` · required

Subnets the connectivity automation allocates PSC endpoint IPs from.
References resolve to the GcpSubnetwork's self-link; both engines
normalize self-link URLs to the relative resource path
(projects/{project}/regions/{region}/subnetworks/{name}) the Service
Connectivity API expects. The subnets must live in the same region as
the policy and inside the policy's network. At least one is required.
Regular-purpose subnets work — no special PSC purpose is needed for
service connection policies (unlike PSC NAT subnets for published
services).

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.pscConfig.limit

`int32`

Maximum number of PSC connections the automation may create under this
policy. 0 (unset) leaves the limit to GCP's default. Useful as a
guardrail in shared networks: it caps how many managed-service
instances can attach through this policy before an operator has to
deliberately raise it.

- rule: {"int32":{"gte":0}}

### spec.pscConfig.producerInstanceLocation

`string`

Authorization mechanism deciding which producer projects the
automation will connect to. Leave empty for GCP's default behavior
(any producer instance the consumer project provisions).
CUSTOM_RESOURCE_HIERARCHY_LEVELS restricts producers to the resource
hierarchy entries listed in
allowed_google_producers_resource_hierarchy_levels.

- rule: producer_instance_location must be PRODUCER_INSTANCE_LOCATION_UNSPECIFIED or CUSTOM_RESOURCE_HIERARCHY_LEVELS

### spec.pscConfig.allowedGoogleProducersResourceHierarchyLevels

`[]string`

Projects, folders, or organizations producer instances may live in,
each as 'projects/{id-or-number}', 'folders/{number}', or
'organizations/{number}' (e.g. projects/my-project-id, folders/891,
organizations/123). Only consulted when producer_instance_location is
CUSTOM_RESOURCE_HIERARCHY_LEVELS.

- rule: {"repeated":{"items":{"cel":[{"id":"hierarchy_level_format","message":"each entry must be projects/{id-or-number}, folders/{number}, or organizations/{number}","expression":"this.matches('^(projects/[a-z0-9-]+|projects/[0-9]+|folders/[0-9]+|organizations/[0-9]+)$')"}]}}}

### spec.deletionPolicy

`string`

Deletion policy for the policy — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy is deleted (existing PSC endpoints are
               stranded and new instances of the service class can no
               longer connect in this region)
  "PREVENT" -- destroy FAILS; protects the connectivity every managed
               instance under this policy depends on
  "ABANDON" -- the policy is removed from management but keeps
               authorizing connections in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpServiceConnectionPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | Fully qualified policy resource path (projects/{project}/locations/{location}/serviceConnectionPolicies/{name}). The canonical identifier for API calls and audit tooling. |
| `status.outputs.name` | `string` | Short name of the policy — what an operator sees in the console's service connection policy list for the network. |
| `status.outputs.infrastructure` | `string` | The type of underlying resources the automation creates for this policy (PSC for Private Service Connect). Confirms the connectivity mechanism without inspecting individual connections. |
| `status.outputs.etag` | `string` | Server-computed etag. Changes on every policy mutation — useful for change detection when auditing shared networks. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.pscConfig.subnetworks` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## See Also

- [Overview](../README.md)
