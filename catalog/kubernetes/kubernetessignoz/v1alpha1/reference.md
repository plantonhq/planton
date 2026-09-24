# KubernetesSignoz

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesSignozSpec** deploys SigNoz — the all-in-one open-source
observability platform (traces, metrics and logs in ONE UI, stored in
ClickHouse) — from the official `signoz` Helm chart
(https://charts.signoz.io).

WHAT GETS INSTALLED: one consolidated SigNoz server (UI, API, rule
evaluation and alerting in a single binary), one SigNoz OpenTelemetry
Collector (the ingestion gateway every application sends OTLP data
to), and a schema migrator. Nothing ClickHouse-related is installed:
the telemetry store is COMPOSED, never bundled (see THE DATABASE).

THE DATABASE: SigNoz stores every trace, metric and log in ClickHouse
— a required, separate component. Run a KubernetesClickHouse (with
its KubernetesAltinityOperator) and wire it through `clickhouse`; the
fields default-reference that kind's outputs, so the wiring is one
`valueFrom` per field. Any reachable ClickHouse also works with
literal values. The composition keeps every lifecycle honest: the
database (and your telemetry) outlives SigNoz reinstalls, upgrades
roll independently, and deep ClickHouse control (users, profiles,
quotas, keeper topology) lives on the component that owns it. KNOW
THIS (verified live): a chart-bundled database CANNOT uninstall
cleanly — the operator and its installation die in the same release
and the installation's finalizer deadlocks — which is why this
component composes instead of bundling.

SIGNOZ OR THE COMPOSED STACK? SigNoz is the "one product instead of
four" path: where KubernetesKubePrometheusStack + KubernetesGrafana +
KubernetesLoki + KubernetesTempo compose a best-of-breed stack, SigNoz
gives a single OpenTelemetry-native product with one UI. Both are
first-class; pick per team taste.

HOW DATA GETS IN: point any OTLP client at the exported
`otlp_grpc_endpoint` (4317) or `otlp_http_endpoint` (4318). For
cluster-wide telemetry shipping (node logs, kubelet metrics, k8s
events), deploy a KubernetesOtelCollector pointed at the same
endpoints.

SINGLE-INSTANCE TRUTH: the community SigNoz server keeps its own
state (users, dashboards, alert rules) in SQLite on a persistent
volume — a single-writer store, so the server runs exactly one
replica. This is upstream's community-edition posture (the
Postgres-backed HA store is enterprise-only); telemetry data itself
lives in ClickHouse, which scales independently. The ingestion
collector DOES scale horizontally (`otel_collector.replicas` and the
autoscaling arm).

EXPOSURE: everything stays ClusterIP. Expose the UI and the collector
through first-class kinds (KubernetesIngress, Gateway API kinds) over
the exported service handles — this component never creates ingress
objects.

The typed fields cover the chart's meaningful surface; `helm_values`
remains the escape hatch (merged last, Helm `-f` semantics, identical
on both engines) for the long tail — collector pipeline overrides,
migrator tuning — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: the composed ClickHouse connection
# over verified TLS with a Secret-referenced password, a fully tuned
# server (SMTP with a Secret-referenced password, external URL, an
# advanced env entry), the collector with autoscaling and every receiver
# toggle exercised (zipkin ON — the non-default), a private-mirror image
# registry with a pull secret, scheduling, and an escape-hatch entry — so
# the offline tofu plan and pulumi preview proofs cover the full typed
# surface. Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSignoz
metadata:
  name: signoz-hack
spec:
  namespace:
    value: signoz-hack
  createNamespace: true
  chartVersion: 0.133.0
  clickhouse:
    host:
      value: clickhouse-analytics.signoz-hack.svc.cluster.local
    clusterName:
      value: analytics
    tcpPort: 9440
    httpPort: 8443
    username: signoz
    passwordSecret:
      secretName:
        value: analytics-clickhouse-auth
      secretKey: signoz
    secure: true
    verify: true
  server:
    diskSize: 2Gi
    storageClass:
      value: gp3
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: "1"
        memory: 2Gi
    externalUrl: https://signoz.example.com
    smtp:
      address: smtp.example.com:587
      from: signoz@example.com
      username: smtp-user
      passwordSecret:
        name: smtp-auth
        key: password
    env:
      signoz_prometheus_active__query__tracker_enabled: "true"
  otelCollector:
    replicas: 2
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: "1"
        memory: 2Gi
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 8
      targetCpuUtilizationPercent: 60
      targetMemoryUtilizationPercent: 70
    zipkinReceiverEnabled: true
    lowCardinalityExceptionGrouping: true
  clusterName: hack-cluster
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    signoz:
      podAnnotations:
        team: observability
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.133.0` |  |
| `spec.clickhouse` | `KubernetesSignozClickHouse` | yes |  |  |
| `spec.clickhouse.host` | `string \| valueFrom` | yes |  | KubernetesClickHouse (`status.outputs.service_name`) |
| `spec.clickhouse.clusterName` | `string \| valueFrom` |  |  | KubernetesClickHouse (`status.outputs.cluster_name`) |
| `spec.clickhouse.tcpPort` | `int32` |  | `9000` |  |
| `spec.clickhouse.httpPort` | `int32` |  | `8123` |  |
| `spec.clickhouse.username` | `string` | yes |  |  |
| `spec.clickhouse.passwordSecret` | `KubernetesSignozClickHousePassword` | yes |  |  |
| `spec.clickhouse.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesClickHouse (`status.outputs.auth_secret_name`) |
| `spec.clickhouse.passwordSecret.secretKey` | `string` | yes |  |  |
| `spec.clickhouse.secure` | `bool` |  |  |  |
| `spec.clickhouse.verify` | `bool` |  |  |  |
| `spec.server` | `KubernetesSignozServer` |  |  |  |
| `spec.server.diskSize` | `string` |  | `1Gi` |  |
| `spec.server.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.server.resources` | `ContainerResources` |  |  |  |
| `spec.server.resources.limits` | `CpuMemory` |  |  |  |
| `spec.server.resources.limits.cpu` | `string` |  |  |  |
| `spec.server.resources.limits.memory` | `string` |  |  |  |
| `spec.server.resources.requests` | `CpuMemory` |  |  |  |
| `spec.server.resources.requests.cpu` | `string` |  |  |  |
| `spec.server.resources.requests.memory` | `string` |  |  |  |
| `spec.server.externalUrl` | `string` |  |  |  |
| `spec.server.smtp` | `KubernetesSignozSmtp` |  |  |  |
| `spec.server.smtp.address` | `string` | yes |  |  |
| `spec.server.smtp.from` | `string` | yes |  |  |
| `spec.server.smtp.username` | `string` |  |  |  |
| `spec.server.smtp.passwordSecret` | `KubernetesSignozSecretKeyRef` |  |  |  |
| `spec.server.smtp.passwordSecret.name` | `string` | yes |  |  |
| `spec.server.smtp.passwordSecret.key` | `string` | yes |  |  |
| `spec.server.smtp.tlsEnabled` | `bool` |  |  |  |
| `spec.server.env` | `map<string, string>` |  |  |  |
| `spec.otelCollector` | `KubernetesSignozOtelCollector` |  |  |  |
| `spec.otelCollector.replicas` | `int32` |  | `1` |  |
| `spec.otelCollector.resources` | `ContainerResources` |  |  |  |
| `spec.otelCollector.resources.limits` | `CpuMemory` |  |  |  |
| `spec.otelCollector.resources.limits.cpu` | `string` |  |  |  |
| `spec.otelCollector.resources.limits.memory` | `string` |  |  |  |
| `spec.otelCollector.resources.requests` | `CpuMemory` |  |  |  |
| `spec.otelCollector.resources.requests.cpu` | `string` |  |  |  |
| `spec.otelCollector.resources.requests.memory` | `string` |  |  |  |
| `spec.otelCollector.autoscaling` | `KubernetesSignozOtelCollectorAutoscaling` |  |  |  |
| `spec.otelCollector.autoscaling.enabled` | `bool` |  |  |  |
| `spec.otelCollector.autoscaling.minReplicas` | `int32` |  | `1` |  |
| `spec.otelCollector.autoscaling.maxReplicas` | `int32` |  | `11` |  |
| `spec.otelCollector.autoscaling.targetCpuUtilizationPercent` | `int32` |  |  |  |
| `spec.otelCollector.autoscaling.targetMemoryUtilizationPercent` | `int32` |  |  |  |
| `spec.otelCollector.jaegerReceiverEnabled` | `bool` |  | `true` |  |
| `spec.otelCollector.zipkinReceiverEnabled` | `bool` |  |  |  |
| `spec.otelCollector.httpLogsReceiversEnabled` | `bool` |  | `true` |  |
| `spec.otelCollector.lowCardinalityExceptionGrouping` | `bool` |  |  |  |
| `spec.clusterName` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesSignozScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "0.133.0" — the chart version
tracks the SigNoz application version in lockstep). Versions must
exist as SERVED charts in the repository index
(https://charts.signoz.io).

- default: `0.133.0`

### spec.clickhouse

`KubernetesSignozClickHouse` · required

The ClickHouse SigNoz stores all telemetry in — required, composed,
never bundled. Fields default-reference a KubernetesClickHouse
resource's outputs for one-line composition; any reachable
ClickHouse ≥ the version the chart ships works with literal values.

- rule: {"required":true}
- rule: verify: true is meaningless without secure: true — certificate verification only applies to a TLS connection

### spec.clickhouse.host

`string | valueFrom` · required

ClickHouse host — a Service name (same namespace) or a full FQDN
(cross-namespace, e.g.
"clickhouse-analytics.data.svc.cluster.local"). Accepts a literal
or a reference to a KubernetesClickHouse resource (its client
Service name). Ports are declared separately — do not include one
here.

- references: KubernetesClickHouse (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.clickhouse.clusterName

`string | valueFrom`

Logical ClickHouse cluster name — SigNoz runs its distributed DDL
`ON CLUSTER` against this. Accepts a literal or a reference to a
KubernetesClickHouse resource (its cluster name). Empty =
"cluster" (the chart default; correct only if your cluster is
actually named that).

- references: KubernetesClickHouse (`status.outputs.cluster_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.cluster_name}} -- a bare string does not parse

### spec.clickhouse.tcpPort

`int32` · optional (explicit presence)

Native-protocol (TCP) port. Empty = 9000.

- default: `9000`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.clickhouse.httpPort

`int32` · optional (explicit presence)

HTTP interface port. Empty = 8123.

- default: `8123`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.clickhouse.username

`string` · required

ClickHouse username. SigNoz creates and migrates its own databases
(signoz_traces, signoz_metrics, signoz_logs, signoz_meter,
signoz_metadata) with `ON CLUSTER` DDL. On a KubernetesClickHouse,
the simplest honest posture is a user declared with NO grants —
verified live: a no-grants user carries ClickHouse's unrestricted
config-user access, which covers everything the migrator does. If
you constrain the user instead, the grant set must include
"GRANT CLUSTER ON *.*" (distributed DDL is rejected without it)
plus CREATE/DROP/INSERT/SELECT on the signoz_* databases. AND
declare the user's `networks` explicitly (e.g. "0.0.0.0/0" +
"::/0") — verified live: a user declared without networks is
fenced to the ClickHouse pods and localhost by the operator, and
SigNoz's pods are rejected with what reads as a password failure.

- rule: {"required":true}

### spec.clickhouse.passwordSecret

`KubernetesSignozClickHousePassword` · required

The user's password, read from an existing Secret (the chart wires
it as a secretKeyRef — it never lands in rendered values).

- rule: {"required":true}

### spec.clickhouse.passwordSecret.secretName

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesClickHouse resource (its module-owned auth Secret, which
carries one key per declared username). KNOW THIS (a Kubernetes
constraint, not a chart one): a secretKeyRef can only read Secrets
in the workload's OWN namespace — the Secret must exist in the
SigNoz namespace. Co-locate SigNoz with its ClickHouse (the
default composition), or replicate the Secret into the SigNoz
namespace when they live apart.

- references: KubernetesClickHouse (`status.outputs.auth_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClickHouse, name: <that resource's name>, fieldPath: status.outputs.auth_secret_name}} -- a bare string does not parse

### spec.clickhouse.passwordSecret.secretKey

`string` · required

Key within the Secret holding the password. On a
KubernetesClickHouse auth Secret this is the username.

- rule: {"required":true}

### spec.clickhouse.secure

`bool`

Use a TLS connection to ClickHouse.

### spec.clickhouse.verify

`bool`

Verify the TLS certificate. Only meaningful with `secure: true`.

### spec.server

`KubernetesSignozServer`

The SigNoz server — UI, API, rule evaluation and alerting in one
binary. Empty = a 1Gi state volume and the chart's default sizing.

### spec.server.diskSize

`string` · optional (explicit presence)

Size of the persistent volume holding the server's own state —
users, dashboards, alert rules (SQLite). 1Gi is plenty; telemetry
data lives in ClickHouse, not here.

- default: `1Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.server.storageClass

`string | valueFrom`

Storage class for the state volume. Empty = the cluster's default
class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.server.resources

`ContainerResources`

CPU and memory for the server container. Empty = the chart's
defaults (requests only).

### spec.server.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.server.resources.limits.cpu

`string`

### spec.server.resources.limits.memory

`string`

### spec.server.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.server.resources.requests.cpu

`string`

### spec.server.resources.requests.memory

`string`

### spec.server.externalUrl

`string`

The URL under which the SigNoz UI is externally reachable (e.g.
"https://signoz.example.com"). Used to build the links inside
alert notifications and invitation emails — without it they point
at localhost. Set it to whatever hostname you expose the UI on.

### spec.server.smtp

`KubernetesSignozSmtp`

SMTP for alert emails and user invitations. Empty = emailing off.

- rule: an SMTP password requires a username — declare both or neither

### spec.server.smtp.address

`string` · required

SMTP server address as host:port (e.g. "smtp.example.com:587").

- rule: address must be host:port, e.g. 'smtp.example.com:587'
- rule: {"required":true}

### spec.server.smtp.from

`string` · required

The From address for outgoing mail (e.g. "signoz@example.com").

- rule: {"required":true}

### spec.server.smtp.username

`string`

SMTP auth username. Empty = unauthenticated SMTP (an internal
relay).

### spec.server.smtp.passwordSecret

`KubernetesSignozSecretKeyRef`

SMTP auth password, read from an existing Secret (wired as a
secretKeyRef — never rendered into values or config).

### spec.server.smtp.passwordSecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.server.smtp.passwordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.server.smtp.tlsEnabled

`bool`

Use implicit TLS for the SMTP connection. Leave false for
STARTTLS-upgraded ports (587) — the common posture; set true only
for implicit-TLS ports (465).

### spec.server.env

`map<string, string>`

Advanced SigNoz configuration as environment variables. Keys
follow SigNoz's own derivation: `signoz_<section>_<key>` with
embedded underscores doubled (e.g.
`signoz_alertmanager_signoz_external__url`) — the full catalog is
upstream's conf/example.yaml. The typed fields above win on
conflict. Never put secret values here — this map renders as
plain environment variables; secret material belongs in the typed
secret-reference fields.

### spec.otelCollector

`KubernetesSignozOtelCollector`

The SigNoz OpenTelemetry Collector — the ingestion gateway. Empty =
one replica with the chart's default receivers (OTLP gRPC + HTTP,
Jaeger, HTTP log endpoints).

### spec.otelCollector.replicas

`int32` · optional (explicit presence)

Number of collector replicas. Scale this (or enable autoscaling)
as ingest volume grows — the collector is stateless.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.otelCollector.resources

`ContainerResources`

CPU and memory for each collector container. Empty = the chart's
defaults. The collector's memory limiter derives from the limit —
set limits in production.

### spec.otelCollector.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.otelCollector.resources.limits.cpu

`string`

### spec.otelCollector.resources.limits.memory

`string`

### spec.otelCollector.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.otelCollector.resources.requests.cpu

`string`

### spec.otelCollector.resources.requests.memory

`string`

### spec.otelCollector.autoscaling

`KubernetesSignozOtelCollectorAutoscaling`

Horizontal autoscaling for the collector (a standard HPA on
CPU/memory utilization). When enabled, `replicas` is ignored — the
HPA owns the count.

- rule: max_replicas must be greater than or equal to min_replicas

### spec.otelCollector.autoscaling.enabled

`bool`

Enable the HPA.

### spec.otelCollector.autoscaling.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.otelCollector.autoscaling.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas.

- default: `11`
- rule: {"int32":{"gte":1}}

### spec.otelCollector.autoscaling.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target CPU utilization percentage. Empty = 50 (the chart default).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.otelCollector.autoscaling.targetMemoryUtilizationPercent

`int32` · optional (explicit presence)

Target memory utilization percentage. Empty = 50 (the chart
default).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.otelCollector.jaegerReceiverEnabled

`bool` · optional (explicit presence)

Accept spans in Jaeger's legacy protocols (thrift-HTTP 14268 +
gRPC 14250) alongside OTLP. On by default (the chart's grain);
disable when everything speaks OTLP.

- default: `true`

### spec.otelCollector.zipkinReceiverEnabled

`bool`

Accept spans in the Zipkin protocol (9411). Off by default.

### spec.otelCollector.httpLogsReceiversEnabled

`bool` · optional (explicit presence)

Accept logs over plain HTTP (the JSON endpoint 8082 and the
Heroku-drain endpoint 8081). On by default (the chart's grain);
disable to make OTLP the only way in.

- default: `true`

### spec.otelCollector.lowCardinalityExceptionGrouping

`bool`

Group exceptions by name only, dropping per-stack-trace
cardinality. A cardinality-vs-fidelity trade for very
exception-noisy estates; leave false until the exceptions page
itself becomes the cost.

### spec.clusterName

`string`

Kubernetes cluster name attached to telemetry as a resource
attribute (the chart's `global.clusterName`). Set it when multiple
clusters report into shared dashboards — it is how you tell their
data apart.

### spec.imageRegistry

`string`

Registry that replaces the registry part of EVERY image this
component pulls (the SigNoz server, the collector, the schema
migrator) — the air-gap/private-mirror path (the chart's
`global.imageRegistry`). Empty = each image's upstream registry
(docker.io).

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to every workload
(chart `global.imagePullSecrets`). The Secrets must already exist
in the namespace.

### spec.scheduling

`KubernetesSignozScheduling`

Scheduling applied to the SigNoz server, the collector and the
schema migrator.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

Priority class name for the pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (collector pipeline config, migrator tuning) — never the
substitute for them. Do not put secrets here; passwords belong in
the typed secret-reference fields, which keep them out of the
rendered config.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSignoz, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace SigNoz runs in. |
| `status.outputs.service` | `string` | name of the SigNoz server Service (UI + API), e.g. signoz-main. |
| `status.outputs.kube_endpoint` | `string` | in-cluster endpoint of the SigNoz UI/API, e.g. http://signoz-main.observability.svc.cluster.local:8080 |
| `status.outputs.port_forward_command` | `string` | command to port-forward the UI to a developer laptop, e.g. kubectl port-forward svc/signoz-main -n observability 8080:8080 (then open http://localhost:8080). |
| `status.outputs.otel_collector_service` | `string` | name of the ingestion collector Service, e.g. signoz-main-otel-collector. |
| `status.outputs.otlp_grpc_endpoint` | `string` | in-cluster OTLP gRPC ingestion endpoint — point OTLP/gRPC exporters here, e.g. signoz-main-otel-collector.observability.svc.cluster.local:4317 |
| `status.outputs.otlp_http_endpoint` | `string` | in-cluster OTLP HTTP ingestion endpoint — point OTLP/HTTP exporters here, e.g. http://signoz-main-otel-collector.observability.svc.cluster.local:4318 |
| `status.outputs.clickhouse_endpoint` | `string` | ClickHouse native-protocol endpoint SigNoz stores telemetry in — a passthrough of the declared connection (this component installs no ClickHouse). Downstream kinds referencing it compose against the same store SigNoz uses. |
| `status.outputs.clickhouse_username` | `string` | ClickHouse username SigNoz connects as (mirrors the declared connection). |
| `status.outputs.clickhouse_password_secret` | `KubernetesSecretKey` | Secret key holding that user's password — the declared Secret reference, never module-owned; the Secret lives in the SigNoz namespace (a secretKeyRef cannot cross namespaces). |
| `status.outputs.clickhouse_password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.clickhouse_password_secret.key` | `string` | The key within the Kubernetes Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.clickhouse.host` | KubernetesClickHouse | `status.outputs.service_name` |
| `spec.clickhouse.clusterName` | KubernetesClickHouse | `status.outputs.cluster_name` |
| `spec.clickhouse.passwordSecret.secretName` | KubernetesClickHouse | `status.outputs.auth_secret_name` |
| `spec.server.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## See Also

- [Overview](../README.md)
