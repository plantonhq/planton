# GcpRegionNetworkEndpointGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpRegionNetworkEndpointGroupSpec defines a regional network endpoint group
(NEG) — the bridge that lets a load balancer's backend service send traffic
to something other than a group of Compute Engine VMs. A regional NEG names
a set of endpoints in one region; the backend service references the NEG's
self-link in its backends[].group.

The endpoint type decides what the NEG points at:
  - SERVERLESS (the default) — a Cloud Run service, a Cloud Functions
    function, or an App Engine service. This is how serverless workloads sit
    behind an external Application Load Balancer (custom domains, Cloud CDN,
    Cloud Armor, IAP in front of Cloud Run).
  - PRIVATE_SERVICE_CONNECT — a Private Service Connect endpoint fronting a
    published producer service or a Google API.
  - INTERNET_IP_PORT / INTERNET_FQDN_PORT — an external origin (an IP:port or
    an FQDN:port) reached over the internet, e.g. an on-prem or third-party
    backend fronted by a Google load balancer.
  - GCE_VM_IP_PORTMAP — PSC port-mapping to VM IP:port targets.

The whole resource is immutable in GCP: every field is ForceNew, so any
change destroys and recreates the NEG. Because an in-use NEG cannot be
deleted, downstream tooling that recreates one should create the replacement
before destroying the original (create-before-destroy) to avoid a
resourceInUseByAnotherResource error.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRegionNetworkEndpointGroup
metadata:
  name: my-sample-region-neg
spec:
  # GCP project that owns the NEG.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  networkEndpointGroupName: web-neg

  # Region — required; must match the region of the serverless workload.
  region: us-central1

  # SERVERLESS is the default; front a Cloud Run service.
  cloudRun:
    service:
      value: my-cloud-run-service

  # DELETE keeps destroys real; PREVENT belongs on NEGs a production
  # backend service routes through.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.networkEndpointGroupName` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.networkEndpointType` | `string` |  | `SERVERLESS` |  |
| `spec.description` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.pscTargetService` | `string` |  |  |  |
| `spec.pscData` | `GcpRegionNetworkEndpointGroupPscData` |  |  |  |
| `spec.pscData.producerPort` | `string` |  |  |  |
| `spec.cloudRun` | `GcpRegionNetworkEndpointGroupCloudRun` |  |  |  |
| `spec.cloudRun.service` | `string \| valueFrom` |  |  | GcpCloudRun (`status.outputs.service_name`) |
| `spec.cloudRun.tag` | `string` |  |  |  |
| `spec.cloudRun.urlMask` | `string` |  |  |  |
| `spec.cloudFunction` | `GcpRegionNetworkEndpointGroupCloudFunction` |  |  |  |
| `spec.cloudFunction.function` | `string \| valueFrom` |  |  | GcpCloudFunction (`status.outputs.name`) |
| `spec.cloudFunction.urlMask` | `string` |  |  |  |
| `spec.appEngine` | `GcpRegionNetworkEndpointGroupAppEngine` |  |  |  |
| `spec.appEngine.service` | `string` |  |  |  |
| `spec.appEngine.version` | `string` |  |  |  |
| `spec.appEngine.urlMask` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the network endpoint group.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the NEG.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.networkEndpointGroupName

`string`

Name of the network endpoint group in GCP. Must be 1-63 characters:
lowercase letters, digits, and hyphens; must start with a letter and end
with a letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the NEG, briefly breaking
every backend service that references the old self_link.

- rule: network_endpoint_group_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.region

`string`

Region the NEG lives in (e.g. "us-central1"). Required — a regional NEG is
always scoped to one region, and a serverless NEG must be in the same
region as the Cloud Run/Functions/App Engine workload it fronts.
Immutable: a NEG cannot move between regions.

- rule: region is required and must be a valid GCP region name such as us-central1

### spec.networkEndpointType

`string` · optional (explicit presence)

What kind of endpoints this NEG holds (default SERVERLESS). SERVERLESS
fronts Cloud Run/Functions/App Engine; PRIVATE_SERVICE_CONNECT fronts a
PSC endpoint; INTERNET_IP_PORT / INTERNET_FQDN_PORT front an external
origin; GCE_VM_IP_PORTMAP does PSC port mapping. Immutable.

- default: `SERVERLESS`
- rule: network_endpoint_type must be one of SERVERLESS, PRIVATE_SERVICE_CONNECT, INTERNET_IP_PORT, INTERNET_FQDN_PORT, or GCE_VM_IP_PORTMAP

### spec.description

`string`

What this NEG fronts and which backend service consumes it — write it for
the operator tracing a request path later. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.network

`string | valueFrom`

The VPC network for PRIVATE_SERVICE_CONNECT, INTERNET, and
GCE_VM_IP_PORTMAP NEGs. Reference a GcpVpcNetwork or provide a network self-link
directly. Not used by (and rejected for) SERVERLESS NEGs — serverless
platforms are not attached to a VPC here. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom`

The subnetwork for PRIVATE_SERVICE_CONNECT and GCE_VM_IP_PORTMAP NEGs.
Reference a GcpSubnetwork or provide a subnetwork self-link directly.
Not used by serverless or internet NEGs. Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.pscTargetService

`string`

The target of a PRIVATE_SERVICE_CONNECT or INTERNET NEG. For PSC it is the
published service-attachment URL or a Google API bundle name (e.g.
"asia-northeast3-cloudkms.googleapis.com"); required for PSC NEGs.
Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.pscData

`GcpRegionNetworkEndpointGroupPscData`

Extra Private Service Connect settings for a PSC NEG. Only valid when
network_endpoint_type is PRIVATE_SERVICE_CONNECT.

### spec.pscData.producerPort

`string`

The producer port a consumer PSC NEG connects to. Empty connects to the
first port in the producer's advertised port range. Immutable.

- rule: {"string":{"maxLen":"16"}}

### spec.cloudRun

`GcpRegionNetworkEndpointGroupCloudRun`

Front a Cloud Run service.

- rule: a cloud_run block must set service or url_mask (or both)

### spec.cloudRun.service

`string | valueFrom`

The Cloud Run service to route to. Reference a GcpCloudRun resource or
provide the service name directly. GCP resolves endpoints at serving time,
so the service need not exist when the NEG is created. Immutable.

- references: GcpCloudRun (`status.outputs.service_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudRun, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.cloudRun.tag

`string`

Route to a specific Cloud Run revision tag for fine-grained traffic
splitting (e.g. "canary"). Only meaningful with service. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.cloudRun.urlMask

`string`

A URL template that parses the service (and optional tag) out of each
request URL — for host/path-based fan-out to many services from one NEG
(e.g. "<service>.example.com/<tag>"). Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.cloudFunction

`GcpRegionNetworkEndpointGroupCloudFunction`

Front a Cloud Functions (Gen 2) function.

- rule: a cloud_function block must set function or url_mask (or both)

### spec.cloudFunction.function

`string | valueFrom`

The Cloud Functions function name to route to (case-sensitive). Accepts
a literal name or a reference to a GcpCloudFunction resource. GCP
resolves endpoints at serving time. Immutable.

- references: GcpCloudFunction (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudFunction, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.cloudFunction.urlMask

`string`

A URL template that parses the function name out of each request URL — for
routing many functions from one NEG. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.appEngine

`GcpRegionNetworkEndpointGroupAppEngine`

Front an App Engine service. The block may be empty to route to the
default App Engine application.

### spec.appEngine.service

`string`

The App Engine service to route to. Empty routes to the default service.
Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.appEngine.version

`string`

The App Engine service version to route to. Empty routes to the version
split configured in App Engine. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.appEngine.urlMask

`string`

A URL template that parses the service (and version) out of each request
URL. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.deletionPolicy

`string`

Deletion policy for the NEG — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the NEG is deleted (GCP refuses while a backend
               service still references it — create the replacement
               before destroying the original)
  "PREVENT" -- destroy FAILS; protects the attachment a live backend
               service routes through
  "ABANDON" -- the NEG is removed from management but left in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `serverless_requires_exactly_one_block`: a SERVERLESS network endpoint group requires exactly one of cloud_run, cloud_function, or app_engine
- `serverless_blocks_forbidden_otherwise`: cloud_run, cloud_function, and app_engine apply only to SERVERLESS network endpoint groups
- `psc_requires_target_service`: a PRIVATE_SERVICE_CONNECT network endpoint group requires psc_target_service (the published service-attachment URL or Google API bundle)
- `psc_target_service_scope`: psc_target_service applies only to PRIVATE_SERVICE_CONNECT and INTERNET network endpoint groups
- `psc_data_scope`: psc_data applies only to PRIVATE_SERVICE_CONNECT network endpoint groups
- `subnetwork_scope`: subnetwork applies only to PRIVATE_SERVICE_CONNECT and GCE_VM_IP_PORTMAP network endpoint groups
- `network_not_for_serverless`: network applies only to non-serverless network endpoint groups (PSC, INTERNET, or GCE_VM_IP_PORTMAP)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpRegionNetworkEndpointGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the network endpoint group. This is the value a backend service references in backends[].group — the composition handle that puts this NEG behind a load balancer. Format: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/networkEndpointGroups/{name} |
| `status.outputs.network_endpoint_group_name` | `string` | Name of the network endpoint group as it exists in GCP. |
| `status.outputs.network_endpoint_type` | `string` | The endpoint type of the NEG (SERVERLESS, PRIVATE_SERVICE_CONNECT, INTERNET_IP_PORT, INTERNET_FQDN_PORT, or GCE_VM_IP_PORTMAP), echoed for tooling that resolves the serving chain. |
| `status.outputs.region` | `string` | Region the network endpoint group lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.cloudRun.service` | GcpCloudRun | `status.outputs.service_name` |
| `spec.cloudFunction.function` | GcpCloudFunction | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBackendService | `spec.backends[].group` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
