# AWS SES Configuration Set

Deploys an Amazon SES (SESv2) configuration set — the named group of sending rules that email identities and individual send calls opt into. A configuration set is the policy layer of the SES graph: TLS posture, dedicated IP pool, open/click tracking, suppression, deliverability dashboards, and event publishing are defined once here, and every identity that references the set ([AwsSesEmailIdentity](/infra-catalog/aws-ses-email-identity)) inherits them. Event publishing is where SES becomes observable — bounce/complaint feedback loops, the difference between a healthy sender reputation and a suspended account, are built from exactly the event destinations this set carries.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SES Configuration Set** -- the named rule group (the set name derives from the resource name), with the sending kill switch, reputation metrics, and suppression overrides
- **Delivery Options** -- TLS enforcement, the retry budget, and the dedicated IP pool binding, when configured
- **Tracking Options** -- the custom open/click tracking redirect domain, when configured
- **VDM Overrides** -- per-set Virtual Deliverability Manager positions overriding the account configuration, when configured
- **Event Destinations** -- one AWS sub-resource per named destination, each streaming its selected email events into CloudWatch metrics, an SNS topic, an EventBridge bus, a Kinesis Firehose stream, or a Pinpoint application
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Destination resources** -- event destinations reference other Planton resources: an [AwsSnsTopic](/infra-catalog/aws-sns-topic) for the classic feedback loop, an AwsKinesisFirehose plus an [AwsIamRole](/infra-catalog/aws-iam-role) for analytics, or an AwsEventBridgeBus for rule-based routing. Create them first, or paste literal ARNs.

### AWS Account

- **Same region as your identities** -- SES resources are regional; an identity can only reference a configuration set in its own region.
- **For Firehose destinations** -- the IAM role must trust `ses.amazonaws.com` and allow `firehose:PutRecordBatch` on the delivery stream.
- **For EventBridge destinations** -- AWS only supports the account's default bus for SES event publishing today.

## Deploy

### Console

Open the deployment store, find **AWS SES Configuration Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Minimal Transactional** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesConfigurationSet
metadata:
  name: transactional-prod
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  reputationMetricsEnabled: true
  suppressedReasons:
    - BOUNCE
    - COMPLAINT
  eventDestinations:
    - name: bounce-feedback
      matchingEventTypes:
        - BOUNCE
        - COMPLAINT
      snsTopic:
        valueFrom:
          kind: AwsSnsTopic
          name: ses-feedback
          fieldPath: status.outputs.topic_arn
```

```shell
planton apply -f ses-configuration-set.yaml
```

This creates a production transactional set with reputation metrics on, both suppression reasons honored, and an SNS bounce/complaint feedback loop. An Infra Job tracks the provisioning in real time.

### InfraChart

When the set deploys alongside its event-destination targets in one chart, wire the destination references via ValueFromRef:

```yaml
spec:
  reputationMetricsEnabled: true
  eventDestinations:
    - name: bounce-feedback
      matchingEventTypes:
        - BOUNCE
        - COMPLAINT
      snsTopic:
        valueFrom:
          kind: AwsSnsTopic
          name: ses-feedback
          fieldPath: status.outputs.topic_arn
    - name: event-analytics
      matchingEventTypes:
        - SEND
        - DELIVERY
        - OPEN
        - CLICK
      firehose:
        deliveryStream:
          valueFrom:
            kind: AwsKinesisFirehose
            name: ses-events
            fieldPath: status.outputs.delivery_stream_arn
        iamRole:
          valueFrom:
            kind: AwsIamRole
            name: ses-firehose-delivery
            fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the SNS topic, Firehose stream, and IAM role first, then provisions the configuration set — and the AwsSesEmailIdentity resources that reference its `configuration_set_name` output deploy after it.

## Key Configuration

These are the most important decisions when configuring a configuration set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The kill switch** -- `sendingEnabled` stops every send that uses the set the moment it flips off, without touching identities or application config. Leave it unset for normal operation (the AWS default is enabled); it exists so you can contain a compromised credential in seconds.

**Reputation metrics** -- off by default AWS-side, and the single most important switch for a production sender: SES suspends accounts over bounce (~5%) and complaint (~0.1%) rates, and this is the built-in way to watch them. Alarm on the published CloudWatch metrics before AWS acts on them.

**Event destinations** -- each named destination streams its selected events (sends, bounces, complaints, opens, clicks, ...) into exactly one service. The production baseline is a bounce/complaint feedback loop: CloudWatch metrics for alarms, or an SNS topic your application subscribes to so bounced addresses leave your lists automatically. Destinations are independent sub-resources — add and remove them without touching the set.

**Suppression override** -- an empty `suppressedReasons` list inherits the account-level default; listing BOUNCE and/or COMPLAINT takes an explicit per-set position. Both are strongly recommended for any non-transactional sender.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** (optional) | `eventDestinations[].snsTopic` | `status.outputs.topic_arn` |
| **AwsEventBridgeBus** (optional) | `eventDestinations[].eventBus` | `status.outputs.bus_arn` |
| **AwsKinesisFirehose** (optional) | `eventDestinations[].firehose.deliveryStream` | `status.outputs.delivery_stream_arn` |
| **AwsIamRole** (optional) | `eventDestinations[].firehose.iamRole` | `status.outputs.role_arn` |

Pinpoint destinations take a literal ARN (Pinpoint is not modeled in this catalog). A set with no event destinations is a leaf.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `configuration_set_name` | The set's name | AwsSesEmailIdentity's `configurationSet` reference; explicit SendEmail calls |
| `configuration_set_arn` | Amazon Resource Name of the set | IAM policies scoping who may send under this set |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Transactional baseline** -- reputation metrics on, both suppression reasons honored, AWS transport defaults. Start from the **Minimal Transactional** preset.

**Observable sender** -- the baseline plus a CloudWatch event destination dimensioned by message tags, so bounce rate slices by campaign or tenant. Start from the **CloudWatch Event Destination** preset.

## Works With

- [**AWS SES Email Identity**](/infra-catalog/aws-ses-email-identity) -- the verified sending domain or address that inherits this set as its default (references `configuration_set_name`)
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- the classic bounce/complaint feedback-loop destination
- [**AWS Kinesis Firehose**](/infra-catalog/aws-kinesis-firehose) -- streams email events to S3, Redshift, or OpenSearch for durable analytics
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the role SES assumes for Firehose event delivery
