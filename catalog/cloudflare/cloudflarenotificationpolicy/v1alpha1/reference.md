# CloudflareNotificationPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareNotificationPolicySpec creates a notification policy: "when
THIS happens, tell THESE destinations." One policy watches one alert type
(a tunnel degrading, an origin failing health checks, a usage threshold,
a DDoS event, ...) and fans the alert out to email addresses, PagerDuty
services, and webhook destinations. A plain CRUD object -- real create,
update, delete.

The `filters` message narrows WHICH events of the alert type fire the
policy. Which filter fields a given alert type reads is owned by
Cloudflare's API (69 types x 43 filter fields is an API-side pairing);
each filter field's comment names the alert families that use it, with
the provider's own wording. Many alert types are themselves gated by plan
or product subscription -- the gate is Cloudflare's, surfaced at create
time.

## Example

```yaml
# Complete example manifest for CloudflareNotificationPolicy.
# Pages the on-call inbox and a webhook destination when origin health
# checks fail, narrowed to one health check status.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareNotificationPolicy
metadata:
  name: origin-health-alerts
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-health-alerts
  alert_type: health_check_status_notification
  description: page on-call when an origin health check turns unhealthy
  mechanisms:
    emails:
      - oncall@example.com
    webhook_ids:
      - value: "REPLACE_WITH_NOTIFICATION_WEBHOOK_UUID"
  filters:
    status:
      - Unhealthy
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.alertType` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  |  |  |
| `spec.alertInterval` | `string` |  |  |  |
| `spec.mechanisms` | `CloudflareNotificationPolicyMechanisms` | yes |  |  |
| `spec.mechanisms.emails` | `[]string` |  |  |  |
| `spec.mechanisms.pagerdutyIds` | `[]string` |  |  |  |
| `spec.mechanisms.webhookIds` | `[]string \| valueFrom` |  |  | CloudflareNotificationWebhook (`status.outputs.webhook_id`) |
| `spec.filters` | `CloudflareNotificationPolicyFilters` |  |  |  |
| `spec.filters.actions` | `[]string` |  |  |  |
| `spec.filters.affectedAsns` | `[]string` |  |  |  |
| `spec.filters.affectedComponents` | `[]string` |  |  |  |
| `spec.filters.affectedLocations` | `[]string` |  |  |  |
| `spec.filters.airportCode` | `[]string` |  |  |  |
| `spec.filters.alertTriggerPreferences` | `[]string` |  |  |  |
| `spec.filters.alertTriggerPreferencesValue` | `[]string` |  |  |  |
| `spec.filters.enabled` | `[]string` |  |  |  |
| `spec.filters.environment` | `[]string` |  |  |  |
| `spec.filters.event` | `[]string` |  |  |  |
| `spec.filters.eventSource` | `[]string` |  |  |  |
| `spec.filters.eventType` | `[]string` |  |  |  |
| `spec.filters.groupBy` | `[]string` |  |  |  |
| `spec.filters.healthCheckId` | `[]string` |  |  |  |
| `spec.filters.incidentImpact` | `[]string` |  |  |  |
| `spec.filters.inputId` | `[]string` |  |  |  |
| `spec.filters.insightClass` | `[]string` |  |  |  |
| `spec.filters.limit` | `[]string` |  |  |  |
| `spec.filters.logoTag` | `[]string` |  |  |  |
| `spec.filters.megabitsPerSecond` | `[]string` |  |  |  |
| `spec.filters.newHealth` | `[]string` |  |  |  |
| `spec.filters.newStatus` | `[]string` |  |  |  |
| `spec.filters.packetsPerSecond` | `[]string` |  |  |  |
| `spec.filters.poolId` | `[]string` |  |  |  |
| `spec.filters.popNames` | `[]string` |  |  |  |
| `spec.filters.product` | `[]string` |  |  |  |
| `spec.filters.projectId` | `[]string` |  |  |  |
| `spec.filters.protocol` | `[]string` |  |  |  |
| `spec.filters.queryTag` | `[]string` |  |  |  |
| `spec.filters.requestsPerSecond` | `[]string` |  |  |  |
| `spec.filters.selectors` | `[]string` |  |  |  |
| `spec.filters.services` | `[]string` |  |  |  |
| `spec.filters.slo` | `[]string` |  |  |  |
| `spec.filters.status` | `[]string` |  |  |  |
| `spec.filters.targetHostname` | `[]string` |  |  |  |
| `spec.filters.targetIp` | `[]string` |  |  |  |
| `spec.filters.targetZoneName` | `[]string` |  |  |  |
| `spec.filters.trafficExclusions` | `[]string` |  |  |  |
| `spec.filters.tunnelId` | `[]string` |  |  |  |
| `spec.filters.tunnelName` | `[]string` |  |  |  |
| `spec.filters.type` | `[]string` |  |  |  |
| `spec.filters.where` | `[]string` |  |  |  |
| `spec.filters.zones` | `[]string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the policy belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The policy's name, shown in the dashboard's notifications list.

- rule: {"required":true}

### spec.alertType

`string` · required

The event class this policy watches. The 69 values are the provider's
own list at v5.23.0.

- rule: alert_type must be one of the provider's notification alert types at v5.23.0 (abuse_report_alert, access_custom_certificate_expiration_type, advanced_ddos_attack_l4_alert, advanced_ddos_attack_l7_alert, advanced_http_alert_error, bgp_hijack_notification, billing_usage_alert, block_notification_block_removed, block_notification_new_block, block_notification_review_rejected, bot_traffic_basic_alert, brand_protection_alert, brand_protection_digest, clickhouse_alert_fw_anomaly, clickhouse_alert_fw_ent_anomaly, cloudforce_one_request_notification, cni_maintenance_notification, custom_analytics, custom_bot_detection_alert, custom_ssl_certificate_event_type, dedicated_ssl_certificate_event_type, device_connectivity_anomaly_alert, dos_attack_l4, dos_attack_l7, expiring_service_token_alert, failing_logpush_job_disabled_alert, fbm_auto_advertisement, fbm_dosd_attack, fbm_volumetric_attack, health_check_status_notification, hostname_aop_custom_certificate_expiration_type, http_alert_edge_error, http_alert_origin_error, image_notification, image_resizing_notification, incident_alert, load_balancing_health_alert, load_balancing_pool_enablement_alert, logo_match_alert, magic_tunnel_health_check_event, magic_wan_tunnel_health, maintenance_event_notification, mtls_certificate_store_certificate_expiration_type, pages_event_alert, radar_notification, real_origin_monitoring, scriptmonitor_alert_new_code_change_detections, scriptmonitor_alert_new_hosts, scriptmonitor_alert_new_malicious_hosts, scriptmonitor_alert_new_malicious_scripts, scriptmonitor_alert_new_malicious_url, scriptmonitor_alert_new_max_length_resource_url, scriptmonitor_alert_new_resources, secondary_dns_all_primaries_failing, secondary_dns_primaries_failing, secondary_dns_warning, secondary_dns_zone_successfully_updated, secondary_dns_zone_validation_warning, security_insights_alert, sentinel_alert, stream_live_notifications, synthetic_test_latency_alert, synthetic_test_low_availability_alert, traffic_anomalies_alert, tunnel_health_event, tunnel_update_event, universal_ssl_event_type, web_analytics_metrics_update, zone_aop_custom_certificate_expiration_type)
- rule: {"required":true}

### spec.description

`string`

A description of what the policy watches, shown in the dashboard.

### spec.enabled

`bool` · optional (explicit presence)

Whether the policy fires. Cloudflare's default is enabled; the modules
send the flag only when set.

### spec.alertInterval

`string`

Minimum time between successive notifications for the same condition
(e.g. "30m", "2h"). Only some alert types honor a refire interval;
empty uses Cloudflare's per-type behavior.

### spec.mechanisms

`CloudflareNotificationPolicyMechanisms` · required

Where alerts are delivered. At least one destination must be declared.

- rule: {"required":true}
- rule: declare at least one destination (an email address, a PagerDuty service, or a webhook)

### spec.mechanisms.emails

`[]string`

Email addresses to notify.

### spec.mechanisms.pagerdutyIds

`[]string`

PagerDuty service UUIDs to page (from the account's PagerDuty
integrations, connected in the Cloudflare dashboard).

### spec.mechanisms.webhookIds

`[]string | valueFrom`

Webhook destinations to invoke, by webhook UUID.
When using value_from, defaults to CloudflareNotificationWebhook kind and status.outputs.webhook_id field path.

- references: CloudflareNotificationWebhook (`status.outputs.webhook_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareNotificationWebhook, name: <that resource's name>, fieldPath: status.outputs.webhook_id}} -- a bare string does not parse

### spec.filters

`CloudflareNotificationPolicyFilters`

Narrows which events of the alert type fire this policy. Empty fires on
every event of the type.

### spec.filters.actions

`[]string`

Usage depends on specific alert type.

### spec.filters.affectedAsns

`[]string`

Used for configuring radar_notification.

### spec.filters.affectedComponents

`[]string`

Used for configuring incident_alert.

### spec.filters.affectedLocations

`[]string`

Used for configuring radar_notification.

### spec.filters.airportCode

`[]string`

Used for configuring maintenance_event_notification.

### spec.filters.alertTriggerPreferences

`[]string`

Usage depends on specific alert type.

### spec.filters.alertTriggerPreferencesValue

`[]string`

Usage depends on specific alert type.

### spec.filters.enabled

`[]string`

Used for configuring load_balancing_pool_enablement_alert (the pool
enablement state, as "true"/"false" strings).

### spec.filters.environment

`[]string`

Used for configuring pages_event_alert.

### spec.filters.event

`[]string`

Used for configuring pages_event_alert.

### spec.filters.eventSource

`[]string`

Used for configuring load_balancing_health_alert.

### spec.filters.eventType

`[]string`

Usage depends on specific alert type.

### spec.filters.groupBy

`[]string`

Usage depends on specific alert type.

### spec.filters.healthCheckId

`[]string`

Used for configuring health_check_status_notification.

### spec.filters.incidentImpact

`[]string`

Used for configuring incident_alert.

- rule: incident_impact entries must be one of INCIDENT_IMPACT_NONE, INCIDENT_IMPACT_MINOR, INCIDENT_IMPACT_MAJOR, INCIDENT_IMPACT_CRITICAL

### spec.filters.inputId

`[]string`

Used for configuring stream_live_notifications.

### spec.filters.insightClass

`[]string`

Used for configuring security_insights_alert.

### spec.filters.limit

`[]string`

Used for configuring billing_usage_alert (the usage threshold).

### spec.filters.logoTag

`[]string`

Used for configuring logo_match_alert.

### spec.filters.megabitsPerSecond

`[]string`

Used for configuring advanced_ddos_attack_l4_alert.

### spec.filters.newHealth

`[]string`

Used for configuring load_balancing_health_alert.

### spec.filters.newStatus

`[]string`

Used for configuring tunnel_health_event.

### spec.filters.packetsPerSecond

`[]string`

Used for configuring advanced_ddos_attack_l4_alert.

### spec.filters.poolId

`[]string`

Usage depends on specific alert type.

### spec.filters.popNames

`[]string`

Usage depends on specific alert type.

### spec.filters.product

`[]string`

Used for configuring billing_usage_alert (the product the threshold
watches).

### spec.filters.projectId

`[]string`

Used for configuring pages_event_alert.

### spec.filters.protocol

`[]string`

Used for configuring advanced_ddos_attack_l4_alert.

### spec.filters.queryTag

`[]string`

Usage depends on specific alert type.

### spec.filters.requestsPerSecond

`[]string`

Used for configuring advanced_ddos_attack_l7_alert.

### spec.filters.selectors

`[]string`

Usage depends on specific alert type.

### spec.filters.services

`[]string`

Used for configuring clickhouse_alert_fw_ent_anomaly.

### spec.filters.slo

`[]string`

Usage depends on specific alert type.

### spec.filters.status

`[]string`

Used for configuring health_check_status_notification.

### spec.filters.targetHostname

`[]string`

Used for configuring advanced_ddos_attack_l7_alert.

### spec.filters.targetIp

`[]string`

Used for configuring advanced_ddos_attack_l4_alert.

### spec.filters.targetZoneName

`[]string`

Used for configuring advanced_ddos_attack_l7_alert.

### spec.filters.trafficExclusions

`[]string`

Used for configuring traffic_anomalies_alert.

- rule: traffic_exclusions entries must be security_events

### spec.filters.tunnelId

`[]string`

Used for configuring tunnel_health_event.

### spec.filters.tunnelName

`[]string`

Usage depends on specific alert type.

### spec.filters.type

`[]string`

Usage depends on specific alert type.

### spec.filters.where

`[]string`

Usage depends on specific alert type.

### spec.filters.zones

`[]string`

Usage depends on specific alert type (the zones the policy watches).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareNotificationPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The Cloudflare-assigned UUID of the policy. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.mechanisms.webhookIds` | CloudflareNotificationWebhook | `status.outputs.webhook_id` |

## See Also

- [Overview](../README.md)
