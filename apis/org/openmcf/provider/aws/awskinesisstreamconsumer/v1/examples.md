# AwsKinesisStreamConsumer Examples

## 1. Minimal Consumer with Direct ARN

The simplest possible enhanced fan-out consumer. Registers with an existing stream using a literal ARN.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: my-consumer
  org: acme
  env: dev
  id: my-consumer-dev
spec:
  streamArn:
    value: arn:aws:kinesis:us-east-1:123456789012:stream/my-stream
```

## 2. Consumer with Stream Reference (valueFrom)

References an OpenMCF-managed Kinesis stream. The platform resolves the ARN at deployment time and builds a dependency edge in the infra chart DAG.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: analytics-consumer
  org: acme
  env: production
  id: analytics-consumer-prod
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: clickstream-events
      fieldPath: status.outputs.stream_arn
```

## 3. Multiple Consumers on the Same Stream

Enhanced fan-out enables multiple independent consumers, each with dedicated 2 MB/s throughput. Deploy each as a separate resource.

**Consumer 1 — Real-time dashboard:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: dashboard-consumer
  org: acme
  env: production
  id: dashboard-consumer-prod
  labels:
    team: frontend
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
```

**Consumer 2 — Audit trail:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: audit-consumer
  org: acme
  env: production
  id: audit-consumer-prod
  labels:
    team: compliance
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
```

**Consumer 3 — Analytics pipeline:**

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: analytics-consumer
  org: acme
  env: production
  id: analytics-consumer-prod
  labels:
    team: data-engineering
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
```

## 4. Consumer for Lambda Enhanced Fan-Out

After deploying this consumer, configure a Lambda event source mapping with the `consumer_arn` output to enable enhanced fan-out delivery.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: lambda-processor
  org: acme
  env: production
  id: lambda-processor-prod
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: payment-events
      fieldPath: status.outputs.stream_arn
```

After deployment, the `consumer_arn` output (e.g., `arn:aws:kinesis:us-east-1:123456789012:stream/payment-events/consumer/lambda-processor:1234567890`) is used in the Lambda event source mapping configuration to enable push-based delivery.

## 5. Consumer with Full Metadata

Shows all metadata fields for production use including organization, environment, ID, and labels.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: telemetry-consumer
  org: acme
  env: production
  id: telemetry-consumer-prod
  labels:
    team: platform
    cost-center: infrastructure
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: streaming
    pulumi.openmcf.org/stack.name: prod.AwsKinesisStreamConsumer.telemetry-consumer
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: telemetry-stream
      fieldPath: status.outputs.stream_arn
```
