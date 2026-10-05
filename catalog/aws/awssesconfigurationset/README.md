# AwsSesConfigurationSet

An **Amazon SES (SESv2) configuration set** is the named group of sending rules — TLS posture, IP pool, open/click tracking, suppression, deliverability dashboards, and event publishing — that email identities and individual send calls opt into.

Event destinations are folded satellites: each named destination is its own AWS sub-resource, materialized per-name by the modules so destinations can be added and removed independently.

## When to Use

- **Shared delivery posture** — Define TLS, suppression, and tracking once; every identity referencing the set inherits it.
- **Bounce/complaint observability** — Publish email events to CloudWatch, EventBridge, Firehose, SNS, or Pinpoint.
- **Reputation monitoring** — Enable CloudWatch reputation metrics for production senders.
- **Per-set kill switch** — Disable sending for every identity using the set without touching application config.

## When NOT to Use

- For a one-off test send with AWS defaults — identities work without a configuration set.
- For account-level suppression list management — that is an account setting, not a set.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | — | AWS region for the set. |
| `deliveryOptions.tlsPolicy` | string | No | OPTIONAL | `REQUIRE` or `OPTIONAL`. |
| `deliveryOptions.maxDeliverySeconds` | int | No | AWS default (14h) | 300–50400 seconds. |
| `deliveryOptions.sendingPoolName` | string | No | shared IP space | Dedicated IP pool name. |
| `reputationMetricsEnabled` | bool | No | false | Publish bounce/complaint rates to CloudWatch. |
| `sendingEnabled` | bool | No | true | Per-set sending kill switch. |
| `suppressedReasons` | list | No | account default | `BOUNCE`, `COMPLAINT`. |
| `trackingOptions` | object | No | — | Custom open/click tracking domain. |
| `vdmOptions` | object | No | — | Virtual Deliverability Manager overrides. |
| `eventDestinations[]` | list | No | — | Named event publishing destinations. |

## Outputs

| Output | Description |
|--------|-------------|
| `configuration_set_arn` | ARN for IAM policies scoping sends under the set. |
| `configuration_set_name` | Name (from `metadata.name`) referenced by identities and SendEmail. |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSesConfigurationSet
metadata:
  name: transactional-set
spec:
  region: us-west-2
  deliveryOptions:
    tlsPolicy: REQUIRE
  suppressedReasons:
    - BOUNCE
    - COMPLAINT
```

## Related Kinds

- [AwsSesEmailIdentity](../awssesemailidentity/README.md) — References a configuration set as its default sending rules.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
