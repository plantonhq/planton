# GcpCloudArmorPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudArmorPolicySpec defines the configuration for a Google Cloud
Armor security policy -- a WAF and DDoS protection layer that sits
in front of backend services, load balancers, and CDN-enabled backends.

Cloud Armor policies contain a set of prioritized rules that evaluate
incoming traffic and take actions (allow, deny, rate-limit, redirect).
Rules are evaluated from highest priority (lowest number) to lowest
priority (highest number). The first matching rule's action is applied.

One kind, two scopes. Leave region empty for a GLOBAL policy, the one
a global external Application Load Balancer's backend service or a
backend bucket attaches; set region for a REGIONAL policy, the one a
regional backend service (regional external or internal ALB, or a
passthrough Network Load Balancer) attaches. Scopes must match: a
regional backend service accepts only a regional policy. A policy
cannot move between scopes.

Policy types, and where each scope allows them:

  - **CLOUD_ARMOR** (default; both scopes): Backend security policies
    for HTTP(S) load balancers. Full WAF, rate limiting; on the global
    scope also redirect, header injection, reCAPTCHA, and Adaptive
    Protection.

  - **CLOUD_ARMOR_EDGE** (both scopes): Edge security policies for Cloud
    CDN and backend buckets. Limited to IP-based and geo-based rules.

  - **CLOUD_ARMOR_INTERNAL_SERVICE** (global only): Policies for internal
    Traffic Director services. Limited feature set.

  - **CLOUD_ARMOR_NETWORK** (regional only): packet-level policies for
    passthrough Network Load Balancers, protocol forwarding, and VMs with
    public IPs -- network DDoS protection (ddos_protection_config),
    custom packet fields (user_defined_fields), and L3/L4 rules
    (rules[].network_match).

The policy type is immutable after creation (ForceNew).

The regional collection carries no labels, Adaptive Protection,
reCAPTCHA options, request-body inspection size, redirect action,
header injection, or reCAPTCHA token options; each is rejected when
region is set so a manifest fails before the API sees it.

The default rule contract: every Cloud Armor policy carries a default
rule at priority 2147483647. Creating a policy with NO rules lets the
API add a default "allow all" rule automatically; providing ANY rules
requires the set to include the priority-2147483647 default explicitly
(the API rejects rule sets without it, so this is enforced pre-deploy).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudArmorPolicy
metadata:
  name: test-waf-policy
spec:
  projectId:
    value: my-gcp-project
  policyName: test-waf-policy
  description: Test WAF policy with rate limiting and DDoS protection
  type: CLOUD_ARMOR
  adaptiveProtectionConfig:
    enableLayer7DdosDefense: true
    ruleVisibility: STANDARD
  advancedOptionsConfig:
    jsonParsing: STANDARD
    logLevel: VERBOSE
    # Inspect deeper request bodies than the 8KB default.
    requestBodyInspectionSize: 16KB
  # User labels, merged with the platform labels.
  labels:
    team: security
  # Ephemeral test fixture: delete the policy on destroy.
  deletionPolicy: DELETE
  rules:
    - action: allow
      priority: 100
      description: Allow internal traffic
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - 10.0.0.0/8
          - 172.16.0.0/12
    - action: deny(403)
      priority: 200
      description: Block known bad regions
      match:
        expression: "origin.region_code == 'CN' || origin.region_code == 'RU'"
    - action: throttle
      priority: 300
      description: Rate limit per IP
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
      rateLimitOptions:
        conformAction: allow
        exceedAction: deny(429)
        enforceOnKey: IP
        rateLimitThreshold:
          count: 100
          intervalSec: 60
    - action: deny(403)
      priority: 2147483647
      description: Default deny rule
      match:
        versionedExpr: SRC_IPS_V1
        srcIpRanges:
          - "*"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.policyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.adaptiveProtectionConfig` | `GcpCloudArmorAdaptiveProtectionConfig` |  |  |  |
| `spec.adaptiveProtectionConfig.enableLayer7DdosDefense` | `bool` |  |  |  |
| `spec.adaptiveProtectionConfig.ruleVisibility` | `string` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs` | `[]GcpCloudArmorThresholdConfig` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].name` | `string` | yes |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployConfidenceThreshold` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployImpactedBaselineThreshold` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployLoadThreshold` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployExpirationSec` | `int32` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].detectionAbsoluteQps` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].detectionLoadThreshold` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].detectionRelativeToBaselineQps` | `double` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs` | `[]GcpCloudArmorTrafficGranularityConfig` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].type` | `string` | yes |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].value` | `string` |  |  |  |
| `spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].enableEachUniqueValue` | `bool` |  |  |  |
| `spec.advancedOptionsConfig` | `GcpCloudArmorAdvancedOptionsConfig` |  |  |  |
| `spec.advancedOptionsConfig.jsonParsing` | `string` |  |  |  |
| `spec.advancedOptionsConfig.logLevel` | `string` |  |  |  |
| `spec.advancedOptionsConfig.userIpRequestHeaders` | `[]string` |  |  |  |
| `spec.advancedOptionsConfig.jsonCustomConfig` | `GcpCloudArmorJsonCustomConfig` |  |  |  |
| `spec.advancedOptionsConfig.jsonCustomConfig.contentTypes` | `[]string` | yes |  |  |
| `spec.advancedOptionsConfig.requestBodyInspectionSize` | `string` |  |  |  |
| `spec.recaptchaOptionsConfig` | `GcpCloudArmorRecaptchaOptionsConfig` |  |  |  |
| `spec.recaptchaOptionsConfig.redirectSiteKey` | `string` | yes |  |  |
| `spec.rules` | `[]GcpCloudArmorRule` |  |  |  |
| `spec.rules[].action` | `string` | yes |  |  |
| `spec.rules[].priority` | `int32` | yes |  |  |
| `spec.rules[].match` | `GcpCloudArmorRuleMatch` |  |  |  |
| `spec.rules[].match.versionedExpr` | `string` |  |  |  |
| `spec.rules[].match.srcIpRanges` | `[]string` |  |  |  |
| `spec.rules[].match.expression` | `string` |  |  |  |
| `spec.rules[].match.exprOptions` | `GcpCloudArmorRecaptchaOptions` |  |  |  |
| `spec.rules[].match.exprOptions.actionTokenSiteKeys` | `[]string` |  |  |  |
| `spec.rules[].match.exprOptions.sessionTokenSiteKeys` | `[]string` |  |  |  |
| `spec.rules[].networkMatch` | `GcpCloudArmorNetworkMatch` |  |  |  |
| `spec.rules[].networkMatch.srcIpRanges` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.destIpRanges` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.ipProtocols` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.srcPorts` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.destPorts` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.srcRegionCodes` | `[]string` |  |  |  |
| `spec.rules[].networkMatch.srcAsns` | `[]int64` |  |  |  |
| `spec.rules[].networkMatch.userDefinedFields` | `[]GcpCloudArmorNetworkMatchUserDefinedField` |  |  |  |
| `spec.rules[].networkMatch.userDefinedFields[].name` | `string` | yes |  |  |
| `spec.rules[].networkMatch.userDefinedFields[].values` | `[]string` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].preview` | `bool` |  |  |  |
| `spec.rules[].rateLimitOptions` | `GcpCloudArmorRateLimitOptions` |  |  |  |
| `spec.rules[].rateLimitOptions.conformAction` | `string` | yes |  |  |
| `spec.rules[].rateLimitOptions.exceedAction` | `string` | yes |  |  |
| `spec.rules[].rateLimitOptions.enforceOnKey` | `string` |  |  |  |
| `spec.rules[].rateLimitOptions.enforceOnKeyName` | `string` |  |  |  |
| `spec.rules[].rateLimitOptions.enforceOnKeyConfigs` | `[]GcpCloudArmorEnforceOnKeyConfig` |  |  |  |
| `spec.rules[].rateLimitOptions.enforceOnKeyConfigs[].enforceOnKeyType` | `string` | yes |  |  |
| `spec.rules[].rateLimitOptions.enforceOnKeyConfigs[].enforceOnKeyName` | `string` |  |  |  |
| `spec.rules[].rateLimitOptions.rateLimitThreshold` | `GcpCloudArmorRateThreshold` | yes |  |  |
| `spec.rules[].rateLimitOptions.rateLimitThreshold.count` | `int32` | yes |  |  |
| `spec.rules[].rateLimitOptions.rateLimitThreshold.intervalSec` | `int32` | yes |  |  |
| `spec.rules[].rateLimitOptions.banThreshold` | `GcpCloudArmorRateThreshold` |  |  |  |
| `spec.rules[].rateLimitOptions.banThreshold.count` | `int32` | yes |  |  |
| `spec.rules[].rateLimitOptions.banThreshold.intervalSec` | `int32` | yes |  |  |
| `spec.rules[].rateLimitOptions.banDurationSec` | `int32` |  |  |  |
| `spec.rules[].rateLimitOptions.exceedRedirectOptions` | `GcpCloudArmorRedirectConfig` |  |  |  |
| `spec.rules[].rateLimitOptions.exceedRedirectOptions.type` | `string` | yes |  |  |
| `spec.rules[].rateLimitOptions.exceedRedirectOptions.target` | `string` |  |  |  |
| `spec.rules[].redirectOptions` | `GcpCloudArmorRedirectConfig` |  |  |  |
| `spec.rules[].redirectOptions.type` | `string` | yes |  |  |
| `spec.rules[].redirectOptions.target` | `string` |  |  |  |
| `spec.rules[].headerAction` | `GcpCloudArmorHeaderAction` |  |  |  |
| `spec.rules[].headerAction.requestHeadersToAdds` | `[]GcpCloudArmorRequestHeader` | yes |  |  |
| `spec.rules[].headerAction.requestHeadersToAdds[].headerName` | `string` | yes |  |  |
| `spec.rules[].headerAction.requestHeadersToAdds[].headerValue` | `string` |  |  |  |
| `spec.rules[].preconfiguredWafConfig` | `GcpCloudArmorPreconfiguredWafConfig` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions` | `[]GcpCloudArmorWafExclusion` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].targetRuleSet` | `string` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].targetRuleIds` | `[]string` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders` | `[]GcpCloudArmorWafExclusionFieldParams` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders[].operator` | `string` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders[].value` | `string` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies` | `[]GcpCloudArmorWafExclusionFieldParams` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies[].operator` | `string` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies[].value` | `string` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestUris` | `[]GcpCloudArmorWafExclusionFieldParams` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestUris[].operator` | `string` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestUris[].value` | `string` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams` | `[]GcpCloudArmorWafExclusionFieldParams` |  |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams[].operator` | `string` | yes |  |  |
| `spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams[].value` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.ddosProtectionConfig` | `GcpCloudArmorDdosProtectionConfig` |  |  |  |
| `spec.ddosProtectionConfig.ddosProtection` | `string` | yes |  |  |
| `spec.userDefinedFields` | `[]GcpCloudArmorUserDefinedField` |  |  |  |
| `spec.userDefinedFields[].name` | `string` |  |  |  |
| `spec.userDefinedFields[].base` | `string` | yes |  |  |
| `spec.userDefinedFields[].offset` | `int32` |  |  |  |
| `spec.userDefinedFields[].size` | `int32` |  |  |  |
| `spec.userDefinedFields[].mask` | `string` |  |  |  |
| `spec.networkEdgeSecurityService` | `GcpCloudArmorNetworkEdgeSecurityService` |  |  |  |
| `spec.networkEdgeSecurityService.name` | `string` |  |  |  |
| `spec.networkEdgeSecurityService.description` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the security policy will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.policyName

`string`

Name of the security policy in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a letter or number.
If not specified, defaults to metadata.name.

- rule: policy_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.description

`string`

Description of the security policy. Max 2048 characters.

- rule: {"string":{"maxLen":"2048"}}

### spec.type

`string`

Policy type. Determines which features are available and where
the policy can be attached. CLOUD_ARMOR (default) and CLOUD_ARMOR_EDGE
exist on both scopes; CLOUD_ARMOR_INTERNAL_SERVICE only globally;
CLOUD_ARMOR_NETWORK only regionally (set region).

Immutable after creation (ForceNew).

- rule: type must be CLOUD_ARMOR, CLOUD_ARMOR_EDGE, CLOUD_ARMOR_INTERNAL_SERVICE, or CLOUD_ARMOR_NETWORK

### spec.adaptiveProtectionConfig

`GcpCloudArmorAdaptiveProtectionConfig`

Adaptive Protection configuration for automatic Layer 7 DDoS detection.
A global-policy lever (rejected when region is set).

- rule: threshold_configs require enable_layer_7_ddos_defense to be true

### spec.adaptiveProtectionConfig.enableLayer7DdosDefense

`bool`

Enable Cloud Armor Adaptive Protection for Layer 7 DDoS defense.
When true, traffic anomalies are detected and alerts are generated.

### spec.adaptiveProtectionConfig.ruleVisibility

`string`

Rule visibility mode for auto-generated adaptive protection rules.
"STANDARD" (default) creates rules visible to all policy viewers.
"PREMIUM" requires Cloud Armor Managed Protection Plus.

- rule: rule_visibility must be STANDARD or PREMIUM

### spec.adaptiveProtectionConfig.thresholdConfigs

`[]GcpCloudArmorThresholdConfig`

Per-granularity detection and auto-deploy threshold overrides.
Requires enable_layer_7_ddos_defense.

### spec.adaptiveProtectionConfig.thresholdConfigs[].name

`string` · required

Name of the config. Must be 1-63 characters, RFC1035-compliant, and
unique within the policy.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployConfidenceThreshold

`double` · optional (explicit presence)

Confidence threshold (0.0-1.0) an attack signature must reach before
auto-deploying a mitigation rule.

- rule: {"double":{"lte":1,"gte":0}}

### spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployImpactedBaselineThreshold

`double` · optional (explicit presence)

Maximum share (0.0-1.0) of baseline (good) traffic the auto-deployed
mitigation may impact.

- rule: {"double":{"lte":1,"gte":0}}

### spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployLoadThreshold

`double` · optional (explicit presence)

Load threshold above which auto-deploy considers the backend under
attack.

### spec.adaptiveProtectionConfig.thresholdConfigs[].autoDeployExpirationSec

`int32` · optional (explicit presence)

Lifetime in seconds of an auto-deployed mitigation rule.

- rule: {"int32":{"gte":0}}

### spec.adaptiveProtectionConfig.thresholdConfigs[].detectionAbsoluteQps

`double` · optional (explicit presence)

Detection: absolute queries-per-second considered anomalous.

### spec.adaptiveProtectionConfig.thresholdConfigs[].detectionLoadThreshold

`double` · optional (explicit presence)

Detection: load threshold relative to backend capacity.

### spec.adaptiveProtectionConfig.thresholdConfigs[].detectionRelativeToBaselineQps

`double` · optional (explicit presence)

Detection: QPS relative to the learned baseline considered anomalous.

### spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs

`[]GcpCloudArmorTrafficGranularityConfig`

Granular traffic units this threshold config applies to.

- rule: enable_each_unique_value can only be true when value is empty

### spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].type

`string` · required

Type of granularity: "HTTP_HEADER_HOST" or "HTTP_PATH".

- rule: type must be HTTP_HEADER_HOST or HTTP_PATH
- rule: {"required":true}

### spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].value

`string`

A specific value of the configured type that constitutes a traffic
unit (e.g. one Host name). Leave empty with enable_each_unique_value
to treat every unique value as its own unit.

### spec.adaptiveProtectionConfig.thresholdConfigs[].trafficGranularityConfigs[].enableEachUniqueValue

`bool`

When true, traffic matching EACH unique value of the type is a
separate traffic unit. Only valid when value is empty.

### spec.advancedOptionsConfig

`GcpCloudArmorAdvancedOptionsConfig`

Advanced policy-level options: JSON parsing, logging, IP resolution
(both scopes); request_body_inspection_size (global only).

- rule: json_custom_config applies only when json_parsing is STANDARD or STANDARD_WITH_GRAPHQL

### spec.advancedOptionsConfig.jsonParsing

`string`

JSON parsing mode for request body inspection.
"DISABLED" (default): No JSON parsing.
"STANDARD": Parse JSON bodies for WAF rule evaluation.
"STANDARD_WITH_GRAPHQL": Parse JSON and GraphQL bodies.

- rule: json_parsing must be DISABLED, STANDARD, or STANDARD_WITH_GRAPHQL

### spec.advancedOptionsConfig.logLevel

`string`

Logging verbosity.
"NORMAL" (default): Standard logging.
"VERBOSE": Detailed logging including matched rule and request details.

- rule: log_level must be NORMAL or VERBOSE

### spec.advancedOptionsConfig.userIpRequestHeaders

`[]string`

Custom headers to check for the true client IP address.
Used when traffic passes through a CDN or reverse proxy that
sets the client IP in a custom header. If empty, GCP uses the
connection source IP.

### spec.advancedOptionsConfig.jsonCustomConfig

`GcpCloudArmorJsonCustomConfig`

Additional Content-Type values to parse as JSON. Only meaningful
when json_parsing is STANDARD or STANDARD_WITH_GRAPHQL.

### spec.advancedOptionsConfig.jsonCustomConfig.contentTypes

`[]string` · required

Custom Content-Type header values to parse as JSON
(e.g. "application/vnd.api+json").

- rule: {"repeated":{"minItems":"1"}}

### spec.advancedOptionsConfig.requestBodyInspectionSize

`string`

How much of each request body the WAF inspects. One of "8KB", "16KB",
"32KB", "48KB", "64KB" (provider default 8KB). Larger sizes catch
attacks buried deeper in POST bodies at higher processing cost; bodies
beyond the limit pass uninspected. Mutable.

- rule: request_body_inspection_size must be one of: 8KB, 16KB, 32KB, 48KB, 64KB

### spec.recaptchaOptionsConfig

`GcpCloudArmorRecaptchaOptionsConfig`

Policy-level reCAPTCHA site key for GOOGLE_RECAPTCHA redirects.
A global-policy lever (rejected when region is set).

### spec.recaptchaOptionsConfig.redirectSiteKey

`string` · required

The reCAPTCHA site key, created from the reCAPTCHA API. The user is
responsible for the key's validity.

- rule: {"required":true}

### spec.rules

`[]GcpCloudArmorRule`

Security rules. Rules are evaluated in priority order (lowest number
first). Each rule matches traffic and applies an action.

Leave empty to get the API's automatic default "allow all" rule; a
non-empty set must include the priority-2147483647 default rule.

- rule: rate_limit_options is required for throttle/rate_based_ban actions and must not be set otherwise
- rule: redirect_options is required for the redirect action and must not be set otherwise
- rule: a rule matches through exactly one arm: match (HTTP request attributes) or network_match (packet headers, CLOUD_ARMOR_NETWORK policies only)

### spec.rules[].action

`string` · required

Action to take when the rule matches.
- "allow": Permit the request
- "deny(403)": Block with 403 Forbidden
- "deny(404)": Block with 404 Not Found
- "deny(502)": Block with 502 Bad Gateway
- "redirect": Redirect to a configured target (requires redirect_options;
  GLOBAL CLOUD_ARMOR policies only -- rejected when region is set)
- "throttle": Rate-limit the traffic (requires rate_limit_options)
- "rate_based_ban": Rate-limit then ban (requires rate_limit_options)

- rule: action must be one of: allow, deny(403), deny(404), deny(502), redirect, throttle, rate_based_ban
- rule: {"required":true}

### spec.rules[].priority

`int32` · required · optional (explicit presence)

Rule priority. Lower values are evaluated first; 0 is the HIGHEST
priority Google accepts and is a legal value. Range: 0 to 2147483647.
Each rule must have a unique priority. Priority 2147483647 is the
default rule (match "*"). Leave gaps (100, 200, ...) so a rule can be
slotted in later without renumbering.

- rule: {"required":true,"int32":{"lte":2147483647,"gte":0}}

### spec.rules[].match

`GcpCloudArmorRuleMatch`

HTTP-level traffic-matching condition (source IP ranges or a CEL
expression over the request). Exactly one of match / network_match
is set per rule.

- rule: exactly one of versioned_expr (with src_ip_ranges) or expression must be set
- rule: src_ip_ranges is required when versioned_expr is set
- rule: expr_options applies only to CEL-expression matches (set expression)

### spec.rules[].match.versionedExpr

`string`

Predefined match expression. The only supported value is "SRC_IPS_V1",
which matches traffic based on source IP address ranges.
When set, src_ip_ranges must also be provided.

- rule: versioned_expr must be SRC_IPS_V1 if set

### spec.rules[].match.srcIpRanges

`[]string`

Source IP CIDR ranges to match against.
Required when versioned_expr is "SRC_IPS_V1". Max 10 ranges per rule.
Use "*" to match all IP addresses.
Examples: ["192.168.1.0/24", "10.0.0.0/8"] or ["*"]

- rule: {"repeated":{"maxItems":"10"}}

### spec.rules[].match.expression

`string`

CEL expression for advanced matching. Supports request attributes
such as origin.region_code, request.headers['X-Custom'], request.path,
inIpRange(origin.ip, '1.2.3.0/24'), and more.
Mutually exclusive with versioned_expr.
Example: "origin.region_code == 'US'" or "request.path.matches('/api/.*')"

### spec.rules[].match.exprOptions

`GcpCloudArmorRecaptchaOptions`

reCAPTCHA site-key options for expressions that evaluate reCAPTCHA
tokens. Only meaningful together with expression.

- rule: at least one of action_token_site_keys or session_token_site_keys must be set

### spec.rules[].match.exprOptions.actionTokenSiteKeys

`[]string`

Site keys used to validate reCAPTCHA action-tokens.

### spec.rules[].match.exprOptions.sessionTokenSiteKeys

`[]string`

Site keys used to validate reCAPTCHA session-tokens.

### spec.rules[].networkMatch

`GcpCloudArmorNetworkMatch`

Packet-level (L3/L4) match condition for a regional
CLOUD_ARMOR_NETWORK policy. Exactly one of match / network_match is
set per rule; rejected on any other policy type or scope.

### spec.rules[].networkMatch.srcIpRanges

`[]string`

Source IPv4/IPv6 addresses or CIDR prefixes, in standard text format.

### spec.rules[].networkMatch.destIpRanges

`[]string`

Destination IPv4/IPv6 addresses or CIDR prefixes, in standard text
format.

### spec.rules[].networkMatch.ipProtocols

`[]string`

IPv4 protocol / IPv6 next header (after extension headers). Each
element is an 8-bit unsigned decimal number (e.g. "6"), a range (e.g.
"253-254"), or one of the names "tcp", "udp", "icmp", "esp", "ah",
"ipip", "sctp".

### spec.rules[].networkMatch.srcPorts

`[]string`

Source port numbers for TCP/UDP/SCTP. Each element is a 16-bit unsigned
decimal number (e.g. "80") or a range (e.g. "0-1023").

### spec.rules[].networkMatch.destPorts

`[]string`

Destination port numbers for TCP/UDP/SCTP. Each element is a 16-bit
unsigned decimal number (e.g. "80") or a range (e.g. "0-1023").

### spec.rules[].networkMatch.srcRegionCodes

`[]string`

Two-letter ISO 3166-1 alpha-2 country codes associated with the
source IP address (e.g. "US", "DE").

### spec.rules[].networkMatch.srcAsns

`[]int64`

BGP Autonomous System Numbers associated with the source IP address
(e.g. 15169 for Google). 32-bit ASNs exceed the signed 32-bit range,
so the type is 64-bit.

### spec.rules[].networkMatch.userDefinedFields

`[]GcpCloudArmorNetworkMatchUserDefinedField`

Matches on the policy's user-defined packet fields, each naming a field
from user_defined_fields and listing the values that match it.

### spec.rules[].networkMatch.userDefinedFields[].name

`string` · required

Name of the user-defined field, exactly as given in the policy's
user_defined_fields definition.

- rule: {"required":true}

### spec.rules[].networkMatch.userDefinedFields[].values

`[]string` · required

Matching values of the field. Each element is a 32-bit unsigned decimal
or hexadecimal (0x-prefixed) number, e.g. "64" or "0x8F00", or a range
such as "0x400-0x7ff". Any listed value matches.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].description

`string`

Human-readable description of the rule (max 64 characters).

- rule: {"string":{"maxLen":"64"}}

### spec.rules[].preview

`bool`

If true, the rule is in preview mode: matched traffic is logged but
the action is not enforced. Use preview to test rules before enabling.

### spec.rules[].rateLimitOptions

`GcpCloudArmorRateLimitOptions`

Rate limit configuration. Required when action is "throttle" or
"rate_based_ban".

- rule: enforce_on_key and enforce_on_key_configs are mutually exclusive
- rule: enforce_on_key_name is required for HTTP_HEADER/HTTP_COOKIE and must not be set otherwise
- rule: exceed_redirect_options is required when exceed_action is redirect and must not be set otherwise

### spec.rules[].rateLimitOptions.conformAction

`string` · required

Action to take when traffic is below the threshold. Must be "allow".

- rule: conform_action must be 'allow'
- rule: {"required":true}

### spec.rules[].rateLimitOptions.exceedAction

`string` · required

Action to take when traffic exceeds the threshold.
Valid values: "redirect", "deny(403)", "deny(404)", "deny(429)", "deny(502)".

- rule: exceed_action must be one of: redirect, deny(403), deny(404), deny(429), deny(502)
- rule: {"required":true}

### spec.rules[].rateLimitOptions.enforceOnKey

`string`

Single key on which to enforce the rate limit. Determines how requests
are grouped for counting. If empty (and no enforce_on_key_configs),
defaults to "ALL" (single counter for all matched traffic).

Common values:
  - "ALL": Single counter for all traffic
  - "IP": Per source IP
  - "HTTP_HEADER": Per value of a specific header (set enforce_on_key_name)
  - "XFF_IP": Per IP from X-Forwarded-For header
  - "HTTP_COOKIE": Per cookie value (set enforce_on_key_name)
  - "HTTP_PATH": Per URL path
  - "SNI": Per TLS Server Name Indication
  - "REGION_CODE": Per client country/region
Mutually exclusive with enforce_on_key_configs.

- rule: enforce_on_key must be one of: ALL, IP, HTTP_HEADER, XFF_IP, HTTP_COOKIE, HTTP_PATH, SNI, REGION_CODE, TLS_JA3_FINGERPRINT, TLS_JA4_FINGERPRINT, USER_IP

### spec.rules[].rateLimitOptions.enforceOnKeyName

`string`

Name of the HTTP header or cookie when enforce_on_key is
HTTP_HEADER or HTTP_COOKIE.

### spec.rules[].rateLimitOptions.enforceOnKeyConfigs

`[]GcpCloudArmorEnforceOnKeyConfig`

Composite rate-limit key: the listed components' values are
concatenated to form the key requests are counted against
(e.g. IP + HTTP_PATH limits each client per path).
Mutually exclusive with enforce_on_key.

- rule: enforce_on_key_name is required for HTTP_HEADER/HTTP_COOKIE key types and must not be set otherwise

### spec.rules[].rateLimitOptions.enforceOnKeyConfigs[].enforceOnKeyType

`string` · required

The key type this component contributes.

- rule: enforce_on_key_type must be one of: ALL, IP, HTTP_HEADER, XFF_IP, HTTP_COOKIE, HTTP_PATH, SNI, REGION_CODE, TLS_JA3_FINGERPRINT, TLS_JA4_FINGERPRINT, USER_IP
- rule: {"required":true}

### spec.rules[].rateLimitOptions.enforceOnKeyConfigs[].enforceOnKeyName

`string`

Name of the HTTP header or cookie when enforce_on_key_type is
HTTP_HEADER or HTTP_COOKIE.

### spec.rules[].rateLimitOptions.rateLimitThreshold

`GcpCloudArmorRateThreshold` · required

Rate limit threshold: when the request count exceeds this value
within the interval, the exceed_action is applied.

- rule: {"required":true}

### spec.rules[].rateLimitOptions.rateLimitThreshold.count

`int32` · required

Number of requests that triggers the threshold.

- rule: {"required":true,"int32":{"gt":0}}

### spec.rules[].rateLimitOptions.rateLimitThreshold.intervalSec

`int32` · required

Window of time in seconds over which the count is measured.
The API accepts a fixed set of windows.

- rule: interval_sec must be one of: 60, 120, 180, 240, 300, 600, 900, 1200
- rule: {"required":true}

### spec.rules[].rateLimitOptions.banThreshold

`GcpCloudArmorRateThreshold`

Ban threshold for rate_based_ban actions. When traffic exceeds
this threshold after already exceeding the rate_limit_threshold,
the source is banned entirely.

### spec.rules[].rateLimitOptions.banThreshold.count

`int32` · required

Number of requests that triggers the threshold.

- rule: {"required":true,"int32":{"gt":0}}

### spec.rules[].rateLimitOptions.banThreshold.intervalSec

`int32` · required

Window of time in seconds over which the count is measured.
The API accepts a fixed set of windows.

- rule: interval_sec must be one of: 60, 120, 180, 240, 300, 600, 900, 1200
- rule: {"required":true}

### spec.rules[].rateLimitOptions.banDurationSec

`int32`

Duration of the ban in seconds when using rate_based_ban.
Range: 60 to 86400 (1 minute to 24 hours).

- rule: ban_duration_sec must be between 60 and 86400 seconds

### spec.rules[].rateLimitOptions.exceedRedirectOptions

`GcpCloudArmorRedirectConfig`

Redirect configuration when exceed_action is "redirect".

- rule: target is required for EXTERNAL_302 and must not be set for GOOGLE_RECAPTCHA

### spec.rules[].rateLimitOptions.exceedRedirectOptions.type

`string` · required

Redirect type. EXTERNAL_302 sends a 302 redirect to the target URL.
GOOGLE_RECAPTCHA redirects to a Google reCAPTCHA challenge page
(customize the site key with the policy-level recaptcha_options_config).

- rule: type must be EXTERNAL_302 or GOOGLE_RECAPTCHA
- rule: {"required":true}

### spec.rules[].rateLimitOptions.exceedRedirectOptions.target

`string`

Target URL for EXTERNAL_302 redirects. Required when type is
EXTERNAL_302; must not be set when type is GOOGLE_RECAPTCHA.

### spec.rules[].redirectOptions

`GcpCloudArmorRedirectConfig`

Redirect configuration. Required when action is "redirect". A
global-policy lever: the regional collection has no redirect action.

- rule: target is required for EXTERNAL_302 and must not be set for GOOGLE_RECAPTCHA

### spec.rules[].redirectOptions.type

`string` · required

Redirect type. EXTERNAL_302 sends a 302 redirect to the target URL.
GOOGLE_RECAPTCHA redirects to a Google reCAPTCHA challenge page
(customize the site key with the policy-level recaptcha_options_config).

- rule: type must be EXTERNAL_302 or GOOGLE_RECAPTCHA
- rule: {"required":true}

### spec.rules[].redirectOptions.target

`string`

Target URL for EXTERNAL_302 redirects. Required when type is
EXTERNAL_302; must not be set when type is GOOGLE_RECAPTCHA.

### spec.rules[].headerAction

`GcpCloudArmorHeaderAction`

Custom headers to inject into matching requests before forwarding
to the backend. Only supported for GLOBAL CLOUD_ARMOR type policies
(rejected when region is set).

### spec.rules[].headerAction.requestHeadersToAdds

`[]GcpCloudArmorRequestHeader` · required

Headers to add to matching requests.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].headerAction.requestHeadersToAdds[].headerName

`string` · required

HTTP header name.

- rule: {"required":true}

### spec.rules[].headerAction.requestHeadersToAdds[].headerValue

`string`

HTTP header value. If the header already exists, it is overwritten.

### spec.rules[].preconfiguredWafConfig

`GcpCloudArmorPreconfiguredWafConfig`

Preconfigured WAF rule exclusions. Use this to carve out exceptions
for specific request fields that trigger false positives in WAF rules.
Only supported for CLOUD_ARMOR type policies.

### spec.rules[].preconfiguredWafConfig.exclusions

`[]GcpCloudArmorWafExclusion` · required

Exclusions from preconfigured WAF rules.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].preconfiguredWafConfig.exclusions[].targetRuleSet

`string` · required

Target WAF rule set to exclude from. Uses the ModSecurity rule set
identifiers such as "sqli-v33-stable", "xss-v33-stable",
"rce-v33-stable", "lfi-v33-stable", etc.

- rule: {"required":true}

### spec.rules[].preconfiguredWafConfig.exclusions[].targetRuleIds

`[]string`

Specific rule IDs within the rule set to exclude. If empty,
the exclusion applies to all rules in the set.

### spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders

`[]GcpCloudArmorWafExclusionFieldParams`

Request headers to exclude from WAF evaluation.

- rule: value is required for all operators except EQUALS_ANY, which takes no value

### spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders[].operator

`string` · required

Comparison operator for matching the field.

- rule: operator must be one of: EQUALS, STARTS_WITH, ENDS_WITH, CONTAINS, EQUALS_ANY
- rule: {"required":true}

### spec.rules[].preconfiguredWafConfig.exclusions[].requestHeaders[].value

`string`

Value to match against. Required unless operator is EQUALS_ANY
(which matches any value for the given field).

### spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies

`[]GcpCloudArmorWafExclusionFieldParams`

Request cookies to exclude from WAF evaluation.

- rule: value is required for all operators except EQUALS_ANY, which takes no value

### spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies[].operator

`string` · required

Comparison operator for matching the field.

- rule: operator must be one of: EQUALS, STARTS_WITH, ENDS_WITH, CONTAINS, EQUALS_ANY
- rule: {"required":true}

### spec.rules[].preconfiguredWafConfig.exclusions[].requestCookies[].value

`string`

Value to match against. Required unless operator is EQUALS_ANY
(which matches any value for the given field).

### spec.rules[].preconfiguredWafConfig.exclusions[].requestUris

`[]GcpCloudArmorWafExclusionFieldParams`

Request URIs to exclude from WAF evaluation.

- rule: value is required for all operators except EQUALS_ANY, which takes no value

### spec.rules[].preconfiguredWafConfig.exclusions[].requestUris[].operator

`string` · required

Comparison operator for matching the field.

- rule: operator must be one of: EQUALS, STARTS_WITH, ENDS_WITH, CONTAINS, EQUALS_ANY
- rule: {"required":true}

### spec.rules[].preconfiguredWafConfig.exclusions[].requestUris[].value

`string`

Value to match against. Required unless operator is EQUALS_ANY
(which matches any value for the given field).

### spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams

`[]GcpCloudArmorWafExclusionFieldParams`

Request query parameters to exclude from WAF evaluation.

- rule: value is required for all operators except EQUALS_ANY, which takes no value

### spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams[].operator

`string` · required

Comparison operator for matching the field.

- rule: operator must be one of: EQUALS, STARTS_WITH, ENDS_WITH, CONTAINS, EQUALS_ANY
- rule: {"required":true}

### spec.rules[].preconfiguredWafConfig.exclusions[].requestQueryParams[].value

`string`

Value to match against. Required unless operator is EQUALS_ANY
(which matches any value for the given field).

### spec.labels

`map<string, string>`

User labels attached to the security policy, merged with Planton's
platform labels (which win on key conflicts). Mutable. A global-policy
lever: the regional collection carries no labels at all (rejected when
region is set; a regional policy also receives no platform labels).

### spec.deletionPolicy

`string`

Deletion policy for the security policy — what happens on destroy:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy and every rule in it are deleted (GCP
               refuses while a backend service still attaches it);
               the traffic it filtered flows unfiltered once detached
  "PREVENT" -- destroy FAILS; protects the WAF/DDoS shield in front
               of production backends
  "ABANDON" -- the policy is removed from management but keeps
               enforcing in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.region

`string`

The scope selector. Empty builds a GLOBAL security policy (attached by
a global backend service or a backend bucket in front of a global
external Application Load Balancer, or by Cloud CDN); a region name
such as us-central1 builds a REGIONAL one (attached by a regional
backend service: the regional external and internal ALBs, and -- as a
CLOUD_ARMOR_NETWORK policy -- the passthrough Network Load Balancers,
protocol forwarding, and public-IP VMs in that region). Scopes must
match what attaches the policy. The global-only levers (labels,
adaptive_protection_config, recaptcha_options_config,
request_body_inspection_size, the CLOUD_ARMOR_INTERNAL_SERVICE type,
header_action, redirect, expr_options) are rejected when region is set;
the network-policy levers (CLOUD_ARMOR_NETWORK, ddos_protection_config,
user_defined_fields, network_match) are rejected when it is empty.
Immutable: a policy cannot move between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global security policy

### spec.ddosProtectionConfig

`GcpCloudArmorDdosProtectionConfig`

Network DDoS protection level for a regional CLOUD_ARMOR_NETWORK
policy: STANDARD is free and always on; ADVANCED and ADVANCED_PREVIEW
need Cloud Armor Enterprise and the region enrolled through
network_edge_security_service.

### spec.ddosProtectionConfig.ddosProtection

`string` · required

Protection level:
- STANDARD: basic always-on protection, included with the load
  balancer -- no subscription
- ADVANCED: the additional network-layer protections of Cloud Armor
  Enterprise (Managed Protection Plus); the project must be enrolled
- ADVANCED_PREVIEW: ADVANCED in preview mode -- Google logs what it
  would mitigate without mitigating; use it to observe before enforcing
ADVANCED and ADVANCED_PREVIEW take effect only once the region is
enrolled through network_edge_security_service.

- rule: ddos_protection must be one of: STANDARD, ADVANCED, ADVANCED_PREVIEW
- rule: {"required":true}

### spec.userDefinedFields

`[]GcpCloudArmorUserDefinedField`

Custom packet fields (up to 4 bytes at a fixed header offset) that a
regional CLOUD_ARMOR_NETWORK policy's rules match through
network_match.user_defined_fields. Names must be unique.

### spec.userDefinedFields[].name

`string`

Name of this field. Must be unique within the policy; rules name it in
network_match.user_defined_fields.

### spec.userDefinedFields[].base

`string` · required

The header the offset is measured from:
- IPV4: the beginning of the IPv4 header
- IPV6: the beginning of the IPv6 header
- TCP: the beginning of the TCP header, skipping any IPv4 options or
  IPv6 extension headers; not present for non-first fragments
- UDP: the beginning of the UDP header, likewise

- rule: base must be one of: IPV4, IPV6, TCP, UDP
- rule: {"required":true}

### spec.userDefinedFields[].offset

`int32` · optional (explicit presence)

Offset of the first byte of the field (in network byte order) relative
to base. 0 is the first byte of the header and is a real position, so
presence matters: unset lets Google apply its default.

- rule: {"int32":{"gte":0}}

### spec.userDefinedFields[].size

`int32` · optional (explicit presence)

Size of the field in bytes, 1 to 4.

- rule: {"int32":{"lte":4,"gte":1}}

### spec.userDefinedFields[].mask

`string`

Bitwise-AND mask applied to the field before matching, as a
hexadecimal number starting with "0x" (e.g. "0x8F00"). The last byte
of the field (network byte order) corresponds to the least significant
byte of the mask.

- rule: mask must be a 0x-prefixed hexadecimal number of up to 4 bytes, e.g. 0x8F00

### spec.networkEdgeSecurityService

`GcpCloudArmorNetworkEdgeSecurityService`

Enroll this policy's region in advanced network DDoS protection by
creating the region's network edge security service with this policy
attached. One per region per project; requires ddos_protection_config.

### spec.networkEdgeSecurityService.name

`string`

Name of the service in GCP (RFC 1035). Defaults to the policy's name.
Immutable.

- rule: name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.networkEdgeSecurityService.description

`string`

Free-text description of the service.

- rule: {"string":{"maxLen":"2048"}}

## Validation Rules

- `rules_include_default`: a non-empty rule set must include the default rule at priority 2147483647 (match '*')
- `rule_priorities_unique`: each rule must have a unique priority
- `labels_global_only`: labels are a global-policy lever — the regional security policy collection carries no labels; clear region or remove labels
- `adaptive_protection_global_only`: adaptive_protection_config is a global-policy lever — Adaptive Protection watches global Application Load Balancers only; clear region or remove adaptive_protection_config
- `recaptcha_options_global_only`: recaptcha_options_config is a global-policy lever — the regional collection has no reCAPTCHA redirect; clear region or remove recaptcha_options_config
- `request_body_inspection_size_global_only`: advanced_options_config.request_body_inspection_size is a global-policy lever — the regional collection inspects the default 8KB only; clear region or remove request_body_inspection_size
- `internal_service_type_global_only`: type CLOUD_ARMOR_INTERNAL_SERVICE is a global-policy type (Traffic Director); clear region or choose CLOUD_ARMOR, CLOUD_ARMOR_EDGE, or CLOUD_ARMOR_NETWORK
- `header_action_global_only`: rules[].header_action is a global-policy lever — the regional collection cannot inject request headers; clear region or remove header_action from every rule
- `redirect_action_global_only`: the redirect rule action is a global CLOUD_ARMOR lever — the regional collection has no redirect; clear region or change the action
- `exceed_redirect_global_only`: rate_limit_options.exceed_action redirect is a global-policy lever — a regional rate limit exceeds to deny(STATUS); clear region or change exceed_action
- `expr_options_global_only`: rules[].match.expr_options (reCAPTCHA token site keys) is a global-policy lever; clear region or remove expr_options from every rule
- `network_type_regional_only`: type CLOUD_ARMOR_NETWORK is a regional-policy type (passthrough Network Load Balancers); set region or choose CLOUD_ARMOR, CLOUD_ARMOR_EDGE, or CLOUD_ARMOR_INTERNAL_SERVICE
- `ddos_protection_requires_network_type`: ddos_protection_config applies only to a regional policy of type CLOUD_ARMOR_NETWORK
- `user_defined_fields_require_network_type`: user_defined_fields apply only to a regional policy of type CLOUD_ARMOR_NETWORK
- `network_match_requires_network_type`: rules[].network_match applies only to a regional policy of type CLOUD_ARMOR_NETWORK; HTTP policies match through rules[].match
- `network_match_fields_defined`: every network_match.user_defined_fields entry must name a field declared in the policy's user_defined_fields
- `user_defined_field_names_unique`: each user-defined field must have a unique name within the policy
- `edge_service_requires_ddos_protection`: network_edge_security_service enrolls a region in advanced network DDoS protection and needs ddos_protection_config on this regional CLOUD_ARMOR_NETWORK policy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudArmorPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | Fully qualified resource ID of the security policy. Format: projects/{project}/global/securityPolicies/{name} for a global policy, projects/{project}/regions/{region}/securityPolicies/{name} for a regional one. |
| `status.outputs.policy_name` | `string` | Name of the security policy as it exists in GCP. |
| `status.outputs.policy_self_link` | `string` | Self-link URI of the security policy. This is the value used when attaching the policy to backend services, load balancers, or CDN configurations; a regional link carries regions/{region} where the global one says global, and a backend service accepts only a policy of its own scope. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/securityPolicies/{name} |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint of the policy. Used for optimistic concurrency control when updating the policy outside of IaC. |
| `status.outputs.region` | `string` | Region of a regional security policy; empty for a global one, so a consumer can tell the scope from the outputs alone. |
| `status.outputs.network_edge_security_service_self_link` | `string` | Self-link of the network edge security service created when the spec declares network_edge_security_service (the region's enrollment in advanced network DDoS protection); empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBackendBucket | `spec.edgeSecurityPolicy` | `status.outputs.policy_self_link` |
| GcpBackendService | `spec.securityPolicy` | `status.outputs.policy_self_link` |
| GcpBackendService | `spec.edgeSecurityPolicy` | `status.outputs.policy_self_link` |

## See Also

- [Overview](../README.md)
