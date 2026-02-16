# AwsSqsQueue Pulumi Module Architecture

## Module Structure

```
module/
├── main.go      # Entry point: provider setup, orchestration, delegates to queue()
├── locals.go    # Locals struct: derived queue name, tags, spec references
├── outputs.go   # Output key constants matching AwsSqsQueueStackOutputs
└── queue.go     # SQS queue resource creation and output exports
```

## Data Flow

1. **main.go** receives `AwsSqsQueueStackInput` containing the target resource and provider config
2. **locals.go** derives the queue name (appending `.fifo` for FIFO queues) and constructs AWS tags from metadata
3. **queue.go** creates the `sqs.Queue` resource with:
   - Delivery settings (visibility timeout, retention, max size, delay, long polling)
   - FIFO settings (deduplication, throughput limit) — only for FIFO queues
   - Dead letter queue via serialized redrive policy JSON
   - Encryption (KMS or SSE-SQS, mutually exclusive)
   - Access policy serialized from `google.protobuf.Struct` to JSON
4. **outputs.go** defines constants for stack output keys: `queue_url`, `queue_arn`, `queue_name`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **google.protobuf.Struct**: Serialized to JSON via `spec.Policy.AsMap()` + `json.Marshal()`.
- **Redrive policy**: Constructed as a Go map and serialized to JSON inline, matching the AWS SQS API's JSON format.
- **Zero means default**: Numeric fields left at 0 are not set in the Pulumi args, letting AWS apply its defaults.
