# KubernetesEnvoyFilter

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesEnvoyFilterSpec defines an Istio EnvoyFilter: a namespaced, expert-only
escape hatch that customizes the Envoy proxy configuration istiod generates for
selected workloads. It applies a list of ordered patches (add/merge/remove/insert/
replace) to low-level Envoy xDS objects (listeners, filter chains, network/HTTP
filters, route configurations, virtual hosts, routes, and clusters).

100% fidelity with the upstream istio.io/api EnvoyFilter
(networking/v1alpha3/envoy_filter.proto, served as networking.istio.io/v1alpha3 — this
is the only typed Istio API kind still on v1alpha3; it has NOT graduated to v1),
pinned to the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly after
the Planton namespaced envelope (namespace); there is no nested
`envoy_filter` sub-message.

EXPERT-ONLY ESCAPE HATCH: EnvoyFilter patches Envoy's internal xDS API directly. The
patch payload (`config_patches[].patch.value`) is free-form JSON (google.protobuf.Struct)
that istiod merges into generated config with no schema validation — a malformed patch
can break a workload's traffic. Prefer first-class typed Istio APIs where they exist; use
EnvoyFilter only for capabilities not yet modeled by a higher-level API.

Attachment model (upstream): `workload_selector` (select pods/VMs by label) and
`target_refs` (attach to a Gateway/GatewayClass/Service/ServiceEntry) are mutually
exclusive — at most one may be set (enforced below). Both omitted is valid: the filter
then applies to all workloads in the resource's namespace (or, in the mesh root namespace,
to all applicable workloads mesh-wide).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesEnvoyFilter
metadata:
  name: test-envoy-filter
spec:
  namespace:
    value: test-namespace
  workload_selector:
    labels:
      app: test-app
  priority: 10
  config_patches:
    - apply_to: HTTP_FILTER
      match:
        context: SIDECAR_INBOUND
        proxy:
          proxy_version: "^1\\.30.*"
        listener:
          port_number: 8080
          filter_chain:
            filter:
              name: envoy.filters.network.http_connection_manager
              sub_filter:
                name: envoy.filters.http.router
      patch:
        operation: INSERT_BEFORE
        filter_class: AUTHZ
        value:
          name: envoy.filters.http.lua
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
            inline_code: |
              function envoy_on_request(handle)
                handle:headers():add("x-planton", "test")
              end
    - apply_to: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        cluster:
          service: reviews.default.svc.cluster.local
      patch:
        operation: MERGE
        value:
          connect_timeout: 5s
    - apply_to: HTTP_FILTER
      match:
        context: WAYPOINT
        waypoint:
          port_number: 8080
          filter:
            name: envoy.filters.network.http_connection_manager
            sub_filter:
              name: envoy.filters.http.router
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.cors
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.workloadSelector` | `KubernetesIstioApiNetworkingWorkloadSelector` |  |  |  |
| `spec.workloadSelector.labels` | `map<string, string>` |  |  |  |
| `spec.configPatches` | `[]KubernetesEnvoyFilterConfigPatch` |  |  |  |
| `spec.configPatches[].applyTo` | `string` |  |  |  |
| `spec.configPatches[].match` | `KubernetesEnvoyFilterEnvoyConfigObjectMatch` |  |  |  |
| `spec.configPatches[].match.context` | `string` |  |  |  |
| `spec.configPatches[].match.proxy` | `KubernetesEnvoyFilterProxyMatch` |  |  |  |
| `spec.configPatches[].match.proxy.proxyVersion` | `string` |  |  |  |
| `spec.configPatches[].match.proxy.metadata` | `map<string, string>` |  |  |  |
| `spec.configPatches[].match.listener` | `KubernetesEnvoyFilterListenerMatch` |  |  |  |
| `spec.configPatches[].match.listener.portNumber` | `uint32` |  |  |  |
| `spec.configPatches[].match.listener.filterChain` | `KubernetesEnvoyFilterFilterChainMatch` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.name` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.sni` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.transportProtocol` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.applicationProtocols` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.filter` | `KubernetesEnvoyFilterFilterMatch` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.filter.name` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.filter.subFilter` | `KubernetesEnvoyFilterSubFilterMatch` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.filter.subFilter.name` | `string` |  |  |  |
| `spec.configPatches[].match.listener.filterChain.destinationPort` | `uint32` |  |  |  |
| `spec.configPatches[].match.listener.listenerFilter` | `string` |  |  |  |
| `spec.configPatches[].match.listener.name` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration` | `KubernetesEnvoyFilterRouteConfigurationMatch` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.portNumber` | `uint32` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.portName` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.gateway` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost` | `KubernetesEnvoyFilterVirtualHostMatch` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost.name` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost.domainName` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost.route` | `KubernetesEnvoyFilterRouteMatch` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost.route.name` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.vhost.route.action` | `string` |  |  |  |
| `spec.configPatches[].match.routeConfiguration.name` | `string` |  |  |  |
| `spec.configPatches[].match.cluster` | `KubernetesEnvoyFilterClusterMatch` |  |  |  |
| `spec.configPatches[].match.cluster.portNumber` | `uint32` |  |  |  |
| `spec.configPatches[].match.cluster.service` | `string` |  |  |  |
| `spec.configPatches[].match.cluster.subset` | `string` |  |  |  |
| `spec.configPatches[].match.cluster.name` | `string` |  |  |  |
| `spec.configPatches[].match.waypoint` | `KubernetesEnvoyFilterWaypointMatch` |  |  |  |
| `spec.configPatches[].match.waypoint.filter` | `KubernetesEnvoyFilterWaypointFilterMatch` |  |  |  |
| `spec.configPatches[].match.waypoint.filter.name` | `string` |  |  |  |
| `spec.configPatches[].match.waypoint.filter.subFilter` | `KubernetesEnvoyFilterWaypointSubFilterMatch` |  |  |  |
| `spec.configPatches[].match.waypoint.filter.subFilter.name` | `string` |  |  |  |
| `spec.configPatches[].match.waypoint.portNumber` | `uint32` |  |  |  |
| `spec.configPatches[].match.waypoint.route` | `KubernetesEnvoyFilterWaypointRouteMatch` |  |  |  |
| `spec.configPatches[].match.waypoint.route.name` | `string` |  |  |  |
| `spec.configPatches[].patch` | `KubernetesEnvoyFilterPatch` |  |  |  |
| `spec.configPatches[].patch.operation` | `string` |  |  |  |
| `spec.configPatches[].patch.value` | `object` |  |  |  |
| `spec.configPatches[].patch.filterClass` | `string` |  |  |  |
| `spec.priority` | `int32` |  |  |  |
| `spec.targetRefs` | `[]KubernetesIstioApiPolicyTargetReference` |  |  |  |
| `spec.targetRefs[].group` | `string` |  |  |  |
| `spec.targetRefs[].kind` | `string` | yes |  |  |
| `spec.targetRefs[].name` | `string \| valueFrom` | yes |  | KubernetesGateway (`status.outputs.gateway_name`) |
| `spec.targetRefs[].namespace` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the EnvoyFilter is created. An EnvoyFilter in a workload namespace
affects only workloads in that namespace; one in the Istio mesh root namespace (e.g.
istio-system) affects all applicable workloads mesh-wide.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.workloadSelector

`KubernetesIstioApiNetworkingWorkloadSelector`

Selects the pods/VMs on which these patches are applied, by label. If omitted, the
patches apply to all workload instances in the same namespace (or, in the mesh root
namespace, to all applicable workloads mesh-wide). Mutually exclusive with `target_refs`
(enforced above). Reuses the shared networking/v1alpha3 WorkloadSelector — the
EnvoyFilter CRD applies the SAME selector constraints as ServiceEntry (max 256 labels,
each value <= 63 chars, no wildcard values), so the shared type is faithful here.

INFRA-CHART COMPOSABILITY: workload_selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod against pod
labels and creates NO automatic DAG edge to any workload resource. To order this
EnvoyFilter after the workloads it patches in an infra chart, an author MUST express the
dependency via metadata.relationships, e.g.:
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

### spec.configPatches

`[]KubernetesEnvoyFilterConfigPatch`

One or more patches with match conditions, applied in list order within a context.
An EnvoyFilter with no patches is valid upstream (a no-op), so this list is not
required.

### spec.configPatches[].applyTo

`string` · optional (explicit presence)

Where in the Envoy configuration the patch should be applied. The match is expected to
select the appropriate object based on this value (e.g. HTTP_FILTER expects a listener
match with a network-filter selection on the HTTP connection manager and a sub-filter
selection; CLUSTER expects a cluster match, not a listener match). Unset leaves it
unspecified upstream. external standard exception — matches the Istio EnvoyFilter.ApplyTo
enum (BOOTSTRAP is accepted for fidelity but is DEPRECATED upstream).

- rule: {"string":{"in":["LISTENER","FILTER_CHAIN","NETWORK_FILTER","HTTP_FILTER","ROUTE_CONFIGURATION","VIRTUAL_HOST","HTTP_ROUTE","CLUSTER","EXTENSION_CONFIG","BOOTSTRAP","LISTENER_FILTER"]}}

### spec.configPatches[].match

`KubernetesEnvoyFilterEnvoyConfigObjectMatch`

Match conditions that must be met before the patch is applied. If omitted, the patch
applies broadly (subject to apply_to and context defaults).

- rule: at most one of listener, route_configuration, cluster, or waypoint may be set

### spec.configPatches[].match.context

`string` · optional (explicit presence)

The config-generation context to match on: istiod generates Envoy config in the context
of a gateway, inbound sidecar traffic, outbound sidecar traffic, or an ambient waypoint
proxy. Unset matches ANY (all objects in sidecars, gateways, and waypoints) upstream.
external standard exception — matches the Istio EnvoyFilter.PatchContext enum.

- rule: {"string":{"in":["ANY","SIDECAR_INBOUND","SIDECAR_OUTBOUND","GATEWAY","WAYPOINT"]}}

### spec.configPatches[].match.proxy

`KubernetesEnvoyFilterProxyMatch`

Match on properties of the proxy itself (istio proxy version and/or node metadata).

### spec.configPatches[].match.proxy.proxyVersion

`string` · optional (explicit presence)

A golang (RE2) regular expression selecting proxies by Istio proxy version. The version
is read from the proxy's `ISTIO_VERSION` node metadata.

### spec.configPatches[].match.proxy.metadata

`map<string, string>`

Match on node metadata supplied by the proxy when connecting to istiod. Only string
key-value pairs are processed; all specified keys must be present and match exactly.

### spec.configPatches[].match.listener

`KubernetesEnvoyFilterListenerMatch`

Match on Envoy listener attributes.

### spec.configPatches[].match.listener.portNumber

`uint32` · optional (explicit presence)

The service/gateway port traffic is sent to/received on. If unset, matches all listeners.
Only service ports (not pod ports) should be used to match inbound listeners.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.configPatches[].match.listener.filterChain

`KubernetesEnvoyFilterFilterChainMatch`

Match a specific filter chain in the listener. If set, the patch is applied only to that
filter chain (and a specific filter within it, if specified), not to other filter chains.

### spec.configPatches[].match.listener.filterChain.name

`string` · optional (explicit presence)

The name assigned to the filter chain.

### spec.configPatches[].match.listener.filterChain.sni

`string` · optional (explicit presence)

The SNI value used by the filter chain's match condition. Evaluates to false if the
filter chain has no SNI match.

### spec.configPatches[].match.listener.filterChain.transportProtocol

`string` · optional (explicit presence)

Applies only to SIDECAR_INBOUND. A transport protocol to match against a new connection
(detected by the tls_inspector listener filter). Accepted values: `raw_buffer` (default,
no protocol detected) and `tls`.

### spec.configPatches[].match.listener.filterChain.applicationProtocols

`string` · optional (explicit presence)

Applies only to sidecars. A comma-separated set of application protocols to match against
a new connection (detected by listener filters such as http_inspector). Accepted values
include: h2, http/1.1, http/1.0.

### spec.configPatches[].match.listener.filterChain.filter

`KubernetesEnvoyFilterFilterMatch`

The specific filter within the chain to apply the patch to. Set name to
`envoy.filters.network.http_connection_manager` to add/patch the HTTP connection manager.

### spec.configPatches[].match.listener.filterChain.filter.name

`string` · optional (explicit presence)

The filter name to match on. For standard Envoy filters, use the canonical filter name.

### spec.configPatches[].match.listener.filterChain.filter.subFilter

`KubernetesEnvoyFilterSubFilterMatch`

The next-level filter within this filter to match (typically an HTTP filter inside the
HTTP connection manager network filter).

### spec.configPatches[].match.listener.filterChain.filter.subFilter.name

`string` · optional (explicit presence)

The filter name to match on.

### spec.configPatches[].match.listener.filterChain.destinationPort

`uint32` · optional (explicit presence)

The destination_port used by the filter chain's match condition. Evaluates to false if
the filter chain has no destination_port match.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.configPatches[].match.listener.listenerFilter

`string` · optional (explicit presence)

Match a specific listener filter. If set, the patch is applied to that listener filter.

### spec.configPatches[].match.listener.name

`string` · optional (explicit presence)

Match a specific listener by name. istiod-generated listeners are typically named IP:Port.

### spec.configPatches[].match.routeConfiguration

`KubernetesEnvoyFilterRouteConfigurationMatch`

Match on Envoy HTTP route configuration attributes.

### spec.configPatches[].match.routeConfiguration.portNumber

`uint32` · optional (explicit presence)

The service port (or, for GATEWAY context, the gateway server port) for which this route
configuration was generated. If omitted, applies to route configurations for all ports.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.configPatches[].match.routeConfiguration.portName

`string` · optional (explicit presence)

Applicable only to GATEWAY context: the gateway server port name for which this route
configuration was generated.

### spec.configPatches[].match.routeConfiguration.gateway

`string` · optional (explicit presence)

Applicable only to GATEWAY context: the Istio Gateway config's `namespace/name` for which
this route configuration was generated. Use together with port_number/port_name to select
the Envoy route configuration for a specific HTTPS server in a gateway.

### spec.configPatches[].match.routeConfiguration.vhost

`KubernetesEnvoyFilterVirtualHostMatch`

Match a specific virtual host inside the route configuration and apply the patch to it.

### spec.configPatches[].match.routeConfiguration.vhost.name

`string` · optional (explicit presence)

The virtual host name. Istio names virtual hosts as `host:port`, where host typically
corresponds to a VirtualService host or a registry service hostname.

### spec.configPatches[].match.routeConfiguration.vhost.domainName

`string` · optional (explicit presence)

Match a domain name served by the virtual host. If this domain is in the virtual host's
domain list, the patch is applied.

### spec.configPatches[].match.routeConfiguration.vhost.route

`KubernetesEnvoyFilterRouteMatch`

Match a specific route within the virtual host.

### spec.configPatches[].match.routeConfiguration.vhost.route.name

`string` · optional (explicit presence)

The route name. Default routes are named `default`; routes generated from a VirtualService
carry the name used in the VirtualService's HTTP routes.

### spec.configPatches[].match.routeConfiguration.vhost.route.action

`string` · optional (explicit presence)

Match a route with a specific action type. Unset matches ANY action upstream. external
standard exception — matches the Istio RouteMatch.Action enum.

- rule: {"string":{"in":["ANY","ROUTE","REDIRECT","DIRECT_RESPONSE"]}}

### spec.configPatches[].match.routeConfiguration.name

`string` · optional (explicit presence)

The route configuration name to match (e.g. the internally-generated `http_proxy` route
configuration for all sidecars).

### spec.configPatches[].match.cluster

`KubernetesEnvoyFilterClusterMatch`

Match on Envoy cluster attributes.

### spec.configPatches[].match.cluster.portNumber

`uint32` · optional (explicit presence)

The service port for which this cluster was generated. If omitted, applies to clusters
for any port. For an inbound cluster, this is the service target port.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.configPatches[].match.cluster.service

`string` · optional (explicit presence)

The fully-qualified service name for this cluster. If omitted, applies to clusters for
any service. For service-entry services, this equals the service entry's host. Ignored
for inbound clusters.

### spec.configPatches[].match.cluster.subset

`string` · optional (explicit presence)

The subset associated with the service. If omitted, applies to any subset of a service.

### spec.configPatches[].match.cluster.name

`string` · optional (explicit presence)

The exact name of the cluster to match (e.g. the internally-generated `Passthrough`
cluster). To match by name only, leave all other fields empty.

### spec.configPatches[].match.waypoint

`KubernetesEnvoyFilterWaypointMatch`

Match on ambient waypoint-proxy attributes (a filter within the waypoint's
filter chain, a service port, or a named route).

### spec.configPatches[].match.waypoint.filter

`KubernetesEnvoyFilterWaypointFilterMatch`

The name of a specific filter (optionally a sub-filter within it) to apply the
patch to.

### spec.configPatches[].match.waypoint.filter.name

`string` · optional (explicit presence)

The filter name to match on.

### spec.configPatches[].match.waypoint.filter.subFilter

`KubernetesEnvoyFilterWaypointSubFilterMatch`

The next-level filter within this filter to match on.

### spec.configPatches[].match.waypoint.filter.subFilter.name

`string` · optional (explicit presence)

The filter name to match on.

### spec.configPatches[].match.waypoint.portNumber

`uint32` · optional (explicit presence)

The service port to match on (1-65535).

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.configPatches[].match.waypoint.route

`KubernetesEnvoyFilterWaypointRouteMatch`

Match a specific route by name (the default generated Route objects are named
`default`).

### spec.configPatches[].match.waypoint.route.name

`string` · optional (explicit presence)

The route name to match on (default generated Route objects are named `default`).

### spec.configPatches[].patch

`KubernetesEnvoyFilterPatch`

The patch to apply, along with the operation.

### spec.configPatches[].patch.operation

`string` · optional (explicit presence)

How the patch is applied to the selected configuration. external standard exception —
matches the Istio EnvoyFilter.Patch.Operation enum.
  MERGE         — proto-merge the value into the generated config.
  ADD           — add to an existing list (listeners/clusters/vhosts/filters); ignored for
                  ROUTE_CONFIGURATION and HTTP_ROUTE.
  REMOVE        — remove the selected object (no value required); ignored for
                  ROUTE_CONFIGURATION and HTTP_ROUTE.
  INSERT_BEFORE — insert before the selected filter/route.
  INSERT_AFTER  — insert after the selected filter/route.
  INSERT_FIRST  — insert at the front of the list.
  REPLACE       — replace a named filter's contents; valid only for HTTP_FILTER and
                  NETWORK_FILTER.

- rule: {"string":{"in":["MERGE","ADD","REMOVE","INSERT_BEFORE","INSERT_AFTER","INSERT_FIRST","REPLACE"]}}

### spec.configPatches[].patch.value

`object`

The free-form JSON config of the object being patched, merged using proto-merge semantics
with the existing Envoy config. This is the irreducible expert-only escape hatch: it is an
arbitrary google.protobuf.Struct (the upstream CRD marks `patch.value` as
preserveUnknownFields), so istiod applies no schema validation here. A malformed value can
break the patched workload's traffic. Not required (REMOVE needs no value).

### spec.configPatches[].patch.filterClass

`string` · optional (explicit presence)

The filter insertion point relative to istiod's implicitly-inserted filters, used with the
ADD operation. Preferred over INSERT_* operations, which rely on potentially-unstable
filter names. Unset lets the control plane decide. external standard exception — matches
the Istio EnvoyFilter.Patch.FilterClass enum.
  AUTHN — insert after Istio authentication filters.
  AUTHZ — insert after Istio authorization filters.
  STATS — insert before Istio stats filters.

- rule: {"string":{"in":["AUTHN","AUTHZ","STATS"]}}

### spec.priority

`int32` · optional (explicit presence)

Defines the order in which patch sets are applied within a context. Patch sets in the
mesh root namespace are applied before those in the workload namespace; within a patch
set, patches apply in `config_patches` list order. The default is 0; negative values are
processed before the default, positive values after. Sort key (ascending): priority,
creation time, fully-qualified resource name. Unset leaves the upstream default (0).

### spec.targetRefs

`[]KubernetesIstioApiPolicyTargetReference`

Attaches these patches to specific resources instead of selecting workloads by label.
Upstream supports: `kind: Gateway`/`GatewayClass` (group gateway.networking.k8s.io),
`kind: Service` (core group, waypoints only), and `kind: ServiceEntry` (group
networking.istio.io). Mutually exclusive with `workload_selector` (enforced above).
Upstream allows at most 16. Waypoint proxies REQUIRE target_refs; selector-based
policies are ignored for waypoints.

INFRA-CHART COMPOSABILITY: each target_ref's `name` is a foreign key that
defaults to a KubernetesGateway (its status.outputs.gateway_name): wiring it with
valueFrom orders this EnvoyFilter after the gateway it patches. A literal `value:`
covers a Service, a ServiceEntry, or anything created outside Planton; istiod
resolves group/kind/name at runtime, so order this EnvoyFilter after such a target
with metadata.relationships (`uses` -> KubernetesService / KubernetesServiceEntry).
See the kind's "Composing in Infra Charts" docs.

- rule: {"repeated":{"maxItems":"16"}}

### spec.targetRefs[].group

`string`

Group of the target resource. Empty for the core API group (Services). Faithful
to the upstream pattern (empty, or a DNS-1123 subdomain).

- rule: {"string":{"maxLen":"253","pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"}}

### spec.targetRefs[].kind

`string` · required

Kind of the target resource (e.g. Gateway, Service, ServiceEntry). Required.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.targetRefs[].name

`string | valueFrom` · required

Name of the target resource. Required. Defaults to a KubernetesGateway foreign
key (the policy attaches to a Gateway API Gateway) — in an infra chart, wire it
with valueFrom so the policy deploys after its gateway. For other target kinds,
pass the literal name with `value:`. Upstream bounds the name at 253 characters;
the API server enforces that at apply (a StringValueOrRef carries no bound).

- references: KubernetesGateway (`status.outputs.gateway_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_name}} -- a bare string does not parse

### spec.targetRefs[].namespace

`string`

Namespace of the target resource. Cross-namespace attachment is not supported
upstream in the 1.30 line, so this must be empty (the target is resolved in the
policy's own namespace). Mirrors the upstream XValidation rule
"cross namespace referencing is not currently supported".

- rule: cross-namespace target references are not supported; leave namespace empty

## Validation Rules

- `envoy_filter.workload_selector_xor_target_refs`: at most one of workload_selector or target_refs may be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesEnvoyFilter, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.envoy_filter_name` | `string` | Name of the created EnvoyFilter (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the EnvoyFilter was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.targetRefs[].name` | KubernetesGateway | `status.outputs.gateway_name` |

## See Also

- [Overview](../README.md)
