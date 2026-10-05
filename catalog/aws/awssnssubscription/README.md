# AwsSnsSubscription

SNS subscription resource for Planton. Provisions the delivery edge that wires an SNS topic to a target endpoint — an SQS queue, a Lambda function, an HTTP/S endpoint, an email address, an SMS number, a Kinesis Data Firehose stream, or a mobile platform application endpoint.

## When to use

- You want a topic's messages delivered somewhere — the subscription IS the fan-out edge in the pub/sub graph.
- You need per-consumer control: message filtering, raw delivery, a delivery DLQ, or archived-message replay, each owned by the consumer rather than the topic.
- You are subscribing an endpoint you own to a topic owned by another team or account (reference the topic by literal ARN).

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| SNS topic | The topic this subscription delivers from | `AwsSnsTopic` |
| Endpoint | The delivery target, varies by protocol | `AwsSqsQueue`, `AwsLambda`, `AwsKinesisFirehose`, plain URL/email/phone |

For `sqs` delivery the queue's own resource `policy` must grant `sqs:SendMessage` to `sns.amazonaws.com` (scoped by `aws:SourceArn` to the topic). Subscription creation succeeds without it, but every delivery is silently dropped.

## API envelope

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSnsSubscription
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | **yes** | — | AWS region (must match the topic). |
| `topicArn` | StringValueOrRef | **yes** | — | Topic to subscribe to. Default ref: `AwsSnsTopic` / `topic_arn`. **ForceNew**. |
| `protocol` | string | **yes** | — | One of `sqs`, `lambda`, `http`, `https`, `email`, `email-json`, `sms`, `firehose`, `application`. **ForceNew**. |
| `endpoint` | StringValueOrRef | **yes** | — | Delivery target; format depends on protocol. **ForceNew**. |
| `filterPolicy` | object | no | — | JSON filter selecting which messages are delivered. |
| `filterPolicyScope` | string | no | `MessageAttributes` | `MessageAttributes` or `MessageBody`; requires `filterPolicy`. |
| `rawMessageDelivery` | bool | no | `false` | Deliver the message body as-is without the JSON envelope (SQS/HTTP/S/Firehose). |
| `deadLetterConfig` | object | no | — | SQS queue receiving messages whose delivery ultimately failed. |
| `deliveryPolicy` | string | no | topic policy | Per-subscription HTTP/S retry policy JSON override. |
| `replayPolicy` | object | no | — | Replay archived messages from a FIFO topic's archive to backfill this consumer. |
| `subscriptionRoleArn` | StringValueOrRef | firehose | — | Role SNS assumes to write to the Firehose stream. Default ref: `AwsIamRole` / `role_arn`. |
| `endpointAutoConfirms` | bool | no | `false` | HTTP/S endpoint self-confirms the subscription handshake. |
| `confirmationTimeoutMinutes` | int32 | no | `1` | Minutes to wait for HTTP/S confirmation before failing. |

## Outputs

| Output | Description |
|---|---|
| `subscription_arn` | The subscription ARN — AWS API identity. |
| `owner_id` | AWS account ID that owns the subscription. |
| `pending_confirmation` | True while awaiting endpoint confirmation. |
| `confirmation_was_authenticated` | True when the confirmation was signed/authenticated. |

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSnsSubscription
metadata:
  name: fulfillment-events
spec:
  region: us-west-2
  topicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: order-events
      fieldPath: status.outputs.topic_arn
  protocol: sqs
  endpoint:
    valueFrom:
      kind: AwsSqsQueue
      name: fulfillment-queue
      fieldPath: status.outputs.queue_arn
  rawMessageDelivery: true
  filterPolicy:
    event_type:
      - order_placed
      - order_shipped
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
