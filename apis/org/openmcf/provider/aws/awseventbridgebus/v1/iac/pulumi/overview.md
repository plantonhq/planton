# AwsEventBridgeBus Pulumi Module Architecture

## Module Structure

```
module/
├── main.go        # Entry point: provider setup, orchestration, delegates to eventBus()
├── locals.go      # Locals struct: tags, spec references
├── outputs.go     # Output key constants matching AwsEventBridgeBusStackOutputs
└── event_bus.go   # EventBridge bus resource creation and output exports
```

## Data Flow

1. **main.go** receives `AwsEventBridgeBusStackInput` containing the target resource and provider config
2. **locals.go** constructs AWS tags from metadata and stores spec references
3. **event_bus.go** creates the `cloudwatch.EventBus` resource with:
   - Bus name derived from `metadata.name`
   - Optional description
   - Optional KMS encryption via `kms_key_identifier`
   - Optional partner event source via `event_source_name`
   - Optional dead letter config (SQS queue ARN)
   - Optional log config (level and include_detail)
4. **outputs.go** defines constants for stack output keys: `bus_name`, `bus_arn`

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **Conditional nested blocks**: Dead letter config and log config are only set when the corresponding spec fields are non-nil.
- **Simple resource**: EventBridge bus is one of the simplest AWS resources — a single `cloudwatch.NewEventBus` call with optional nested blocks.
