# AWS Event-Driven Pipeline

Provisions an event-driven processing pipeline using EventBridge for event routing, SQS for reliable message processing with dead-letter queues, optional SNS fan-out notifications, and optional Step Functions workflow orchestration.

Applications publish events to the custom EventBridge bus. Rules match events by pattern and route them to SQS queues for processing. Failed messages are automatically moved to a dead-letter queue after configurable retry attempts.

## Architecture

```
  Event Producers
       │
       ▼
┌──────────────────┐
│AwsEventBridgeBus │
│ (custom bus)     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐      ┌──────────────────┐
│AwsEventBridgeRule│─────▶│   AwsSqsQueue    │
│ (event routing)  │      │  (processing)    │
│                  │      │                  │
│  DLQ target ─────│──┐   │  DLQ ──────────┐ │
└──────────────────┘  │   └────────────────│─┘
                      ▼                    ▼
               ┌──────────────────┐
               │   AwsSqsQueue    │
               │  (dead-letter)   │
               └──────────────────┘

┌──────────────────┐  ┌──────────────────────┐
│  AwsSnsTopic     │  │  AwsStepFunction     │
│  (fan-out)       │  │  (orchestration)     │
└──────────────────┘  └──────────────────────┘

┌──────────────────────────┐  ┌──────────────────┐
│  AwsIamRole              │  │ AwsCloudwatchLG  │
│  (SF execution role)     │  │ (SF exec logs)   │
└──────────────────────────┘  └──────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsEventBridgeBus, AwsSqsQueue (DLQ), AwsSnsTopic,
                     AwsCloudwatchLogGroup, AwsIamRole
Layer 1 (dep DLQ):   AwsSqsQueue (processing)
Layer 1 (dep IAM+CW): AwsStepFunction
Layer 2 (dep bus+SQS): AwsEventBridgeRule
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| EventBridge Bus | `AwsEventBridgeBus` | messaging | Always | Custom event bus for application events |
| SQS Queue (DLQ) | `AwsSqsQueue` | messaging | Always | Dead letter queue for failed messages |
| SQS Queue (processing) | `AwsSqsQueue` | messaging | Always | Event processing queue with DLQ redrive |
| EventBridge Rule | `AwsEventBridgeRule` | messaging | Always | Routes events from bus to SQS target |
| SNS Topic | `AwsSnsTopic` | messaging | `notificationsEnabled` | Fan-out event notifications |
| CloudWatch Log Group | `AwsCloudwatchLogGroup` | monitoring | `orchestrationEnabled` | Step Functions execution logs |
| IAM Role | `AwsIamRole` | identity | `orchestrationEnabled` | Step Functions execution role |
| Step Function | `AwsStepFunction` | compute | `orchestrationEnabled` | Workflow orchestration |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `bus_name` | EventBridge custom bus name | `app-events` | Yes |
| `bus_description` | Bus description | `Custom event bus for application events` | No |
| `rule_name` | EventBridge rule name | `route-to-processor` | Yes |
| `rule_description` | Rule description | `Routes application events to the SQS processing queue` | No |
| `processing_queue_name` | SQS processing queue name | `event-processor` | Yes |
| `processing_queue_visibility_timeout` | Visibility timeout (seconds) | `60` | No |
| `dlq_name` | Dead letter queue name | `event-processor-dlq` | Yes |
| `dlq_max_receive_count` | Failed attempts before DLQ | `3` | No |
| **Notifications** | | | |
| `notificationsEnabled` | Create SNS topic | `false` | No |
| `sns_topic_name` | SNS topic name | `event-notifications` | No |
| **Orchestration** | | | |
| `orchestrationEnabled` | Create Step Functions state machine | `false` | No |
| `step_function_name` | State machine name | `event-workflow` | No |
| `log_group_name` | CloudWatch log group for SF logs | `event-pipeline-logs` | No |

## Common Configurations

### Minimal (EventBridge + SQS)

```yaml
notificationsEnabled: false
orchestrationEnabled: false
```

### With Notifications

```yaml
notificationsEnabled: true
orchestrationEnabled: false
```

### Full Pipeline with Orchestration

```yaml
notificationsEnabled: true
orchestrationEnabled: true
```

## Publishing Events

Send events to the custom bus using the AWS SDK:

```python
import boto3, json

client = boto3.client('events')
client.put_events(
    Entries=[{
        'Source': 'app-events',
        'DetailType': 'OrderCreated',
        'Detail': json.dumps({'orderId': '12345', 'amount': 99.99}),
        'EventBusName': 'app-events'
    }]
)
```

## Important Notes

- The EventBridge rule matches events with `source` equal to the bus name. Update the `eventPattern` in the rule resource to match your specific event structure.
- SQS processing queue messages are retried `dlq_max_receive_count` times before moving to the DLQ. Set `processing_queue_visibility_timeout` higher than your consumer's processing time.
- The DLQ retains messages for **14 days** (maximum SQS retention). Monitor and reprocess DLQ messages before they expire.
- The Step Function receives a **placeholder definition** (single Pass state). Replace it with your actual workflow after deployment.
- Step Functions execution role includes SQS and SNS permissions scoped to this chart's resources. Add additional permissions for Lambda, DynamoDB, or other services your workflow invokes.
