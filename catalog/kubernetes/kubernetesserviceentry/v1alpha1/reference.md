# KubernetesServiceEntry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesServiceEntrySpec defines an Istio ServiceEntry: a namespaced resource that
adds entries into Istio's internal service registry, so that mesh workloads can route
to and apply policy against services that are not part of the platform's own service
discovery (e.g. external APIs reached over the public internet, or VMs/legacy services
brought into the mesh).

100% fidelity with the upstream istio.io/api ServiceEntry
(networking/v1alpha3/service_entry.proto, served as networking.istio.io/v1), pinned to
the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly after the
Planton namespaced envelope (namespace); there is no nested
`service_entry` sub-message.

Attachment model (upstream): `endpoints` (static addresses) and `workload_selector`
(select in-mesh pods/VMs by label) are mutually exclusive — at most one may be set
(enforced below). Both omitted is valid (e.g. a `resolution: DNS` external host).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesServiceEntry
metadata:
  name: test-service-entry
spec:
  namespace:
    value: test-namespace
  hosts:
    - api.external-example.com
  addresses:
    - 198.51.100.10
  location: MESH_EXTERNAL
  resolution: STATIC
  ports:
    - number: 443
      name: https
      protocol: TLS
      target_port: 8443
  endpoints:
    - address: 198.51.100.10
      ports:
        https: 8443
      locality: us-west2/us-west2-a
  export_to:
    - "."
  subject_alt_names:
    - api.external-example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.hosts` | `[]string` | yes |  |  |
| `spec.addresses` | `[]string` |  |  |  |
| `spec.ports` | `[]KubernetesServiceEntryPort` |  |  |  |
| `spec.ports[].number` | `uint32` | yes |  |  |
| `spec.ports[].protocol` | `string` |  |  |  |
| `spec.ports[].name` | `string` | yes |  |  |
| `spec.ports[].targetPort` | `uint32` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.resolution` | `string` |  |  |  |
| `spec.endpoints` | `[]KubernetesServiceEntryEndpoint` |  |  |  |
| `spec.endpoints[].address` | `string` |  |  |  |
| `spec.endpoints[].ports` | `map<string, uint32>` |  |  |  |
| `spec.endpoints[].labels` | `map<string, string>` |  |  |  |
| `spec.endpoints[].network` | `string` |  |  |  |
| `spec.endpoints[].locality` | `string` |  |  |  |
| `spec.endpoints[].weight` | `uint32` |  |  |  |
| `spec.endpoints[].serviceAccount` | `string` |  |  |  |
| `spec.exportTo` | `[]string` |  |  |  |
| `spec.subjectAltNames` | `[]string` |  |  |  |
| `spec.workloadSelector` | `KubernetesIstioApiNetworkingWorkloadSelector` |  |  |  |
| `spec.workloadSelector.labels` | `map<string, string>` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the ServiceEntry is created. By default the service is visible to
the whole mesh; `export_to` narrows the visibility scope.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.hosts

`[]string` · required

The hosts associated with the ServiceEntry. Used to match against the destination of
requests (the HTTP Host/Authority header, the TLS SNI, or — with resolution DNS and no
endpoints — the DNS name to resolve). Required, at least one. A bare wildcard `*` is
not a valid host (partial wildcards such as `*.example.com` are allowed). Upstream
allows up to 256.

- rule: {"repeated":{"minItems":"1","maxItems":"256","items":{"cel":[{"id":"service_entry_host.not_bare_wildcard","message":"host cannot be a bare wildcard '*'","expression":"this != '*'"}],"string":{"minLen":"1"}}}}

### spec.addresses

`[]string`

The virtual IP addresses (or CIDR prefixes) associated with the service. Used to match
the destination IP of requests; if empty, requests are matched by port only. CIDR
prefixes are honored only with NONE or STATIC resolution (enforced above). Upstream
allows up to 256, each up to 64 characters.

- rule: {"repeated":{"maxItems":"256","items":{"string":{"maxLen":"64"}}}}

### spec.ports

`[]KubernetesServiceEntryPort`

The ports exposed by the external service. Each port's `name` and `number` must be
unique across the list (enforced above). Upstream allows up to 256.

- rule: {"repeated":{"maxItems":"256"}}

### spec.ports[].number

`uint32` · required

A valid port number (1-65535). Required.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.ports[].protocol

`string` · optional (explicit presence)

The protocol exposed on the port. The upstream documentation requires one of
HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, or TLS (TLS implies SNI-based routing without
terminating the connection). Optional; when unset the proxy treats the port as TCP.

- rule: {"string":{"in":["HTTP","HTTPS","GRPC","HTTP2","MONGO","TCP","TLS"]}}

### spec.ports[].name

`string` · required

The label assigned to the port. Required and unique across the ServiceEntry's ports
(uniqueness enforced at the spec level). Upstream allows up to 256 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.ports[].targetPort

`uint32` · optional (explicit presence)

The port on the backing endpoint where traffic is received. When unset, defaults to
`number`.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.location

`string` · optional (explicit presence)

Whether the service should be considered external to the mesh (MESH_EXTERNAL) or part
of the mesh (MESH_INTERNAL). Unset defaults to MESH_EXTERNAL upstream. MESH_INTERNAL is
typically used together with `workload_selector` to bring unmanaged infrastructure
(e.g. VMs) into the mesh.

- rule: {"string":{"in":["MESH_EXTERNAL","MESH_INTERNAL"]}}

### spec.resolution

`string` · optional (explicit presence)

How the sidecar/proxy resolves the service's endpoint IPs. Unset defaults to NONE
upstream. NONE forwards to the original destination IP (no endpoints); STATIC uses the
IPs in `endpoints`; DNS resolves the hosts (or endpoint addresses) asynchronously;
DNS_ROUND_ROBIN is like DNS but pins to the first resolved IP per new connection;
DYNAMIC_DNS resolves the ACTUAL requested hostname at request time (a dynamic
forward proxy for wildcard hosts — HTTP-family and TLS ports only, no addresses
or endpoints; enforced above, mirroring the istiod webhook).

- rule: {"string":{"in":["NONE","STATIC","DNS","DNS_ROUND_ROBIN","DYNAMIC_DNS"]}}

### spec.endpoints

`[]KubernetesServiceEntryEndpoint`

Static endpoints backing the service. Mutually exclusive with `workload_selector`
(enforced above). Not permitted with NONE resolution; at most one with
DNS_ROUND_ROBIN. Upstream allows up to 4096.

- rule: {"repeated":{"maxItems":"4096"}}
- rule: endpoint requires an address or a network
- rule: a unix:// endpoint address may not include ports

### spec.endpoints[].address

`string` · optional (explicit presence)

The endpoint address without a port (an IP, or a hostname when resolution is DNS, or
a `unix://` domain socket path). If empty, `network` is required. Upstream allows up
to 256 characters.

- rule: a unix:// endpoint address must be an absolute path or an abstract socket (@)
- rule: a unix:// endpoint address may not be a directory (trailing '/')
- rule: {"string":{"maxLen":"256"}}

### spec.endpoints[].ports

`map<string, uint32>`

The service-port-name -> endpoint-port mapping. Keys are port names (matching a
ServiceEntry port `name`); values are the port on this endpoint. Do not use with a
`unix://` address. Upstream allows up to 128 entries.

- rule: {"map":{"maxPairs":"128","keys":{"string":{"maxLen":"63","pattern":"^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}},"values":{"uint32":{"lte":65535,"gte":1}}}}

### spec.endpoints[].labels

`map<string, string>`

Labels associated with the endpoint. Upstream allows up to 256 entries.

- rule: {"map":{"maxPairs":"256"}}

### spec.endpoints[].network

`string` · optional (explicit presence)

The L3 network the endpoint belongs to, used to group endpoints across clusters.
Required when `address` is empty. Upstream allows up to 2048 characters.

- rule: {"string":{"maxLen":"2048"}}

### spec.endpoints[].locality

`string` · optional (explicit presence)

The locality (region/zone/sub-zone) of the endpoint, used for locality-aware load
balancing (e.g. `us-west2/us-west2-a`). Upstream allows up to 2048 characters.

- rule: {"string":{"maxLen":"2048"}}

### spec.endpoints[].weight

`uint32` · optional (explicit presence)

The load-balancing weight of the endpoint; higher values receive proportionally more
traffic.

### spec.endpoints[].serviceAccount

`string` · optional (explicit presence)

The service account associated with the workload when a sidecar is present. Must be in
the same namespace as the ServiceEntry. Upstream allows up to 253 characters.

- rule: {"string":{"maxLen":"253"}}

### spec.exportTo

`[]string`

The namespaces to which this service is exported (made visible). Default is all
namespaces. `.` exports to the declaring namespace only; `*` exports to all.

### spec.subjectAltNames

`[]string`

If set, the proxy verifies that the server certificate's subject alternate name
matches one of these values. Applies when originating TLS to the external service.

### spec.workloadSelector

`KubernetesIstioApiNetworkingWorkloadSelector`

Selects in-mesh workloads (by pod/VM label) that back this service, instead of listing
static `endpoints`. Mutually exclusive with `endpoints` (enforced above). Only
meaningful with MESH_INTERNAL location.

INFRA-CHART COMPOSABILITY: workload_selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod against pod
labels and creates NO automatic DAG edge to any workload resource. To order this
ServiceEntry after the workloads it fronts in an infra chart, an author MUST express
the dependency via metadata.relationships, e.g.:
  metadata:
    relationships:
      - kind: KubernetesDeployment
        name: "{{ values.app }}"
        type: depends_on
See the kind's "Composing in Infra Charts" docs for the full pattern.

### spec.workloadSelector.labels

`map<string, string>`

One or more labels indicating the set of pods/VMs the resource applies to.
Faithful to istio.io/api `istio.networking.v1alpha3.WorkloadSelector.labels`,
whose ServiceEntry CRD constraints are: max 256 entries; each value <= 63 chars;
and wildcards ('*') are not permitted in values. NOTE: unlike the policy selector
(`match_labels` above), the ServiceEntry CRD does NOT enforce non-empty keys or
no-wildcard keys, so those two rules are deliberately omitted here — adding them would
reject configurations the CRD accepts (match the upstream validated outcome).

- rule: wildcard ('*') is not allowed in label selector values
- rule: {"map":{"maxPairs":"256","values":{"string":{"maxLen":"63"}}}}

## Validation Rules

- `service_entry.workload_selector_xor_endpoints`: at most one of workload_selector or endpoints may be set
- `service_entry.cidr_addresses_require_none_or_static`: CIDR addresses are allowed only for NONE or STATIC resolution
- `service_entry.none_resolution_forbids_endpoints`: endpoints cannot be set when resolution is NONE
- `service_entry.dns_round_robin_single_endpoint`: DNS_ROUND_ROBIN resolution allows at most one endpoint
- `service_entry.dynamic_dns_wildcard_hosts`: every host must be wildcarded (e.g. '*.example.com') when resolution is DYNAMIC_DNS
- `service_entry.dynamic_dns_no_addresses`: addresses cannot be set when resolution is DYNAMIC_DNS — destinations are derived from the wildcard hosts
- `service_entry.dynamic_dns_no_endpoints`: endpoints cannot be set when resolution is DYNAMIC_DNS — destination IPs are resolved from the requested hostname
- `service_entry.dynamic_dns_http_family_ports`: only HTTP, TLS, GRPC, and HTTP2 port protocols are supported when resolution is DYNAMIC_DNS
- `service_entry.port_numbers_unique`: port number cannot be duplicated across ports
- `service_entry.port_names_unique`: port name cannot be duplicated across ports

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesServiceEntry, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_entry_name` | `string` | Name of the created ServiceEntry (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the ServiceEntry was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
