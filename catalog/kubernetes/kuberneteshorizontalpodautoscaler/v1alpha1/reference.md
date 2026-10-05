# KubernetesHorizontalPodAutoscaler

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesHorizontalPodAutoscalerSpec** scales a workload's replica count
automatically between a floor and a ceiling, driven by observed metrics —
the full autoscaling/v2 surface: resource utilization (CPU/memory),
per-container resources, custom per-pod metrics, metrics on other objects,
and external metrics (queue depths, cloud load balancer QPS), plus
fine-grained scaling BEHAVIOR (per-direction velocity policies and
stabilization windows).

When several metrics are configured, each proposes a replica count and the
HIGHEST wins — metrics OR together toward scale-out. Resource metrics
require a metrics source in the cluster (metrics-server for CPU/memory;
custom/external metrics need an adapter such as prometheus-adapter or
KEDA).

A standalone autoscaler is the right shape for scale targets a Planton
workload kind does not manage — operator-managed Deployments, non-Planton
workloads — and for the advanced v2 surface (custom/object/external
metrics, behavior tuning). For simple CPU/memory autoscaling of a Planton
Deployment's own replicas, prefer the workload's built-in
`availability.horizontal_pod_autoscaling` block. Never point BOTH at the
same target: two controllers fighting over one replica count flaps the
fleet.

The workload's `replicas` becomes advisory once an HPA governs it — the
autoscaler owns the count from then on.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises all five
# metric source families and both behavior directions with policies.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHorizontalPodAutoscaler
metadata:
  name: test-hpa
spec:
  namespace:
    value: default
  name: test-hpa
  labels:
    team: platform-engineering
  scale_target:
    api_version: apps/v1
    kind: Deployment
    name:
      value: test-app
  min_replicas: 2
  max_replicas: 20
  metrics:
    - type: resource
      resource:
        name: cpu
        target:
          type: utilization
          average_utilization: 60
    - type: container_resource
      container_resource:
        name: memory
        container: app
        target:
          type: average_value
          average_value: 512Mi
    - type: pods
      pods:
        metric:
          name: requests_per_second
        target:
          type: average_value
          average_value: "100"
    - type: object
      object:
        described_object:
          api_version: networking.k8s.io/v1
          kind: Ingress
          name: main-ingress
        metric:
          name: requests_per_second
        target:
          type: raw_value
          value: 10k
    - type: external
      external:
        metric:
          name: queue_messages_ready
          match_labels:
            queue: orders
        target:
          type: average_value
          average_value: "30"
  behavior:
    scale_up:
      stabilization_window_seconds: 0
      select_policy: max_change
      policies:
        - type: pods
          value: 4
          period_seconds: 15
        - type: percent
          value: 100
          period_seconds: 15
    scale_down:
      stabilization_window_seconds: 600
      select_policy: min_change
      policies:
        - type: percent
          value: 10
          period_seconds: 60
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` |  |  | KubernetesNamespace (`spec.name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.scaleTarget` | `KubernetesHorizontalPodAutoscalerScaleTarget` | yes |  |  |
| `spec.scaleTarget.apiVersion` | `string` |  | `apps/v1` |  |
| `spec.scaleTarget.kind` | `string` |  | `Deployment` |  |
| `spec.scaleTarget.name` | `string \| valueFrom` | yes |  | KubernetesDeployment (`status.outputs.deployment_name`) |
| `spec.minReplicas` | `int32` |  | `1` |  |
| `spec.maxReplicas` | `int32` |  |  |  |
| `spec.metrics` | `[]KubernetesHorizontalPodAutoscalerMetric` |  |  |  |
| `spec.metrics[].type` | `enum` |  |  |  |
| `spec.metrics[].resource` | `KubernetesHorizontalPodAutoscalerResourceMetric` |  |  |  |
| `spec.metrics[].resource.name` | `string` |  |  |  |
| `spec.metrics[].resource.target` | `KubernetesHorizontalPodAutoscalerMetricTarget` | yes |  |  |
| `spec.metrics[].resource.target.type` | `enum` |  |  |  |
| `spec.metrics[].resource.target.averageUtilization` | `int32` |  |  |  |
| `spec.metrics[].resource.target.value` | `string` |  |  |  |
| `spec.metrics[].resource.target.averageValue` | `string` |  |  |  |
| `spec.metrics[].containerResource` | `KubernetesHorizontalPodAutoscalerContainerResourceMetric` |  |  |  |
| `spec.metrics[].containerResource.name` | `string` |  |  |  |
| `spec.metrics[].containerResource.container` | `string` | yes |  |  |
| `spec.metrics[].containerResource.target` | `KubernetesHorizontalPodAutoscalerMetricTarget` | yes |  |  |
| `spec.metrics[].containerResource.target.type` | `enum` |  |  |  |
| `spec.metrics[].containerResource.target.averageUtilization` | `int32` |  |  |  |
| `spec.metrics[].containerResource.target.value` | `string` |  |  |  |
| `spec.metrics[].containerResource.target.averageValue` | `string` |  |  |  |
| `spec.metrics[].pods` | `KubernetesHorizontalPodAutoscalerPodsMetric` |  |  |  |
| `spec.metrics[].pods.metric` | `KubernetesHorizontalPodAutoscalerMetricIdentifier` | yes |  |  |
| `spec.metrics[].pods.metric.name` | `string` | yes |  |  |
| `spec.metrics[].pods.metric.matchLabels` | `map<string, string>` |  |  |  |
| `spec.metrics[].pods.target` | `KubernetesHorizontalPodAutoscalerMetricTarget` | yes |  |  |
| `spec.metrics[].pods.target.type` | `enum` |  |  |  |
| `spec.metrics[].pods.target.averageUtilization` | `int32` |  |  |  |
| `spec.metrics[].pods.target.value` | `string` |  |  |  |
| `spec.metrics[].pods.target.averageValue` | `string` |  |  |  |
| `spec.metrics[].object` | `KubernetesHorizontalPodAutoscalerObjectMetric` |  |  |  |
| `spec.metrics[].object.describedObject` | `KubernetesHorizontalPodAutoscalerObjectReference` | yes |  |  |
| `spec.metrics[].object.describedObject.apiVersion` | `string` |  |  |  |
| `spec.metrics[].object.describedObject.kind` | `string` | yes |  |  |
| `spec.metrics[].object.describedObject.name` | `string` | yes |  |  |
| `spec.metrics[].object.metric` | `KubernetesHorizontalPodAutoscalerMetricIdentifier` | yes |  |  |
| `spec.metrics[].object.metric.name` | `string` | yes |  |  |
| `spec.metrics[].object.metric.matchLabels` | `map<string, string>` |  |  |  |
| `spec.metrics[].object.target` | `KubernetesHorizontalPodAutoscalerMetricTarget` | yes |  |  |
| `spec.metrics[].object.target.type` | `enum` |  |  |  |
| `spec.metrics[].object.target.averageUtilization` | `int32` |  |  |  |
| `spec.metrics[].object.target.value` | `string` |  |  |  |
| `spec.metrics[].object.target.averageValue` | `string` |  |  |  |
| `spec.metrics[].external` | `KubernetesHorizontalPodAutoscalerExternalMetric` |  |  |  |
| `spec.metrics[].external.metric` | `KubernetesHorizontalPodAutoscalerMetricIdentifier` | yes |  |  |
| `spec.metrics[].external.metric.name` | `string` | yes |  |  |
| `spec.metrics[].external.metric.matchLabels` | `map<string, string>` |  |  |  |
| `spec.metrics[].external.target` | `KubernetesHorizontalPodAutoscalerMetricTarget` | yes |  |  |
| `spec.metrics[].external.target.type` | `enum` |  |  |  |
| `spec.metrics[].external.target.averageUtilization` | `int32` |  |  |  |
| `spec.metrics[].external.target.value` | `string` |  |  |  |
| `spec.metrics[].external.target.averageValue` | `string` |  |  |  |
| `spec.behavior` | `KubernetesHorizontalPodAutoscalerBehavior` |  |  |  |
| `spec.behavior.scaleUp` | `KubernetesHorizontalPodAutoscalerScalingRules` |  |  |  |
| `spec.behavior.scaleUp.stabilizationWindowSeconds` | `int32` |  |  |  |
| `spec.behavior.scaleUp.selectPolicy` | `enum` |  | `max_change` |  |
| `spec.behavior.scaleUp.policies` | `[]KubernetesHorizontalPodAutoscalerScalingPolicy` |  |  |  |
| `spec.behavior.scaleUp.policies[].type` | `enum` |  |  |  |
| `spec.behavior.scaleUp.policies[].value` | `int32` |  |  |  |
| `spec.behavior.scaleUp.policies[].periodSeconds` | `int32` |  |  |  |
| `spec.behavior.scaleDown` | `KubernetesHorizontalPodAutoscalerScalingRules` |  |  |  |
| `spec.behavior.scaleDown.stabilizationWindowSeconds` | `int32` |  |  |  |
| `spec.behavior.scaleDown.selectPolicy` | `enum` |  | `max_change` |  |
| `spec.behavior.scaleDown.policies` | `[]KubernetesHorizontalPodAutoscalerScalingPolicy` |  |  |  |
| `spec.behavior.scaleDown.policies[].type` | `enum` |  |  |  |
| `spec.behavior.scaleDown.policies[].value` | `int32` |  |  |  |
| `spec.behavior.scaleDown.policies[].periodSeconds` | `int32` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom`

The namespace the autoscaler lives in — it MUST be the scale target's
own namespace (an HPA cannot scale across namespaces). Accepts a literal
namespace name or a reference to a KubernetesNamespace resource. When
omitted, the autoscaler lands in the cluster's `default` namespace.

- references: KubernetesNamespace (`spec.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the HorizontalPodAutoscaler (its `metadata.name` in the
cluster). Must be a valid DNS subdomain: lowercase alphanumeric
characters, hyphens, and dots, at most 253 characters.

- rule: Name must be a valid DNS subdomain (lowercase alphanumeric, hyphens, and dots, no leading/trailing dots or hyphens)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Additional labels to apply to the HorizontalPodAutoscaler object.
Merged with the standard Planton governance labels.

### spec.annotations

`map<string, string>`

Annotations to apply to the HorizontalPodAutoscaler object.

### spec.scaleTarget

`KubernetesHorizontalPodAutoscalerScaleTarget` · required

The workload whose replica count this autoscaler owns.

- rule: {"required":true}

### spec.scaleTarget.apiVersion

`string` · optional (explicit presence)

The API version of the target.
Default: apps/v1

- default: `apps/v1`

### spec.scaleTarget.kind

`string` · optional (explicit presence)

The kind of the target — "Deployment" (the default), "StatefulSet",
"ReplicaSet", or a custom resource that exposes the scale subresource.
DaemonSets cannot be horizontally scaled (one pod per node by
definition).
Default: Deployment

- default: `Deployment`
- rule: A DaemonSet cannot be a scale target — it runs one pod per node by definition

### spec.scaleTarget.name

`string | valueFrom` · required

The name of the target workload, in the autoscaler's own namespace.
Accepts a literal name or a reference to a KubernetesDeployment resource
(for other kinds, reference their exported name output explicitly).

- references: KubernetesDeployment (`status.outputs.deployment_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesDeployment, name: <that resource's name>, fieldPath: status.outputs.deployment_name}} -- a bare string does not parse

### spec.minReplicas

`int32` · optional (explicit presence)

The replica floor — the autoscaler never scales below it. Kubernetes
defaults to 1. (Scaling to zero is feature-gated upstream and requires
at least one object/external metric; it is deliberately not modeled.)

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.maxReplicas

`int32`

The replica ceiling — required, and the honest capacity conversation:
what is the most this workload may cost? Must be at least min_replicas.

- rule: {"int32":{"gte":1}}

### spec.metrics

`[]KubernetesHorizontalPodAutoscalerMetric`

The metrics driving the scale decision. Each proposes a replica count;
the highest wins. When EMPTY, Kubernetes applies its default: 80%
average CPU utilization (requires pods to declare CPU requests —
utilization is measured against requests).

- rule: Set exactly the source field matching the metric type (resource → resource, container_resource → container_resource, pods → pods, object → object, external → external)

### spec.metrics[].type

`enum`

The metric source family; exactly the matching source field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_type_unspecified` -- Unspecified.
- `resource` -- A pod resource (cpu/memory) averaged across the target's pods — requires metrics-server. The workhorse.
- `container_resource` -- A resource metric for ONE named container in each pod — isolates the app container from sidecars that would skew the pod-level average.
- `pods` -- A custom metric exposed per pod (e.g. requests-per-second), averaged — requires a custom-metrics adapter.
- `object` -- A metric describing ONE other object (e.g. an Ingress's requests-per-second) — requires a custom-metrics adapter.
- `external` -- A metric from outside the cluster (queue depth, cloud LB QPS) — requires an external-metrics adapter. For event-driven scale-to-zero semantics, KEDA builds on this.

### spec.metrics[].resource

`KubernetesHorizontalPodAutoscalerResourceMetric`

Resource metric source (type: resource).

### spec.metrics[].resource.name

`string`

The resource: "cpu" or "memory". CPU is the reliable scaling signal;
memory rarely falls after scale-out, which makes it a poor scale-in
driver.

- rule: {"string":{"in":["cpu","memory"]}}

### spec.metrics[].resource.target

`KubernetesHorizontalPodAutoscalerMetricTarget` · required

The target the average is held at.

- rule: {"required":true}
- rule: Set exactly the value field matching the target type: average_utilization for utilization, value for raw_value, average_value for average_value

### spec.metrics[].resource.target.type

`enum`

How the target value is expressed; the matching value field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_target_type_unspecified` -- Unspecified.
- `utilization` -- A percentage of the pods' RESOURCE REQUESTS, averaged (e.g. 60 = hold average usage at 60% of requests). Resource and container_resource metrics only; requires the pods to declare requests.
- `raw_value` -- A raw quantity compared to the metric directly (object/external metrics). Maps to the API's "Value" target type — named raw_value here because generated code cannot use the bare word "value" as an enum constant.
- `average_value` -- A quantity averaged across the target's pods (pods metrics; the usual form for object/external too, e.g. "30 messages per pod").

### spec.metrics[].resource.target.averageUtilization

`int32` · optional (explicit presence)

Target percentage of requests for utilization targets (e.g. 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.metrics[].resource.target.value

`string`

Target raw quantity for value targets (e.g. "1k", "100").

- rule: value must be a Kubernetes quantity (e.g. "100", "1k", "500m")

### spec.metrics[].resource.target.averageValue

`string`

Target per-pod average quantity for average_value targets (e.g. "30",
"500Mi").

- rule: average_value must be a Kubernetes quantity (e.g. "30", "500Mi")

### spec.metrics[].containerResource

`KubernetesHorizontalPodAutoscalerContainerResourceMetric`

Container resource metric source (type: container_resource).

### spec.metrics[].containerResource.name

`string`

The resource: "cpu" or "memory".

- rule: {"string":{"in":["cpu","memory"]}}

### spec.metrics[].containerResource.container

`string` · required

The container name within the target's pods (e.g. the app container's
name), as declared in the pod template.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].containerResource.target

`KubernetesHorizontalPodAutoscalerMetricTarget` · required

The target the average is held at.

- rule: {"required":true}
- rule: Set exactly the value field matching the target type: average_utilization for utilization, value for raw_value, average_value for average_value

### spec.metrics[].containerResource.target.type

`enum`

How the target value is expressed; the matching value field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_target_type_unspecified` -- Unspecified.
- `utilization` -- A percentage of the pods' RESOURCE REQUESTS, averaged (e.g. 60 = hold average usage at 60% of requests). Resource and container_resource metrics only; requires the pods to declare requests.
- `raw_value` -- A raw quantity compared to the metric directly (object/external metrics). Maps to the API's "Value" target type — named raw_value here because generated code cannot use the bare word "value" as an enum constant.
- `average_value` -- A quantity averaged across the target's pods (pods metrics; the usual form for object/external too, e.g. "30 messages per pod").

### spec.metrics[].containerResource.target.averageUtilization

`int32` · optional (explicit presence)

Target percentage of requests for utilization targets (e.g. 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.metrics[].containerResource.target.value

`string`

Target raw quantity for value targets (e.g. "1k", "100").

- rule: value must be a Kubernetes quantity (e.g. "100", "1k", "500m")

### spec.metrics[].containerResource.target.averageValue

`string`

Target per-pod average quantity for average_value targets (e.g. "30",
"500Mi").

- rule: average_value must be a Kubernetes quantity (e.g. "30", "500Mi")

### spec.metrics[].pods

`KubernetesHorizontalPodAutoscalerPodsMetric`

Per-pod custom metric source (type: pods).

### spec.metrics[].pods.metric

`KubernetesHorizontalPodAutoscalerMetricIdentifier` · required

The metric identity, as exposed by the custom-metrics adapter.

- rule: {"required":true}

### spec.metrics[].pods.metric.name

`string` · required

The metric name, as the metrics adapter exposes it.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].pods.metric.matchLabels

`map<string, string>`

Label selector passed to the metrics adapter to scope the series (e.g.
{queue: "orders"}). Omit to read the metric unscoped.

### spec.metrics[].pods.target

`KubernetesHorizontalPodAutoscalerMetricTarget` · required

The target — pods metrics support average_value (per-pod average).

- rule: {"required":true}
- rule: Set exactly the value field matching the target type: average_utilization for utilization, value for raw_value, average_value for average_value

### spec.metrics[].pods.target.type

`enum`

How the target value is expressed; the matching value field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_target_type_unspecified` -- Unspecified.
- `utilization` -- A percentage of the pods' RESOURCE REQUESTS, averaged (e.g. 60 = hold average usage at 60% of requests). Resource and container_resource metrics only; requires the pods to declare requests.
- `raw_value` -- A raw quantity compared to the metric directly (object/external metrics). Maps to the API's "Value" target type — named raw_value here because generated code cannot use the bare word "value" as an enum constant.
- `average_value` -- A quantity averaged across the target's pods (pods metrics; the usual form for object/external too, e.g. "30 messages per pod").

### spec.metrics[].pods.target.averageUtilization

`int32` · optional (explicit presence)

Target percentage of requests for utilization targets (e.g. 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.metrics[].pods.target.value

`string`

Target raw quantity for value targets (e.g. "1k", "100").

- rule: value must be a Kubernetes quantity (e.g. "100", "1k", "500m")

### spec.metrics[].pods.target.averageValue

`string`

Target per-pod average quantity for average_value targets (e.g. "30",
"500Mi").

- rule: average_value must be a Kubernetes quantity (e.g. "30", "500Mi")

### spec.metrics[].object

`KubernetesHorizontalPodAutoscalerObjectMetric`

Object metric source (type: object).

### spec.metrics[].object.describedObject

`KubernetesHorizontalPodAutoscalerObjectReference` · required

The object the metric describes (e.g. an Ingress).

- rule: {"required":true}

### spec.metrics[].object.describedObject.apiVersion

`string`

The API version of the described object (e.g. "networking.k8s.io/v1").

### spec.metrics[].object.describedObject.kind

`string` · required

The kind of the described object (e.g. "Ingress").

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].object.describedObject.name

`string` · required

The name of the described object, in the autoscaler's namespace.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].object.metric

`KubernetesHorizontalPodAutoscalerMetricIdentifier` · required

The metric identity, as exposed by the custom-metrics adapter.

- rule: {"required":true}

### spec.metrics[].object.metric.name

`string` · required

The metric name, as the metrics adapter exposes it.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].object.metric.matchLabels

`map<string, string>`

Label selector passed to the metrics adapter to scope the series (e.g.
{queue: "orders"}). Omit to read the metric unscoped.

### spec.metrics[].object.target

`KubernetesHorizontalPodAutoscalerMetricTarget` · required

The target — object metrics support value (the object's metric directly)
or average_value (divided across the target's pods).

- rule: {"required":true}
- rule: Set exactly the value field matching the target type: average_utilization for utilization, value for raw_value, average_value for average_value

### spec.metrics[].object.target.type

`enum`

How the target value is expressed; the matching value field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_target_type_unspecified` -- Unspecified.
- `utilization` -- A percentage of the pods' RESOURCE REQUESTS, averaged (e.g. 60 = hold average usage at 60% of requests). Resource and container_resource metrics only; requires the pods to declare requests.
- `raw_value` -- A raw quantity compared to the metric directly (object/external metrics). Maps to the API's "Value" target type — named raw_value here because generated code cannot use the bare word "value" as an enum constant.
- `average_value` -- A quantity averaged across the target's pods (pods metrics; the usual form for object/external too, e.g. "30 messages per pod").

### spec.metrics[].object.target.averageUtilization

`int32` · optional (explicit presence)

Target percentage of requests for utilization targets (e.g. 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.metrics[].object.target.value

`string`

Target raw quantity for value targets (e.g. "1k", "100").

- rule: value must be a Kubernetes quantity (e.g. "100", "1k", "500m")

### spec.metrics[].object.target.averageValue

`string`

Target per-pod average quantity for average_value targets (e.g. "30",
"500Mi").

- rule: average_value must be a Kubernetes quantity (e.g. "30", "500Mi")

### spec.metrics[].external

`KubernetesHorizontalPodAutoscalerExternalMetric`

External metric source (type: external).

### spec.metrics[].external.metric

`KubernetesHorizontalPodAutoscalerMetricIdentifier` · required

The metric identity, as exposed by the external-metrics adapter. The
selector narrows which series is read (e.g. one queue's depth).

- rule: {"required":true}

### spec.metrics[].external.metric.name

`string` · required

The metric name, as the metrics adapter exposes it.

- rule: {"string":{"minLen":"1"}}

### spec.metrics[].external.metric.matchLabels

`map<string, string>`

Label selector passed to the metrics adapter to scope the series (e.g.
{queue: "orders"}). Omit to read the metric unscoped.

### spec.metrics[].external.target

`KubernetesHorizontalPodAutoscalerMetricTarget` · required

The target — external metrics support value or average_value (divided
across the target's pods; the usual choice, e.g. "30 queue messages per
pod").

- rule: {"required":true}
- rule: Set exactly the value field matching the target type: average_utilization for utilization, value for raw_value, average_value for average_value

### spec.metrics[].external.target.type

`enum`

How the target value is expressed; the matching value field must be set.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_metric_target_type_unspecified` -- Unspecified.
- `utilization` -- A percentage of the pods' RESOURCE REQUESTS, averaged (e.g. 60 = hold average usage at 60% of requests). Resource and container_resource metrics only; requires the pods to declare requests.
- `raw_value` -- A raw quantity compared to the metric directly (object/external metrics). Maps to the API's "Value" target type — named raw_value here because generated code cannot use the bare word "value" as an enum constant.
- `average_value` -- A quantity averaged across the target's pods (pods metrics; the usual form for object/external too, e.g. "30 messages per pod").

### spec.metrics[].external.target.averageUtilization

`int32` · optional (explicit presence)

Target percentage of requests for utilization targets (e.g. 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.metrics[].external.target.value

`string`

Target raw quantity for value targets (e.g. "1k", "100").

- rule: value must be a Kubernetes quantity (e.g. "100", "1k", "500m")

### spec.metrics[].external.target.averageValue

`string`

Target per-pod average quantity for average_value targets (e.g. "30",
"500Mi").

- rule: average_value must be a Kubernetes quantity (e.g. "30", "500Mi")

### spec.behavior

`KubernetesHorizontalPodAutoscalerBehavior`

Per-direction scaling velocity and stabilization tuning. Omit for the
Kubernetes defaults: scale up fast (double or +4 pods per 15s, no
stabilization), scale down conservatively (to the recommendation's
300-second high-water mark).

### spec.behavior.scaleUp

`KubernetesHorizontalPodAutoscalerScalingRules`

Scale-UP tuning. Kubernetes default: no stabilization, allow the higher
of doubling or +4 pods per 15 seconds.

- rule: A direction with select_policy "disabled" must not list policies — disabled turns the direction off entirely

### spec.behavior.scaleUp.stabilizationWindowSeconds

`int32` · optional (explicit presence)

Seconds of past recommendations considered before acting in this
direction (0–3600). The flap damper: the safest recommendation in the
window wins. 0 = act immediately.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.behavior.scaleUp.selectPolicy

`enum` · optional (explicit presence)

Which of the listed policies applies.
Default: max_change

- default: `max_change`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_select_policy_unspecified` -- Unspecified. Defaults to max_change.
- `max_change` -- Use the policy allowing the LARGEST change — the default.
- `min_change` -- Use the policy allowing the SMALLEST change — the conservative combinator.
- `disabled` -- Disable scaling in this direction entirely (e.g. freeze scale-down during an incident while leaving scale-up live).

### spec.behavior.scaleUp.policies

`[]KubernetesHorizontalPodAutoscalerScalingPolicy`

The velocity policies for this direction. Each caps the change per
period; select_policy arbitrates between them.

### spec.behavior.scaleUp.policies[].type

`enum`

The unit of the cap.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_scaling_policy_type_unspecified` -- Unspecified.
- `pods` -- Cap by absolute pod count per period.
- `percent` -- Cap by percentage of current replicas per period.

### spec.behavior.scaleUp.policies[].value

`int32`

The maximum change per period: pod count (type: pods) or percentage of
current replicas (type: percent). Must be positive.

- rule: {"int32":{"gte":1}}

### spec.behavior.scaleUp.policies[].periodSeconds

`int32`

The window the cap applies over, in seconds (1–1800).

- rule: {"int32":{"lte":1800,"gte":1}}

### spec.behavior.scaleDown

`KubernetesHorizontalPodAutoscalerScalingRules`

Scale-DOWN tuning. Kubernetes default: 300-second stabilization (scale
down to the recommendation's 5-minute high-water mark), all pods
removable in one step. Lengthen the window for spiky traffic; add a
percent policy to bleed capacity gradually.

- rule: A direction with select_policy "disabled" must not list policies — disabled turns the direction off entirely

### spec.behavior.scaleDown.stabilizationWindowSeconds

`int32` · optional (explicit presence)

Seconds of past recommendations considered before acting in this
direction (0–3600). The flap damper: the safest recommendation in the
window wins. 0 = act immediately.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.behavior.scaleDown.selectPolicy

`enum` · optional (explicit presence)

Which of the listed policies applies.
Default: max_change

- default: `max_change`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_select_policy_unspecified` -- Unspecified. Defaults to max_change.
- `max_change` -- Use the policy allowing the LARGEST change — the default.
- `min_change` -- Use the policy allowing the SMALLEST change — the conservative combinator.
- `disabled` -- Disable scaling in this direction entirely (e.g. freeze scale-down during an incident while leaving scale-up live).

### spec.behavior.scaleDown.policies

`[]KubernetesHorizontalPodAutoscalerScalingPolicy`

The velocity policies for this direction. Each caps the change per
period; select_policy arbitrates between them.

### spec.behavior.scaleDown.policies[].type

`enum`

The unit of the cap.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `kubernetes_horizontal_pod_autoscaler_scaling_policy_type_unspecified` -- Unspecified.
- `pods` -- Cap by absolute pod count per period.
- `percent` -- Cap by percentage of current replicas per period.

### spec.behavior.scaleDown.policies[].value

`int32`

The maximum change per period: pod count (type: pods) or percentage of
current replicas (type: percent). Must be positive.

- rule: {"int32":{"gte":1}}

### spec.behavior.scaleDown.policies[].periodSeconds

`int32`

The window the cap applies over, in seconds (1–1800).

- rule: {"int32":{"lte":1800,"gte":1}}

## Validation Rules

- `replicas.ceiling_gte_floor`: max_replicas must be greater than or equal to min_replicas

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesHorizontalPodAutoscaler, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.horizontal_pod_autoscaler_name` | `string` | The name of the HorizontalPodAutoscaler object as created in the cluster. |
| `status.outputs.namespace` | `string` | The namespace the autoscaler was created in. |
| `status.outputs.scale_target` | `string` | The scale target as "Kind/name" (e.g. "Deployment/checkout") — the workload whose replica count this autoscaler owns. |
| `status.outputs.min_replicas` | `int32` | The configured replica floor. |
| `status.outputs.max_replicas` | `int32` | The configured replica ceiling. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.scaleTarget.name` | KubernetesDeployment | `status.outputs.deployment_name` |

## See Also

- [Overview](../README.md)
