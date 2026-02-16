# GcpPubSubSubscription -- Pulumi Module Overview

## Architecture

```
main.go (entrypoint)
  └── module/
       ├── main.go          -- Resources() orchestrator
       ├── locals.go         -- Locals struct with labels, provider config
       ├── outputs.go        -- Output constant names
       └── subscription.go   -- pubsub.NewSubscription with conditional configs
```

## Resource Graph

The module provisions a single Pub/Sub subscription with conditional configuration
blocks based on the delivery method chosen:

```
pubsub.NewSubscription("pubsub-subscription")
  ├── Pull delivery (default, no extra config)
  ├── Push delivery (SubscriptionPushConfigArgs)
  │   ├── OidcToken (optional)
  │   └── NoWrapper (optional)
  ├── BigQuery delivery (SubscriptionBigqueryConfigArgs)
  ├── Cloud Storage delivery (SubscriptionCloudStorageConfigArgs)
  │   └── AvroConfig (optional)
  ├── DeadLetterPolicy (optional)
  ├── RetryPolicy (optional)
  └── ExpirationPolicy (optional)
```

## Key Implementation Details

1. **Delivery method selection**: The spec's `push_config`, `bigquery_config`, and
   `cloud_storage_config` fields are nil-checked. Only the non-nil config is set
   on the subscription args. If all three are nil, the subscription uses pull delivery.

2. **StringValueOrRef resolution**: All cross-resource reference fields use
   `.GetValue()` to resolve the literal value. This supports both literal values
   and `valueFrom` references resolved by the OpenMCF platform.

3. **Labels**: Framework labels are applied from `locals.GcpLabels` (resource kind,
   name, org, env, id).

4. **Outputs**: The subscription's Pulumi ID (fully qualified path) and name are
   exported for downstream resources.

## Dependencies

| Import | Purpose |
|--------|---------|
| `pulumi-gcp/sdk/v9/go/gcp/pubsub` | Subscription resource and config types |
| `openmcf/.../gcppubsubsubscription/v1` | Proto-generated spec types |
| `openmcf/pkg/.../pulumigoogleprovider` | GCP provider setup |
| `openmcf/pkg/.../stackinput` | Stack input loading |
