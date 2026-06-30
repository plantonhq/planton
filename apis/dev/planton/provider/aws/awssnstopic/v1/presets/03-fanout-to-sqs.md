# Fan-Out to SQS with Filtering

This preset creates a Standard SNS topic with two SQS queue subscriptions, each with a message attribute filter policy. This demonstrates the core SNS fan-out pattern — a single topic distributes messages to multiple specialized queues based on event type.

## When to Use

- Event-driven architectures where different services process different event types
- Microservice decoupling — the publisher sends to one topic; subscribers self-select relevant events
- CQRS patterns where commands and queries route to different processing queues
- Multi-stage processing pipelines with branching event flows

## Key Configuration Choices

- **Two filtered SQS subscriptions** — each queue receives only messages matching its filter policy, reducing unnecessary processing
- **Filter on MessageAttributes** — the most common and performant filtering approach; filter evaluation happens server-side before message delivery
- **Raw message delivery** (`rawMessageDelivery: true`) — delivers the raw message body directly to SQS without the SNS JSON envelope, simplifying consumer deserialization
- **`valueFrom` references** — subscription endpoints reference AwsSqsQueue outputs, creating dependency edges in the infra chart DAG. The SQS queues are provisioned first, then the SNS topic with subscriptions.

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `fulfillment-events` | Name of the SQS queue for fulfillment events | Your actual AwsSqsQueue resource name |
| `billing-events` | Name of the SQS queue for billing events | Your actual AwsSqsQueue resource name |
| Filter policy values | Event types in the filter policies | Your actual event type attribute values |

## Important: SQS Queue Policy

For SNS to deliver messages to an SQS queue, the queue must have an IAM access policy that grants the SNS topic permission to `sqs:SendMessage`. Add a `policy` to each target AwsSqsQueue:

```yaml
spec:
  policy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: sns.amazonaws.com
        Action: sqs:SendMessage
        Resource: "*"
        Condition:
          ArnEquals:
            aws:SourceArn: <topic-arn>
```

## Common Additions

- Add `redriveConfig` to subscriptions for dead letter queue routing on delivery failures
- Add `kmsKeyId` to the topic for encryption at rest
- Add more subscriptions (Lambda, email, HTTPS) for additional fanout targets
- Use `filterPolicyScope: MessageBody` for body-based filtering (more flexible but heavier)

## Related Presets

- **01-standard-topic** — minimal topic without subscriptions
- **02-fifo-with-deduplication** — use when you need exactly-once ordered delivery
