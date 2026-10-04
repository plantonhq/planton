# GcpMonitoringUptimeCheck

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpMonitoringUptimeCheckSpec defines a Cloud Monitoring uptime check —
a probe that Google runs against a target (a public URL, a monitored
resource, a resource group, or a synthetic-monitor Cloud Function) from
multiple regions on a fixed cadence, recording availability and latency
as metrics.

An uptime check on its own only MEASURES. To be paged when the target
goes down, pair it with a GcpMonitoringAlertPolicy whose threshold
condition filters on the uptime_check_passed metric and the check's
uptime_check_id — the composition edge the `uptime_check_id` stack
output exists for.

Exactly one TARGET (monitored_resource | resource_group |
synthetic_monitor) and exactly one CHECK (http_check | tcp_check) must be
configured — mirroring the GCP API's own shape. A synthetic monitor is
the exception: it carries its own probe logic, so GCP forbids http_check
and tcp_check with it.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringUptimeCheck
metadata:
  name: my-sample-uptime-check
spec:
  # GCP project that owns the uptime check.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Shown in the Cloud Monitoring console; omit to default to
  # metadata.name.
  displayName: Website HTTPS availability

  # Max wait for the probe before recording a failure (1s-60s).
  timeout: 10s

  # Cadence: 60s, 300s (default), 600s, or 900s — the only values GCP
  # accepts.
  period: 300s

  # The canonical public-URL target: probe https://example.com/.
  monitoredResource:
    type: uptime_url
    labels:
      host: example.com

  # HTTPS probe with certificate validation — an expired certificate
  # fails the probe instead of passing silently.
  httpCheck:
    path: /
    useSsl: true
    validateSsl: true

  # Fail the probe when the body lacks the marker, even on a 200 — the
  # lying-error-page guard.
  contentMatchers:
    - content: Example Domain
      matcher: CONTAINS_STRING

  # Log failed probes to Cloud Logging for diagnosis.
  logCheckFailures: true

  # User metadata labels, merged with Planton's platform labels.
  labels:
    team: platform

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.timeout` | `string` | yes |  |  |
| `spec.period` | `string` |  |  |  |
| `spec.checkerType` | `string` |  |  |  |
| `spec.selectedRegions` | `[]string` |  |  |  |
| `spec.logCheckFailures` | `bool` |  |  |  |
| `spec.monitoredResource` | `GcpMonitoringUptimeCheckMonitoredResource` |  |  |  |
| `spec.monitoredResource.type` | `string` | yes |  |  |
| `spec.monitoredResource.labels` | `map<string, string>` | yes |  |  |
| `spec.resourceGroup` | `GcpMonitoringUptimeCheckResourceGroup` |  |  |  |
| `spec.resourceGroup.groupId` | `string` |  |  |  |
| `spec.resourceGroup.resourceType` | `string` |  |  |  |
| `spec.syntheticMonitor` | `GcpMonitoringUptimeCheckSyntheticMonitor` |  |  |  |
| `spec.syntheticMonitor.cloudFunction` | `string \| valueFrom` | yes |  | GcpCloudFunction (`status.outputs.function_id`) |
| `spec.httpCheck` | `GcpMonitoringUptimeCheckHttpCheck` |  |  |  |
| `spec.httpCheck.path` | `string` |  |  |  |
| `spec.httpCheck.port` | `int32` |  |  |  |
| `spec.httpCheck.requestMethod` | `string` |  |  |  |
| `spec.httpCheck.useSsl` | `bool` |  |  |  |
| `spec.httpCheck.validateSsl` | `bool` |  |  |  |
| `spec.httpCheck.body` | `string` |  |  |  |
| `spec.httpCheck.contentType` | `string` |  |  |  |
| `spec.httpCheck.customContentType` | `string` |  |  |  |
| `spec.httpCheck.headers` | `map<string, string>` |  |  |  |
| `spec.httpCheck.maskHeaders` | `bool` |  |  |  |
| `spec.httpCheck.authInfo` | `GcpMonitoringUptimeCheckHttpAuthInfo` |  |  |  |
| `spec.httpCheck.authInfo.username` | `string` | yes |  |  |
| `spec.httpCheck.authInfo.password` | `string` (sensitive) |  |  |  |
| `spec.httpCheck.serviceAgentAuthentication` | `GcpMonitoringUptimeCheckServiceAgentAuth` |  |  |  |
| `spec.httpCheck.serviceAgentAuthentication.type` | `string` |  |  |  |
| `spec.httpCheck.acceptedResponseStatusCodes` | `[]GcpMonitoringUptimeCheckStatusCode` |  |  |  |
| `spec.httpCheck.acceptedResponseStatusCodes[].statusClass` | `string` |  |  |  |
| `spec.httpCheck.acceptedResponseStatusCodes[].statusValue` | `int32` |  |  |  |
| `spec.httpCheck.pingConfig` | `GcpMonitoringUptimeCheckPingConfig` |  |  |  |
| `spec.httpCheck.pingConfig.pingsCount` | `int32` |  |  |  |
| `spec.tcpCheck` | `GcpMonitoringUptimeCheckTcpCheck` |  |  |  |
| `spec.tcpCheck.port` | `int32` |  |  |  |
| `spec.tcpCheck.pingConfig` | `GcpMonitoringUptimeCheckPingConfig` |  |  |  |
| `spec.tcpCheck.pingConfig.pingsCount` | `int32` |  |  |  |
| `spec.contentMatchers` | `[]GcpMonitoringUptimeCheckContentMatcher` |  |  |  |
| `spec.contentMatchers[].content` | `string` | yes |  |  |
| `spec.contentMatchers[].matcher` | `string` |  |  |  |
| `spec.contentMatchers[].jsonPathMatcher` | `GcpMonitoringUptimeCheckJsonPathMatcher` |  |  |  |
| `spec.contentMatchers[].jsonPathMatcher.jsonPath` | `string` | yes |  |  |
| `spec.contentMatchers[].jsonPathMatcher.jsonMatcher` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the uptime check. Can be a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string`

Human-friendly name shown in the Cloud Monitoring console. Defaults to
metadata.name when left empty (the GCP API requires a display name).

### spec.timeout

`string` · required

Maximum time to wait for the probe to complete before recording a
failure, as a duration in seconds (e.g. "10s"). Must be between 1s and
60s — the GCP API's own bounds.

- rule: timeout must be a whole-second duration between 1s and 60s, such as 10s or 30s
- rule: {"required":true}

### spec.period

`string`

How often the check runs (default 300s). The GCP API accepts only
60s (1 min), 300s (5 min), 600s (10 min), and 900s (15 min) — the list
the provider documents for this field.

- rule: period must be one of: 60s, 300s, 600s, 900s

### spec.checkerType

`string`

Where the probes originate (default STATIC_IP_CHECKERS — Google's
public checker fleet with published static IPs, the right choice for
internet-facing targets). VPC_CHECKERS probes private targets from
inside your VPC (requires a private checker setup and is chosen
automatically by GCP for private targets).

- rule: checker_type must be STATIC_IP_CHECKERS or VPC_CHECKERS

### spec.selectedRegions

`[]string`

Regions the check runs from (e.g. USA, EUROPE, SOUTH_AMERICA,
ASIA_PACIFIC, or the finer USA_OREGON/USA_IOWA/USA_VIRGINIA). GCP
requires enough regions to cover at least 3 checker locations; leaving
the list empty runs the check from ALL regions — the recommended
default for availability monitoring.

### spec.logCheckFailures

`bool`

Whether failed probes are written to Cloud Logging (default false).
Enable it when diagnosing flaky checks — each failure logs the probe's
observed status and latency.

### spec.monitoredResource

`GcpMonitoringUptimeCheckMonitoredResource`

A public URL or any monitored resource as the probe target. For the
common "is my site up" case, use type uptime_url with labels host
(e.g. example.com) and project_id.

### spec.monitoredResource.type

`string` · required

The monitored-resource type (see the message comment for common
values). GCP validates the type and its label schema server-side.

- rule: {"required":true}

### spec.monitoredResource.labels

`map<string, string>` · required

The labels identifying the concrete resource of that type — which
labels are required depends on the type (uptime_url needs host;
gce_instance needs instance_id and zone; all types accept project_id).

- rule: {"map":{"minPairs":"1"}}

### spec.resourceGroup

`GcpMonitoringUptimeCheckResourceGroup`

A Cloud Monitoring resource GROUP as the probe target — every member
of the group is checked. Groups are created in Cloud Monitoring
(outside this kind); reference one by its group ID.

### spec.resourceGroup.groupId

`string`

The group ID (the last segment of the group's resource name
projects/{p}/groups/{group_id}). At least one of group_id or
resource_type must be set — the provider's own constraint.

### spec.resourceGroup.resourceType

`string`

What the group's members are: INSTANCE (Compute Engine or AWS EC2) or
AWS_ELB_LOAD_BALANCER.

- rule: resource_type must be INSTANCE or AWS_ELB_LOAD_BALANCER

### spec.syntheticMonitor

`GcpMonitoringUptimeCheckSyntheticMonitor`

A synthetic monitor: GCP invokes the referenced 2nd-gen Cloud Function
on the check cadence, and the function's own assertions decide
pass/fail. The function carries the probe logic, so http_check and
tcp_check are forbidden with this target.

### spec.syntheticMonitor.cloudFunction

`string | valueFrom` · required

The fully qualified resource name of the 2nd-gen Cloud Function:
  projects/{project}/locations/{region}/functions/{name}
Can be a literal or a reference to a GcpCloudFunction resource.
Immutable: changing the target function replaces the uptime check.

- references: GcpCloudFunction (`status.outputs.function_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudFunction, name: <that resource's name>, fieldPath: status.outputs.function_id}} -- a bare string does not parse

### spec.httpCheck

`GcpMonitoringUptimeCheckHttpCheck`

HTTP(S) probe configuration — path, port, TLS, auth, expected status
codes.

- rule: custom_content_type requires content_type USER_PROVIDED, and USER_PROVIDED requires custom_content_type

### spec.httpCheck.path

`string`

URL path to probe (default "/"). A missing leading slash is prepended
by GCP.

### spec.httpCheck.port

`int32`

Port to probe. Defaults to 80 without SSL and 443 with SSL. Ports 1024
and below require use_ssl for STATIC_IP_CHECKERS.

- rule: port must be between 1 and 65535 (0 means use the protocol default: 80 for HTTP, 443 for HTTPS)

### spec.httpCheck.requestMethod

`string`

HTTP method (default GET). POST requires a content_type and typically a
body.

- rule: request_method must be GET or POST

### spec.httpCheck.useSsl

`bool`

Probe over HTTPS instead of HTTP (default false). Required when the
target only serves TLS; also required by GCP for low ports on the
public checker fleet.

### spec.httpCheck.validateSsl

`bool`

Verify the target's TLS certificate chain (default false). Only
meaningful with use_ssl — GCP ignores it on plain HTTP. Enable it in
production so an expired certificate fails the probe instead of passing
silently.

### spec.httpCheck.body

`string`

Request body for POST probes, base64-encoded (e.g.
base64("foo=bar") = "Zm9vPWJhcg=="). GCP rejects a body on GET probes.
Setting a body without a content_type fails the API's own validation.

### spec.httpCheck.contentType

`string`

How the body is declared in the request's Content-Type header:
URL_ENCODED (application/x-www-form-urlencoded) or USER_PROVIDED (the
custom_content_type value below).

- rule: content_type must be URL_ENCODED or USER_PROVIDED

### spec.httpCheck.customContentType

`string`

The Content-Type header value sent when content_type is USER_PROVIDED
(e.g. "application/json").

### spec.httpCheck.headers

`map<string, string>`

Headers to send with the probe (e.g. a Host header for name-based
virtual hosting, or an API key). At most 100 headers.

### spec.httpCheck.maskHeaders

`bool`

Hide header values in the console and API responses (default false).
GCP sets this permanently once enabled — turning it back off requires
recreating the check. Enable it whenever headers carry credentials.

### spec.httpCheck.authInfo

`GcpMonitoringUptimeCheckHttpAuthInfo`

HTTP basic authentication for the probe.

### spec.httpCheck.authInfo.username

`string` · required

Basic-auth username.

- rule: {"required":true}

### spec.httpCheck.authInfo.password

`string` · sensitive

Basic-auth password — a secret: the platform stores it as a
managed-secret reference and resolves it just-in-time at deploy.

### spec.httpCheck.serviceAgentAuthentication

`GcpMonitoringUptimeCheckServiceAgentAuth`

Authenticate the probe AS the check's Monitoring service agent using an
OIDC identity token — the keyless way to probe endpoints (Cloud Run,
Cloud Functions) that require authenticated invocations.

### spec.httpCheck.serviceAgentAuthentication.type

`string`

The authentication mechanism. OIDC_TOKEN is the only type GCP currently
supports.

- rule: type must be OIDC_TOKEN

### spec.httpCheck.acceptedResponseStatusCodes

`[]GcpMonitoringUptimeCheckStatusCode`

Response status codes that count as SUCCESS. Empty means "2xx only"
(GCP's default). Each entry is a class (e.g. STATUS_CLASS_2XX) or one
exact status_value — useful when a health endpoint deliberately returns
401/403 to anonymous probes.

- rule: set either status_class or status_value, not both

### spec.httpCheck.acceptedResponseStatusCodes[].statusClass

`string`

A status class: STATUS_CLASS_1XX, STATUS_CLASS_2XX, STATUS_CLASS_3XX,
STATUS_CLASS_4XX, STATUS_CLASS_5XX, or STATUS_CLASS_ANY.

- rule: status_class must be one of: STATUS_CLASS_1XX, STATUS_CLASS_2XX, STATUS_CLASS_3XX, STATUS_CLASS_4XX, STATUS_CLASS_5XX, STATUS_CLASS_ANY

### spec.httpCheck.acceptedResponseStatusCodes[].statusValue

`int32`

One exact HTTP status code (e.g. 401).

- rule: status_value must be a valid HTTP status code (100-599)

### spec.httpCheck.pingConfig

`GcpMonitoringUptimeCheckPingConfig`

Include ICMP pings ahead of the HTTP probe (1-3 pings), recording ping
latency alongside the check result.

### spec.httpCheck.pingConfig.pingsCount

`int32`

Number of ICMP pings to send ahead of the probe (1-3).

- rule: {"int32":{"lte":3,"gte":1}}

### spec.tcpCheck

`GcpMonitoringUptimeCheckTcpCheck`

Plain TCP connect probe — succeeds when the port accepts a connection.

### spec.tcpCheck.port

`int32`

Port to connect to. Required — TCP checks have no protocol default.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.tcpCheck.pingConfig

`GcpMonitoringUptimeCheckPingConfig`

Include ICMP pings ahead of the TCP probe (1-3 pings).

### spec.tcpCheck.pingConfig.pingsCount

`int32`

Number of ICMP pings to send ahead of the probe (1-3).

- rule: {"int32":{"lte":3,"gte":1}}

### spec.contentMatchers

`[]GcpMonitoringUptimeCheckContentMatcher`

Assertions on the response body. All matchers must pass for the probe
to pass. Applies to http_check (response body) and tcp_check (bytes
read); leave empty to assert only on status/connectivity.

- rule: json_path_matcher is required with MATCHES_JSON_PATH / NOT_MATCHES_JSON_PATH and forbidden with other matchers

### spec.contentMatchers[].content

`string` · required

The string, regex, or JSON-path expectation to test the response
against.

- rule: {"required":true}

### spec.contentMatchers[].matcher

`string`

How `content` is interpreted (default CONTAINS_STRING).

- rule: matcher must be one of: CONTAINS_STRING, NOT_CONTAINS_STRING, MATCHES_REGEX, NOT_MATCHES_REGEX, MATCHES_JSON_PATH, NOT_MATCHES_JSON_PATH

### spec.contentMatchers[].jsonPathMatcher

`GcpMonitoringUptimeCheckJsonPathMatcher`

JSON-path details for the MATCHES_JSON_PATH / NOT_MATCHES_JSON_PATH
matchers.

### spec.contentMatchers[].jsonPathMatcher.jsonPath

`string` · required

The JSONPath expression selecting the value to test (e.g.
"$.status" or "$.items[0].state").

- rule: {"required":true}

### spec.contentMatchers[].jsonPathMatcher.jsonMatcher

`string`

How the selected value is compared to `content`: EXACT_MATCH (default)
or REGEX_MATCH.

- rule: json_matcher must be EXACT_MATCH or REGEX_MATCH

### spec.labels

`map<string, string>`

User labels attached to the uptime check for organizing and identifying
it (maps to the provider's user_labels), merged with Planton's platform
labels (which win on key conflicts).

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the uptime check is deleted; alert policies that filter
               on its uptime_check_id stop receiving data
  "PREVENT" -- destroy FAILS; protects production availability
               monitoring from accidental teardown
  "ABANDON" -- the check is removed from management but keeps running
               (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_target`: configure exactly one target: monitored_resource (a URL or resource), resource_group, or synthetic_monitor
- `exactly_one_check`: configure exactly one of http_check or tcp_check (omit both only for a synthetic_monitor target)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMonitoringUptimeCheck, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.uptime_check_name` | `string` | The server-assigned resource name of the uptime check. Format: projects/{project}/uptimeCheckConfigs/{uptime_check_id} |
| `status.outputs.uptime_check_id` | `string` | The bare check ID (the last segment of uptime_check_name) — the value an alert policy's threshold filter references as metric.label.check_id to page on this check's failures. THE composition handle for check-plus-alert wiring. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.syntheticMonitor.cloudFunction` | GcpCloudFunction | `status.outputs.function_id` |

## See Also

- [Overview](../README.md)
