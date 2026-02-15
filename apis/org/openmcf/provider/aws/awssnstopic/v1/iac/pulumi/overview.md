# AwsSnsTopic Pulumi Module Architecture

## Module Structure

```
module/
├── main.go           # Entry point: provider setup, orchestration
├── locals.go         # Locals struct: derived topic name, tags, spec references
├── outputs.go        # Output key constants matching AwsSnsTopicStackOutputs
├── topic.go          # SNS topic resource creation and topic-level output exports
└── subscription.go   # SNS topic subscriptions (iterates repeated field)
```

## Data Flow

1. **main.go** receives `AwsSnsTopicStackInput` containing the target resource and provider config
2. **locals.go** derives the topic name (appending `.fifo` for FIFO topics) and constructs AWS tags from metadata
3. **topic.go** creates the `sns.Topic` resource with:
   - FIFO settings (deduplication, throughput scope) — only for FIFO topics
   - Display name
   - KMS encryption
   - Access policy serialized from `google.protobuf.Struct` to JSON
   - Delivery policy (passed through as string)
   - Tracing config and signature version
4. **subscription.go** iterates `spec.Subscriptions` and creates `sns.TopicSubscription` resources with:
   - Protocol and endpoint (from StringValueOrRef)
   - Filter policy serialized from `google.protobuf.Struct` to JSON
   - Raw message delivery flag
   - Redrive policy (subscription DLQ) serialized to JSON
   - Firehose subscription role ARN
5. **outputs.go** defines constants for stack output keys: `topic_arn`, `topic_name`, `subscription_arns`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **google.protobuf.Struct**: Both `policy` and `filter_policy` are serialized via `.AsMap()` + `json.Marshal()`.
- **Redrive policy**: Constructed as a Go map with `deadLetterTargetArn` and serialized to JSON, matching the SNS API format.
- **Subscription ARN map**: Built as `pulumi.StringMap` keyed by subscription name, following the AwsS3ObjectSet pattern for map-keyed outputs.
- **Topic ARN wiring**: Subscriptions reference the created topic's ARN via `createdTopic.Arn`, ensuring Pulumi tracks the dependency.
