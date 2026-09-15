# AwsSesConfigurationSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSesConfigurationSetSpec defines an Amazon SES (SESv2) configuration
set: the named group of sending rules -- TLS posture, dedicated IP pool,
open/click tracking, suppression, deliverability dashboards, and event
publishing -- that email identities and individual send calls opt into.

A configuration set is the policy layer of the SES graph. Identities
reference one as their default (AwsSesEmailIdentity's configuration_set
field), and any SendEmail call can name one explicitly. Because many
identities share one set, an organization defines its delivery posture
once -- "require TLS, publish bounces to SNS, track engagement" -- and
every sender inherits it.

Event publishing is where SES becomes observable: each named event
destination streams a chosen set of email events (sends, bounces,
complaints, opens, clicks, ...) into CloudWatch metrics, EventBridge,
Kinesis Firehose, SNS, or Pinpoint. Bounce/complaint feedback loops --
the difference between a healthy sender reputation and a suspended
account -- are built from exactly these destinations.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesConfigurationSet
metadata:
  name: test-ses-config-set
  org: test-org
  env: dev
  id: test-ses-config-set-dev
spec:
  region: us-west-2
  reputationMetricsEnabled: true
  deliveryOptions:
    tlsPolicy: REQUIRE
  suppressedReasons:
    - BOUNCE
    - COMPLAINT
  eventDestinations:
    - name: bounce-metrics
      matchingEventTypes:
        - SEND
        - BOUNCE
        - COMPLAINT
      cloudWatch:
        dimensions:
          - name: campaign
            valueSource: MESSAGE_TAG
            defaultValue: none
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deliveryOptions` | `AwsSesConfigurationSetDeliveryOptions` |  |  |  |
| `spec.deliveryOptions.tlsPolicy` | `string` |  | `OPTIONAL` |  |
| `spec.deliveryOptions.maxDeliverySeconds` | `int32` |  |  |  |
| `spec.deliveryOptions.sendingPoolName` | `string` |  |  |  |
| `spec.reputationMetricsEnabled` | `bool` |  |  |  |
| `spec.sendingEnabled` | `bool` |  | `true` |  |
| `spec.suppressedReasons` | `[]string` |  |  |  |
| `spec.trackingOptions` | `AwsSesConfigurationSetTrackingOptions` |  |  |  |
| `spec.trackingOptions.customRedirectDomain` | `string` | yes |  |  |
| `spec.trackingOptions.httpsPolicy` | `string` |  | `OPTIONAL` |  |
| `spec.vdmOptions` | `AwsSesConfigurationSetVdmOptions` |  |  |  |
| `spec.vdmOptions.engagementMetricsEnabled` | `bool` |  | `false` |  |
| `spec.vdmOptions.optimizedSharedDeliveryEnabled` | `bool` |  | `false` |  |
| `spec.eventDestinations` | `[]AwsSesConfigurationSetEventDestination` |  |  |  |
| `spec.eventDestinations[].name` | `string` | yes |  |  |
| `spec.eventDestinations[].enabled` | `bool` |  | `true` |  |
| `spec.eventDestinations[].matchingEventTypes` | `[]string` | yes |  |  |
| `spec.eventDestinations[].cloudWatch` | `AwsSesConfigurationSetEventDestinationCloudWatch` |  |  |  |
| `spec.eventDestinations[].cloudWatch.dimensions` | `[]AwsSesConfigurationSetEventDestinationCloudWatchDimension` | yes |  |  |
| `spec.eventDestinations[].cloudWatch.dimensions[].name` | `string` | yes |  |  |
| `spec.eventDestinations[].cloudWatch.dimensions[].valueSource` | `string` | yes |  |  |
| `spec.eventDestinations[].cloudWatch.dimensions[].defaultValue` | `string` | yes |  |  |
| `spec.eventDestinations[].eventBus` | `string \| valueFrom` |  |  | AwsEventBridgeBus (`status.outputs.bus_arn`) |
| `spec.eventDestinations[].firehose` | `AwsSesConfigurationSetEventDestinationFirehose` |  |  |  |
| `spec.eventDestinations[].firehose.deliveryStream` | `string \| valueFrom` | yes |  | AwsKinesisFirehose (`status.outputs.delivery_stream_arn`) |
| `spec.eventDestinations[].firehose.iamRole` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.eventDestinations[].snsTopic` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.eventDestinations[].pinpointApplicationArn` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the configuration set is created. SES resources
are regional: identities can only reference configuration sets in
their own region.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.deliveryOptions

`AwsSesConfigurationSetDeliveryOptions`

Transport and IP-pool controls for messages sent under this set.

### spec.deliveryOptions.tlsPolicy

`string` · optional (explicit presence)

TLS enforcement for delivery to receiving mail servers:
  REQUIRE  -- deliver only over a TLS-protected connection; fail the
              send if the receiver cannot negotiate TLS. The right
              choice for anything carrying sensitive content.
  OPTIONAL -- attempt TLS but fall back to plaintext (the SMTP
              default, and AWS's default when unset).

- default: `OPTIONAL`
- rule: {"string":{"in":["REQUIRE","OPTIONAL"]}}

### spec.deliveryOptions.maxDeliverySeconds

`int32` · optional (explicit presence)

How long, in seconds (300-50400, i.e. 5 minutes to 14 hours), SES
keeps retrying delivery of a message before giving up and returning a
bounce. Shorter values suit time-sensitive mail (one-time passcodes
are useless after 10 minutes); the AWS default is 14 hours.

- rule: {"int32":{"lte":50400,"gte":300}}

### spec.deliveryOptions.sendingPoolName

`string`

The dedicated IP pool messages under this set are sent from. Pass the
name of a dedicated IP pool created outside this catalog (dedicated
IPs are a paid capacity surface with their own lifecycle). Leave
unset to send from the shared SES IP space.

- rule: {"string":{"maxLen":"64"}}

### spec.reputationMetricsEnabled

`bool`

Whether SES publishes reputation metrics (bounce and complaint rates)
for this set to CloudWatch. Off by default AWS-side; turn it on for
any production sender -- reputation is what SES suspends accounts
over, and this is the built-in way to watch it.

### spec.sendingEnabled

`bool` · optional (explicit presence)

Whether email sending is enabled for this configuration set. This is
the per-set kill switch: flip it off to immediately stop all sends
that use the set (e.g. when a compromised key is spraying spam)
without touching identities or application config. Defaults to true.

- default: `true`

### spec.suppressedReasons

`[]string`

The account-level suppression-list reasons this set honors. When a
recipient address is on the account suppression list for one of these
reasons, SES silently skips sending to it under this set:
  BOUNCE    -- addresses that previously hard-bounced.
  COMPLAINT -- addresses whose owners marked mail as spam.
Both are strongly recommended for any non-transactional sender; leave
the list empty to inherit the account-level default instead of
overriding it.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["BOUNCE","COMPLAINT"]}}}}

### spec.trackingOptions

`AwsSesConfigurationSetTrackingOptions`

Custom open/click tracking configuration. SES rewrites links and
embeds a tracking pixel through its own domain by default; setting a
custom redirect domain keeps tracking URLs on YOUR domain (a CNAME to
SES's tracking endpoint), which looks trustworthy to recipients and
corporate link scanners.

### spec.trackingOptions.customRedirectDomain

`string` · required

The custom subdomain SES rewrites tracking links through. The domain
must CNAME to the regional SES tracking endpoint (e.g.
"click.example.com" -> "r.us-west-2.awstrack.me") -- compose the CNAME
with an AwsRoute53DnsRecord. Required when tracking options are set.

- rule: {"required":true,"string":{"maxLen":"253"}}

### spec.trackingOptions.httpsPolicy

`string` · optional (explicit presence)

The scheme SES uses for the rewritten tracking links:
  REQUIRE           -- https for both open and click tracking.
  REQUIRE_OPEN_ONLY -- https for the open-tracking pixel only.
  OPTIONAL          -- match the scheme of the original link (AWS's
                       default when unset).

- default: `OPTIONAL`
- rule: {"string":{"in":["REQUIRE","REQUIRE_OPEN_ONLY","OPTIONAL"]}}

### spec.vdmOptions

`AwsSesConfigurationSetVdmOptions`

Virtual Deliverability Manager (VDM) toggles for this set, overriding
the account-level VDM configuration.

### spec.vdmOptions.engagementMetricsEnabled

`bool` · optional (explicit presence)

Whether VDM engagement metrics (opens/clicks broken down by ISP,
subject line, and sending identity in the VDM dashboard) are
collected for mail sent under this set. Unset is DISABLED.

- default: `false`

### spec.vdmOptions.optimizedSharedDeliveryEnabled

`bool` · optional (explicit presence)

Whether VDM optimized shared delivery is used: SES picks the shared
IP with the best standing for each receiving ISP instead of rotating
blindly. Only meaningful when sending from the shared IP space (not a
dedicated pool). Unset is DISABLED.

- default: `false`

### spec.eventDestinations

`[]AwsSesConfigurationSetEventDestination`

Named event destinations that publish email events (sends, bounces,
complaints, opens, clicks, ...) from this set into other AWS
services. Each destination is its own AWS sub-resource keyed by name,
materialized per-name by the modules, so entries can be added and
removed independently without touching the set. Names must be unique.

- rule: exactly one destination must be set: cloud_watch, event_bus, firehose, sns_topic, or pinpoint_application_arn

### spec.eventDestinations[].name

`string` · required

The destination's name, unique within the configuration set (used as
the AWS event-destination identifier; changing it replaces the
destination).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[a-zA-Z0-9_-]+$"}}

### spec.eventDestinations[].enabled

`bool` · optional (explicit presence)

Whether the destination actively publishes events. Defaults to true
-- a created-but-disabled destination is almost never what a manifest
means (AWS's own default is false, a common source of silently
missing events; the modules always send this value explicitly).

- default: `true`

### spec.eventDestinations[].matchingEventTypes

`[]string` · required

The email events this destination receives:
  SEND              -- the send API call was accepted.
  REJECT            -- SES refused the message (e.g. a virus was
                       detected).
  BOUNCE            -- the receiver rejected the message (hard
                       bounces; the reputation killer).
  COMPLAINT         -- the recipient marked the message as spam.
  DELIVERY          -- the receiver accepted the message.
  OPEN              -- the recipient opened the message (needs
                       tracking).
  CLICK             -- the recipient clicked a tracked link.
  RENDERING_FAILURE -- a template variable failed to render (template
                       sends only).
  DELIVERY_DELAY    -- delivery is being retried (soft bounce,
                       mailbox full, ...).
  SUBSCRIPTION      -- the recipient changed subscription preferences
                       (list-management senders).

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"in":["SEND","REJECT","BOUNCE","COMPLAINT","DELIVERY","OPEN","CLICK","RENDERING_FAILURE","DELIVERY_DELAY","SUBSCRIPTION"]}}}}

### spec.eventDestinations[].cloudWatch

`AwsSesConfigurationSetEventDestinationCloudWatch`

Publish events as CloudWatch metrics, dimensioned by the configured
sources -- the zero-infrastructure way to alarm on bounce rate.

### spec.eventDestinations[].cloudWatch.dimensions

`[]AwsSesConfigurationSetEventDestinationCloudWatchDimension` · required

The dimensions attached to each published metric. At least one is
required by AWS.

- rule: {"required":true,"repeated":{"minItems":"1"}}

### spec.eventDestinations[].cloudWatch.dimensions[].name

`string` · required

The dimension name as it appears in CloudWatch (1-256 characters).
Example: "campaign", "sender".

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.eventDestinations[].cloudWatch.dimensions[].valueSource

`string` · required

Where the dimension's value comes from on each message:
  MESSAGE_TAG  -- an X-SES-MESSAGE-TAGS tag set by the sender at send
                  time (the common choice: tag sends with campaign or
                  tenant and slice metrics by it).
  EMAIL_HEADER -- a header on the outgoing message.
  LINK_TAG     -- a query-string tag on the clicked link (CLICK
                  events only).

- rule: {"required":true,"string":{"in":["MESSAGE_TAG","EMAIL_HEADER","LINK_TAG"]}}

### spec.eventDestinations[].cloudWatch.dimensions[].defaultValue

`string` · required

The value used when the message carries no value for this dimension
(1-256 characters). Example: "none".

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.eventDestinations[].eventBus

`string | valueFrom`

Publish events onto an EventBridge bus for rule-based routing.
Reference an AwsEventBridgeBus's bus_arn output or pass a literal
ARN. NOTE: AWS only supports the account's DEFAULT bus for SES event
publishing today; rules on the default bus can then forward anywhere.

- references: AwsEventBridgeBus (`status.outputs.bus_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEventBridgeBus, name: <that resource's name>, fieldPath: status.outputs.bus_arn}} -- a bare string does not parse

### spec.eventDestinations[].firehose

`AwsSesConfigurationSetEventDestinationFirehose`

Stream events into a Kinesis Data Firehose delivery stream -- the
firehose-to-S3/Redshift/OpenSearch path for durable event analytics.

### spec.eventDestinations[].firehose.deliveryStream

`string | valueFrom` · required

The Firehose delivery stream that receives the events. Reference an
AwsKinesisFirehose's delivery_stream_arn output or pass a literal ARN.

- references: AwsKinesisFirehose (`status.outputs.delivery_stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisFirehose, name: <that resource's name>, fieldPath: status.outputs.delivery_stream_arn}} -- a bare string does not parse

### spec.eventDestinations[].firehose.iamRole

`string | valueFrom` · required

The IAM role SES assumes to put records into the delivery stream
(trust principal ses.amazonaws.com, with firehose:PutRecordBatch on
the stream). Reference an AwsIamRole's role_arn output or pass a
literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.eventDestinations[].snsTopic

`string | valueFrom`

Publish each event as an SNS message -- the classic bounce/complaint
feedback-loop wiring (subscribe a queue or webhook to the topic).
Reference an AwsSnsTopic's topic_arn output or pass a literal ARN.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.eventDestinations[].pinpointApplicationArn

`string`

Send engagement events to an Amazon Pinpoint project (literal ARN --
Pinpoint is not modeled in this catalog).

## Validation Rules

- `event_destination_names_unique`: each event destination must have a unique name within the configuration set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSesConfigurationSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.configuration_set_arn` | `string` | The Amazon Resource Name (ARN) of the configuration set -- the target for IAM policies that scope who may send under this set. |
| `status.outputs.configuration_set_name` | `string` | The configuration set's name (derived from metadata.name) -- what email identities reference through their configuration_set field and what SendEmail calls name explicitly. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.eventDestinations[].eventBus` | AwsEventBridgeBus | `status.outputs.bus_arn` |
| `spec.eventDestinations[].firehose.deliveryStream` | AwsKinesisFirehose | `status.outputs.delivery_stream_arn` |
| `spec.eventDestinations[].firehose.iamRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.eventDestinations[].snsTopic` | AwsSnsTopic | `status.outputs.topic_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsCognitoUserPool | `spec.emailConfiguration.configurationSet` | `status.outputs.configuration_set_name` |
| AwsSesEmailIdentity | `spec.configurationSet` | `status.outputs.configuration_set_name` |

## See Also

- [Overview](../README.md)
