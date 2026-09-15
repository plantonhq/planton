# DigitalOceanMonitorAlert

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanMonitorAlertSpec models the full digitalocean_monitor_alert
resource surface: an alert policy on DigitalOcean's built-in metrics for
Droplets, load balancers, and managed database clusters, with email and
Slack notification channels.

DigitalOcean's API targets a policy through one untyped id list plus a
tag list; this spec replaces the untyped list with one TYPED reference
list per resource family, so an id can never be paired with the wrong
metric family and resources are wired by reference instead of
hand-copied ids. The modules merge the lists back into the provider's
single entities argument. Every field updates in place.

## Example

```yaml
# Reference manifests for DigitalOceanMonitorAlert -- protovalidate-valid,
# embedded as the reference page's Example block, and the documents the
# offline tofu plans render. Two documents: a droplet CPU alert targeting
# one droplet by reference, and a tag-targeted memory alert with Slack
# delivery.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanMonitorAlert
metadata:
  name: web-cpu-alert
spec:
  description: CPU of web droplets above 90 percent
  metricType: v1/insights/droplet/cpu
  compare: GreaterThan
  value: 90
  window: 5m
  dropletIds:
    # Literal numeric droplet id; use valueFrom to reference a
    # DigitalOceanDroplet resource instead.
    - value: "123456789"
  alerts:
    emails:
      - ops@example.com
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanMonitorAlert
metadata:
  name: fleet-memory-alert
spec:
  description: Memory of the web fleet above 85 percent
  metricType: v1/insights/droplet/memory_utilization_percent
  compare: GreaterThan
  value: 85
  window: 10m
  tags:
    - web
  alerts:
    emails:
      - ops@example.com
    slack:
      - channel: "#alerts"
        url: https://hooks.slack.com/services/EXAMPLE/EXAMPLE/EXAMPLE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.description` | `string` | yes |  |  |
| `spec.metricType` | `string` | yes |  |  |
| `spec.compare` | `string` | yes |  |  |
| `spec.value` | `double` | yes |  |  |
| `spec.window` | `string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.dropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |
| `spec.loadBalancerIds` | `[]string \| valueFrom` |  |  | DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`) |
| `spec.databaseClusterIds` | `[]string \| valueFrom` |  |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.tags` | `[]string` |  |  |  |
| `spec.alerts` | `DigitalOceanMonitorAlertNotifications` | yes |  |  |
| `spec.alerts.emails` | `[]string` |  |  |  |
| `spec.alerts.slack` | `[]DigitalOceanMonitorAlertSlack` |  |  |  |
| `spec.alerts.slack[].channel` | `string` | yes |  |  |
| `spec.alerts.slack[].url` | `string` (sensitive) | yes |  |  |

## Field Details

### spec.description

`string` · required

Human-readable description of the alert policy. DigitalOcean has no
separate name field -- this is the policy's display handle in the
console and in notification messages.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.metricType

`string` · required

The metric the policy watches. DigitalOcean namespaces metrics per
resource family: Droplet metrics under v1/insights/droplet/, load
balancer metrics under v1/insights/lbaas/, and managed database
metrics under v1/dbaas/alerts/. The values are DigitalOcean's own API
paths and their inconsistencies are deliberate facts of that API --
droplet CPU is bare "cpu" (no _utilization_percent suffix) and the
database family uses an _alerts suffix -- never "corrected" here.

- rule: {"required":true,"string":{"in":["v1/insights/droplet/cpu","v1/insights/droplet/memory_utilization_percent","v1/insights/droplet/disk_utilization_percent","v1/insights/droplet/public_outbound_bandwidth","v1/insights/droplet/public_inbound_bandwidth","v1/insights/droplet/private_outbound_bandwidth","v1/insights/droplet/private_inbound_bandwidth","v1/insights/droplet/disk_read","v1/insights/droplet/disk_write","v1/insights/droplet/load_1","v1/insights/droplet/load_5","v1/insights/droplet/load_15","v1/insights/lbaas/avg_cpu_utilization_percent","v1/insights/lbaas/connection_utilization_percent","v1/insights/lbaas/droplet_health","v1/insights/lbaas/tls_connections_per_second_utilization_percent","v1/insights/lbaas/increase_in_http_error_rate_percentage_4xx","v1/insights/lbaas/increase_in_http_error_rate_percentage_5xx","v1/insights/lbaas/increase_in_http_error_rate_count_4xx","v1/insights/lbaas/increase_in_http_error_rate_count_5xx","v1/insights/lbaas/high_http_request_response_time","v1/insights/lbaas/high_http_request_response_time_50p","v1/insights/lbaas/high_http_request_response_time_95p","v1/insights/lbaas/high_http_request_response_time_99p","v1/dbaas/alerts/load_15_alerts","v1/dbaas/alerts/memory_utilization_alerts","v1/dbaas/alerts/disk_utilization_alerts","v1/dbaas/alerts/cpu_alerts"]}}

### spec.compare

`string` · required

How the measured value is compared against the threshold. CamelCase is
this API's spelling; uptime alerts spell the same concept snake_case
(greater_than) -- the two are different DigitalOcean APIs and are
deliberately not unified.

- rule: {"required":true,"string":{"in":["GreaterThan","LessThan"]}}

### spec.value

`double` · required

The threshold the metric is compared against. Units follow the metric:
percent for utilization metrics, load units for load_*, bytes per
second for bandwidth and disk I/O. DigitalOcean stores it as a 32-bit
float, so more than 7 significant digits are silently truncated
upstream.

- rule: {"required":true,"double":{"gte":0}}

### spec.window

`string` · required

The sampling window the metric is aggregated over before comparison.

- rule: {"required":true,"string":{"in":["5m","10m","30m","1h"]}}

### spec.enabled

`bool` · optional (explicit presence)

(Optional) Whether the alert policy is active. Unset defers to
DigitalOcean's default, which is enabled; set false to keep the policy
defined but silent.

- default: `true`

### spec.dropletIds

`[]string | valueFrom`

(Optional) Droplets the policy watches, as literal numeric Droplet ids
or references to DigitalOceanDroplet resources. Valid only with
droplet metrics.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

### spec.loadBalancerIds

`[]string | valueFrom`

(Optional) Load balancers the policy watches, as literal UUIDs or
references to DigitalOceanLoadBalancer resources. Valid only with
load-balancer metrics.

- references: DigitalOceanLoadBalancer (`status.outputs.load_balancer_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.load_balancer_id}} -- a bare string does not parse

### spec.databaseClusterIds

`[]string | valueFrom`

(Optional) Managed database clusters the policy watches, as literal
UUIDs or references to DigitalOceanDatabaseCluster resources. Valid
only with database metrics.

Containment-exempt: an alert policy WATCHES the clusters it names and
never lives in one. On a diagram the alert stays with the watching
layer, with a line to each cluster it watches.

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.tags

`[]string`

(Optional) Droplet tags the policy watches: every Droplet carrying a
listed tag is covered, and membership tracks the tag automatically --
prefer tags over droplet_ids for dynamic fleets. Valid only with
droplet metrics.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.alerts

`DigitalOceanMonitorAlertNotifications` · required

Where alert notifications are delivered. At least one channel is
required -- DigitalOcean rejects a policy that notifies nobody.

- rule: {"required":true}
- rule: at least one notification channel (emails or slack) must be specified

### spec.alerts.emails

`[]string`

(Optional) Email addresses notifications are sent to. DigitalOcean may
require addresses to belong to verified account members -- it rejects
unknown addresses at request time.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.alerts.slack

`[]DigitalOceanMonitorAlertSlack`

(Optional) Slack channels notifications are posted to.

### spec.alerts.slack[].channel

`string` · required

The Slack channel to post to (for example "#alerts").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.alerts.slack[].url

`string` · required · sensitive

The Slack incoming-webhook URL. A credential: DigitalOcean's API does
not mark it sensitive, so it is marked sensitive here and both
provisioners keep it out of plain-text state rendering.

- rule: {"required":true,"string":{"minLen":"1"}}

## Validation Rules

- `spec.droplet_metric_targets`: a droplet metric (v1/insights/droplet/*) targets droplet_ids and/or tags -- not load_balancer_ids or database_cluster_ids
- `spec.lbaas_metric_targets`: a load-balancer metric (v1/insights/lbaas/*) targets load_balancer_ids only -- not droplet_ids, database_cluster_ids, or tags
- `spec.dbaas_metric_targets`: a database metric (v1/dbaas/alerts/*) targets database_cluster_ids only -- not droplet_ids, load_balancer_ids, or tags

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanMonitorAlert, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.alert_id` | `string` | UUID of the alert policy (the API identity, and the import id). Both provisioners read it from the resource id -- the provider's own uuid attribute is declared but never populated at the pinned version. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |
| `spec.loadBalancerIds` | DigitalOceanLoadBalancer | `status.outputs.load_balancer_id` |
| `spec.databaseClusterIds` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
