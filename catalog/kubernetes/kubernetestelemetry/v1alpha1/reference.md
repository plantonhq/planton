# KubernetesTelemetry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesTelemetrySpec defines an Istio Telemetry resource: a namespaced
configuration of how telemetry (traces, metrics, and access logs) is generated
for the workloads it selects.

100% fidelity with the upstream istio.io/api Telemetry
(telemetry/v1alpha1/telemetry.proto, served as telemetry.istio.io/v1), pinned to
the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly after the
Planton namespaced envelope (namespace); there is no nested
`telemetry` sub-message.

Scope semantics (upstream): a Telemetry resource with no selector and no
target_refs applies to all workloads in its namespace (or, in the mesh root
namespace, the whole mesh). At most one of `selector` and `target_refs` may be set
(enforced below). Gateways and waypoints are targeted via `target_refs`; waypoint
proxies require `target_refs` (label `selector` policies are ignored by waypoints).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTelemetry
metadata:
  name: test-telemetry
spec:
  namespace:
    value: test-namespace
  selector:
    match_labels:
      app: reviews
  tracing:
    - match:
        mode: CLIENT
      providers:
        - name: zipkin
      random_sampling_percentage: 10
      disable_span_reporting: false
      enable_istio_tags: true
      custom_tags:
        my_literal_tag:
          literal:
            value: foo
        my_env_tag:
          environment:
            name: POD_NAME
            default_value: unknown
        my_header_tag:
          header:
            name: x-request-id
  metrics:
    - providers:
        - name: prometheus
      reporting_interval: 5s
      overrides:
        - match:
            metric: REQUEST_COUNT
            mode: SERVER
          tag_overrides:
            request_method:
              operation: UPSERT
              value: request.method
            response_code:
              operation: REMOVE
        - match:
            custom_metric: my_custom_metric
          disabled: true
  access_logging:
    - match:
        mode: SERVER
      providers:
        - name: envoy
      filter:
        expression: "response.code >= 400"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.selector` | `KubernetesIstioApiWorkloadSelector` |  |  |  |
| `spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.targetRefs` | `[]KubernetesIstioApiPolicyTargetReference` |  |  |  |
| `spec.targetRefs[].group` | `string` |  |  |  |
| `spec.targetRefs[].kind` | `string` | yes |  |  |
| `spec.targetRefs[].name` | `string \| valueFrom` | yes |  | KubernetesGateway (`status.outputs.gateway_name`) |
| `spec.targetRefs[].namespace` | `string` |  |  |  |
| `spec.tracing` | `[]KubernetesTelemetryTracing` |  |  |  |
| `spec.tracing[].match` | `KubernetesTelemetryTracingSelector` |  |  |  |
| `spec.tracing[].match.mode` | `string` |  |  |  |
| `spec.tracing[].providers` | `[]KubernetesTelemetryProviderRef` |  |  |  |
| `spec.tracing[].providers[].name` | `string` | yes |  |  |
| `spec.tracing[].randomSamplingPercentage` | `double` |  |  |  |
| `spec.tracing[].disableSpanReporting` | `bool` |  |  |  |
| `spec.tracing[].customTags` | `map<string, KubernetesTelemetryCustomTag>` |  |  |  |
| `spec.tracing[].customTags.*.literal` | `KubernetesTelemetryCustomTagLiteral` |  |  |  |
| `spec.tracing[].customTags.*.literal.value` | `string` | yes |  |  |
| `spec.tracing[].customTags.*.environment` | `KubernetesTelemetryCustomTagEnvironment` |  |  |  |
| `spec.tracing[].customTags.*.environment.name` | `string` | yes |  |  |
| `spec.tracing[].customTags.*.environment.defaultValue` | `string` |  |  |  |
| `spec.tracing[].customTags.*.header` | `KubernetesTelemetryCustomTagRequestHeader` |  |  |  |
| `spec.tracing[].customTags.*.header.name` | `string` | yes |  |  |
| `spec.tracing[].customTags.*.header.defaultValue` | `string` |  |  |  |
| `spec.tracing[].customTags.*.formatter` | `KubernetesTelemetryCustomTagFormatter` |  |  |  |
| `spec.tracing[].customTags.*.formatter.value` | `string` | yes |  |  |
| `spec.tracing[].enableIstioTags` | `bool` |  |  |  |
| `spec.tracing[].useRequestIdForTraceSampling` | `bool` |  |  |  |
| `spec.tracing[].disableContextPropagation` | `bool` |  |  |  |
| `spec.metrics` | `[]KubernetesTelemetryMetrics` |  |  |  |
| `spec.metrics[].providers` | `[]KubernetesTelemetryProviderRef` |  |  |  |
| `spec.metrics[].providers[].name` | `string` | yes |  |  |
| `spec.metrics[].overrides` | `[]KubernetesTelemetryMetricsOverride` |  |  |  |
| `spec.metrics[].overrides[].match` | `KubernetesTelemetryMetricSelector` |  |  |  |
| `spec.metrics[].overrides[].match.metric` | `string` |  |  |  |
| `spec.metrics[].overrides[].match.customMetric` | `string` | yes |  |  |
| `spec.metrics[].overrides[].match.mode` | `string` |  |  |  |
| `spec.metrics[].overrides[].disabled` | `bool` |  |  |  |
| `spec.metrics[].overrides[].tagOverrides` | `map<string, KubernetesTelemetryTagOverride>` |  |  |  |
| `spec.metrics[].overrides[].tagOverrides.*.operation` | `string` |  |  |  |
| `spec.metrics[].overrides[].tagOverrides.*.value` | `string` |  |  |  |
| `spec.metrics[].reportingInterval` | `string` |  |  |  |
| `spec.accessLogging` | `[]KubernetesTelemetryAccessLogging` |  |  |  |
| `spec.accessLogging[].match` | `KubernetesTelemetryAccessLoggingSelector` |  |  |  |
| `spec.accessLogging[].match.mode` | `string` |  |  |  |
| `spec.accessLogging[].providers` | `[]KubernetesTelemetryProviderRef` |  |  |  |
| `spec.accessLogging[].providers[].name` | `string` | yes |  |  |
| `spec.accessLogging[].disabled` | `bool` |  |  |  |
| `spec.accessLogging[].filter` | `KubernetesTelemetryAccessLoggingFilter` |  |  |  |
| `spec.accessLogging[].filter.expression` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the Telemetry resource is created. The configuration's scope
is this namespace (or mesh-wide if this is the Istio root namespace and no
selector or target_refs is set).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.selector

`KubernetesIstioApiWorkloadSelector`

Selects the workloads (by pod/VM label) this configuration applies to. When
omitted (and target_refs is also omitted), it applies to every workload in its
namespace (or, in the root namespace, the whole mesh). At most one of `selector`
and `target_refs` may be set (enforced above).

INFRA-CHART COMPOSABILITY: selector is a PLAIN label match, not an Planton
foreign key (StringValueOrRef). It is matched at runtime by istiod against pod
labels and creates NO automatic DAG edge to any workload resource. To order this
configuration after the workload it observes in an infra chart, express the
dependency via metadata.relationships. See the component's "Composing in Infra
Charts" docs for the full pattern.

### spec.selector.matchLabels

`map<string, string>`

One or more labels indicating the set of pods/VMs the policy applies to.
Faithful to istio.io/api `istio.type.v1beta1.WorkloadSelector.match_labels`,
whose upstream CRD constraints are: max 4096 entries; each value <= 63 chars;
label keys must be non-empty; and wildcards ('*') are not permitted in keys or
values. The size/length bounds are expressed via the standard `map` rule; the
non-empty-key and no-wildcard constraints map to upstream's CEL XValidation
rules and are expressed here as field-level CEL.

- rule: label selector keys must not be empty
- rule: wildcard ('*') is not allowed in label selector keys
- rule: wildcard ('*') is not allowed in label selector values
- rule: {"map":{"maxPairs":"4096","values":{"string":{"maxLen":"63"}}}}

### spec.targetRefs

`[]KubernetesIstioApiPolicyTargetReference`

Attaches the configuration to specific resources (Gateway, GatewayClass,
Service, ServiceEntry) instead of selecting workloads by label. At most one of
`selector` and `target_refs` may be set (enforced above). Waypoint proxies
require this field. Upstream allows up to 16.

INFRA-CHART COMPOSABILITY: each target_ref's `name` is a foreign key that
defaults to a KubernetesGateway (its status.outputs.gateway_name): wiring it with
valueFrom orders this configuration after the gateway it observes. A literal `value:`
covers a Service, a ServiceEntry, or anything created outside Planton; istiod
resolves group/kind/name at runtime, so order this configuration after such a target
with metadata.relationships (`uses` -> KubernetesService / KubernetesServiceEntry).
See the component's "Composing in Infra Charts" docs.

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

### spec.tracing

`[]KubernetesTelemetryTracing`

Tracing configuration for the selected workloads. Each entry can enable/disable
span reporting, set the sampling rate, choose providers, and add custom span
tags. Multiple entries are merged by istiod with later entries overriding earlier
ones (and parent-scope resources).

### spec.tracing[].match

`KubernetesTelemetryTracingSelector`

Tailors the tracing configuration to specific traffic conditions (currently the
traffic direction relative to the proxied workload). When omitted, the
configuration applies to both directions.

### spec.tracing[].match.mode

`string` · optional (explicit presence)

The workload's role in the matched traffic. When unset, defaults to
CLIENT_AND_SERVER. One of:
  CLIENT_AND_SERVER — match whether the workload is the source or destination.
  CLIENT            — match when the workload is the traffic source (outbound).
  SERVER            — match when the workload is the traffic destination (inbound).
CLIENT_AND_SERVER/CLIENT/SERVER // external standard exception -- Istio WorkloadMode enum

- rule: {"string":{"in":["CLIENT_AND_SERVER","CLIENT","SERVER"]}}

### spec.tracing[].providers

`[]KubernetesTelemetryProviderRef`

Provider(s) to use for span reporting. If unset, the mesh default tracing
provider is used. NOTE: upstream currently honors only a single provider per
Tracing rule.

### spec.tracing[].providers[].name

`string` · required

Name of a telemetry provider in MeshConfig. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tracing[].randomSamplingPercentage

`double` · optional (explicit presence)

The percentage of requests selected for tracing when no prior sampling decision
exists, in 0.01% increments. Valid range [0.00-100.00]. Defaults to 0% upstream
when unset. Faithful to the upstream `google.protobuf.DoubleValue`
random_sampling_percentage (modeled as an optional scalar with the upstream
Minimum/Maximum bounds).

- rule: {"double":{"lte":100,"gte":0}}

### spec.tracing[].disableSpanReporting

`bool` · optional (explicit presence)

When true, no spans are reported for the selected workloads. Does NOT affect
context propagation or trace sampling. Faithful to the upstream
`google.protobuf.BoolValue` (optional scalar).

### spec.tracing[].customTags

`map<string, KubernetesTelemetryCustomTag>`

Custom tags to add to generated trace spans, keyed by tag name. Each value
supplies the tag from exactly one source: a literal value, an environment
variable, or a request header. NOTE (upstream): when specified, custom_tags
fully replaces any values inherited from parent configuration.

- rule: at most one of literal, environment, header, or formatter may be set

### spec.tracing[].customTags.*.literal

`KubernetesTelemetryCustomTagLiteral`

Adds the same hard-coded value to each span.

### spec.tracing[].customTags.*.literal.value

`string` · required

The tag value to use. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tracing[].customTags.*.environment

`KubernetesTelemetryCustomTagEnvironment`

Adds the value of an environment variable (read by the sidecar) to each span.

### spec.tracing[].customTags.*.environment.name

`string` · required

Name of the environment variable from which to read the tag value. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tracing[].customTags.*.environment.defaultValue

`string`

Fallback value used when the environment variable is not found.

### spec.tracing[].customTags.*.header

`KubernetesTelemetryCustomTagRequestHeader`

Adds the value of a request header to each span.

### spec.tracing[].customTags.*.header.name

`string` · required

Name of the request header from which to read the tag value. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tracing[].customTags.*.header.defaultValue

`string`

Fallback value used when the header is not found.

### spec.tracing[].customTags.*.formatter

`KubernetesTelemetryCustomTagFormatter`

Adds a value computed by an access-log substitution formatter expression
(e.g. `%PROTOCOL%`, `%REQ(:path)%` — the same command operators HTTP access
logging uses), evaluated per request.

### spec.tracing[].customTags.*.formatter.value

`string` · required

The formatter expression to evaluate (same command operators as HTTP access
logging, e.g. `%PROTOCOL%`, `%REQ(:authority)%`). Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tracing[].enableIstioTags

`bool` · optional (explicit presence)

When true (the default), Envoy includes Istio-specific tags in generated trace
spans. Faithful to the upstream `google.protobuf.BoolValue` (optional scalar).

### spec.tracing[].useRequestIdForTraceSampling

`bool` · optional (explicit presence)

Controls whether Envoy bases its sampling decision on the Request ID it
generated. Set false to keep traces intact when the upstream Request ID is not
in Envoy's format. Advanced, hidden-from-docs upstream knob (extended release
channel); carried for full fidelity since the CRD accepts it. Faithful to the
upstream `google.protobuf.BoolValue` (optional scalar).

### spec.tracing[].disableContextPropagation

`bool` · optional (explicit presence)

When true, trace context headers (`traceparent`/`tracestate` for W3C,
`X-B3-*` for Zipkin) are NOT propagated in forwarded requests — spans still
report locally but downstream services start fresh traces. Defaults to false
(context propagates). Faithful to the upstream `google.protobuf.BoolValue`
(optional scalar).

### spec.metrics

`[]KubernetesTelemetryMetrics`

Metrics configuration for the selected workloads. Each entry can choose
providers and apply ordered overrides (enable/disable a metric, add/remove tag
dimensions) to the generated metrics.

### spec.metrics[].providers

`[]KubernetesTelemetryProviderRef`

Provider(s) this configuration applies to. If unset, the mesh default metrics
provider is used.

### spec.metrics[].providers[].name

`string` · required

Name of a telemetry provider in MeshConfig. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.metrics[].overrides

`[]KubernetesTelemetryMetricsOverride`

Ordered overrides applied to metrics generation. Overrides are applied in order
(and on top of inherited mesh/namespace/workload overrides), so list least
specific matches first.

### spec.metrics[].overrides[].match

`KubernetesTelemetryMetricSelector`

Scopes the override to specific metric(s) and workload mode(s). When omitted, the
override applies to all metrics in both client and server modes.

- rule: at most one of metric or custom_metric may be set

### spec.metrics[].overrides[].match.metric

`string` · optional (explicit presence)

One of the well-known Istio standard metrics. One of:
  ALL_METRICS            — apply to all Istio default metrics.
  REQUEST_COUNT          — counter of HTTP/HTTP2/gRPC requests.
  REQUEST_DURATION       — histogram of request durations.
  REQUEST_SIZE           — histogram of request body sizes.
  RESPONSE_SIZE          — histogram of response body sizes.
  TCP_OPENED_CONNECTIONS — counter of opened TCP connections.
  TCP_CLOSED_CONNECTIONS — counter of closed TCP connections.
  TCP_SENT_BYTES         — counter of bytes sent over TCP.
  TCP_RECEIVED_BYTES     — counter of bytes received over TCP.
  GRPC_REQUEST_MESSAGES  — counter of gRPC messages sent from a client.
  GRPC_RESPONSE_MESSAGES — counter of gRPC messages sent from a server.
ALL_METRICS/... // external standard exception -- Istio MetricSelector.IstioMetric enum

- rule: {"string":{"in":["ALL_METRICS","REQUEST_COUNT","REQUEST_DURATION","REQUEST_SIZE","RESPONSE_SIZE","TCP_OPENED_CONNECTIONS","TCP_CLOSED_CONNECTIONS","TCP_SENT_BYTES","TCP_RECEIVED_BYTES","GRPC_REQUEST_MESSAGES","GRPC_RESPONSE_MESSAGES"]}}

### spec.metrics[].overrides[].match.customMetric

`string` · required · optional (explicit presence)

A free-form custom metric name. No validation of custom metrics is provided
upstream beyond non-emptiness.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].overrides[].match.mode

`string` · optional (explicit presence)

The workload mode the override applies to. When unset, defaults to
CLIENT_AND_SERVER. One of CLIENT_AND_SERVER, CLIENT, or SERVER.
CLIENT_AND_SERVER/CLIENT/SERVER // external standard exception -- Istio WorkloadMode enum

- rule: {"string":{"in":["CLIENT_AND_SERVER","CLIENT","SERVER"]}}

### spec.metrics[].overrides[].disabled

`bool` · optional (explicit presence)

When true, disables reporting for the matched metrics. Must be explicitly set to
false to re-enable a metric disabled by a parent configuration. Faithful to the
upstream `google.protobuf.BoolValue` (optional scalar).

### spec.metrics[].overrides[].tagOverrides

`map<string, KubernetesTelemetryTagOverride>`

Tag (dimension) operations to apply to the matched metrics, keyed by tag name.
Each value either upserts the tag with a CEL value expression or removes it.
WARNING (upstream): some providers may not support adding/removing tags.

- rule: value must be set when operation is UPSERT
- rule: value must not be set when operation is REMOVE

### spec.metrics[].overrides[].tagOverrides.*.operation

`string` · optional (explicit presence)

Whether to update/add the tag or remove it. When unset, upstream treats the
operation as UPSERT (the enum default). One of:
  UPSERT — insert or update the tag with `value` (which must then be set).
  REMOVE — exclude the tag from the metric (`value` must not be set).
UPSERT/REMOVE // external standard exception -- Istio TagOverride.Operation enum

- rule: {"string":{"in":["UPSERT","REMOVE"]}}

### spec.metrics[].overrides[].tagOverrides.*.value

`string` · optional (explicit presence)

A CEL expression over Istio/Envoy attributes whose result becomes the tag value.
Considered only when operation is UPSERT. Examples: `string(destination.port)`,
`request.host`.

### spec.metrics[].reportingInterval

`string` · optional (explicit presence)

The interval between TCP metrics reports. Defaults to `5s` upstream. Modeled as a
duration string (the catalog's convention for every upstream Duration field);
must be a valid duration of at least 1ms, mirroring the upstream CRD XValidation
`duration(self) >= duration('1ms')`.

- rule: reporting_interval must be a valid duration of at least 1ms (e.g. "5s")

### spec.accessLogging

`[]KubernetesTelemetryAccessLogging`

Access logging configuration for the selected workloads. Each entry can choose
providers, enable/disable logging, and attach a CEL filter.

### spec.accessLogging[].match

`KubernetesTelemetryAccessLoggingSelector`

Tailors the logging configuration to specific traffic conditions (currently the
traffic direction). When omitted, applies to both directions.

### spec.accessLogging[].match.mode

`string` · optional (explicit presence)

The workload's role in the matched traffic. When unset, defaults to
CLIENT_AND_SERVER. One of CLIENT_AND_SERVER, CLIENT, or SERVER.
CLIENT_AND_SERVER/CLIENT/SERVER // external standard exception -- Istio WorkloadMode enum

- rule: {"string":{"in":["CLIENT_AND_SERVER","CLIENT","SERVER"]}}

### spec.accessLogging[].providers

`[]KubernetesTelemetryProviderRef`

Provider(s) this configuration applies to. If unset, the mesh default logging
provider is used.

### spec.accessLogging[].providers[].name

`string` · required

Name of a telemetry provider in MeshConfig. Required.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.accessLogging[].disabled

`bool` · optional (explicit presence)

When true, no access logs are generated for the selected workloads (for the
selected providers). To re-enable logging disabled by a parent configuration,
set this explicitly to false. Faithful to the upstream `google.protobuf.BoolValue`
(optional scalar).

### spec.accessLogging[].filter

`KubernetesTelemetryAccessLoggingFilter`

A CEL filter selecting which requests/connections are logged. When omitted, all
are logged (subject to provider behavior).

### spec.accessLogging[].filter.expression

`string`

A CEL expression for selecting when requests/connections should be logged, e.g.
`response.code >= 400`. Free-form; upstream applies no schema validation.

## Validation Rules

- `telemetry.selector_xor_target_refs`: at most one of selector or target_refs may be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesTelemetry, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.telemetry_name` | `string` | Name of the created Telemetry resource (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the Telemetry resource was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.targetRefs[].name` | KubernetesGateway | `status.outputs.gateway_name` |

## See Also

- [Overview](../README.md)
