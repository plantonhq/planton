# AwsEventBridgeBus examples

## Minimal Custom Bus

A custom event bus with a description and all other AWS defaults.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: order-events
  labels:
    app: shop
spec:
  description: Custom bus for order processing events
```

## Production Bus with KMS Encryption and DLQ

A custom bus encrypted with a customer-managed KMS key, dead letter queue for undeliverable events, and error-level logging.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: payment-events
spec:
  description: Payment processing event bus with encryption and DLQ
  kmsKeyIdentifier:
    valueFrom:
      kind: AwsKmsKey
      name: payment-key
      fieldPath: status.outputs.key_arn
  deadLetterConfig:
    arn:
      valueFrom:
        kind: AwsSqsQueue
        name: payment-bus-dlq
        fieldPath: status.outputs.queue_arn
  logConfig:
    level: ERROR
    includeDetail: NONE
```

## Bus with Full Trace Logging

A custom bus with verbose logging for debugging event routing during development.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: dev-events
spec:
  description: Development event bus with trace logging
  logConfig:
    level: TRACE
    includeDetail: FULL
```

## Partner Event Bus

A bus for receiving events from a SaaS partner integration. The bus name must match the event source name.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: aws.partner/example.com/tenant123/orders
spec:
  eventSourceName: aws.partner/example.com/tenant123/orders
```

## Bus with DLQ (Literal ARN)

A custom bus with a dead letter queue specified as a direct ARN value.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: notification-events
spec:
  description: Notification bus with dead letter queue
  deadLetterConfig:
    arn:
      value: arn:aws:sqs:us-east-1:123456789012:notification-bus-dlq
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./manifest.yaml | cat
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack myorg/infra/dev --module-dir ./apis/org/openmcf/provider/aws/awseventbridgebus/v1/iac/pulumi | cat
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve | cat
```

> Note: Provider credentials are supplied via stack input, not in the spec.
