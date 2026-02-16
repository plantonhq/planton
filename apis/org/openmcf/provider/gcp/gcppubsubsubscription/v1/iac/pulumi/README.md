# GcpPubSubSubscription -- Pulumi Module

This directory contains the Pulumi Go implementation for the GcpPubSubSubscription component.

## Structure

```
main.go                     -- Pulumi entrypoint (loads stack input, calls module)
module/
  main.go                   -- Resources() function (orchestrates resource creation)
  locals.go                 -- Locals struct initialization (labels, provider config)
  outputs.go                -- Output constant names
  subscription.go           -- pubsub.NewSubscription (all delivery methods)
```

## What It Creates

- `pubsub.NewSubscription("pubsub-subscription")` -- A Pub/Sub subscription supporting:
  - Pull delivery (default)
  - Push delivery with OIDC auth and unwrapped payloads
  - BigQuery streaming delivery
  - Cloud Storage batch delivery
  - Dead-letter policy
  - Retry policy with configurable backoff
  - Expiration policy
  - Message ordering and exactly-once delivery
  - Attribute-based filtering

## Running Locally

```bash
# Build
cd ~/scm/github.com/plantonhq/openmcf
go build ./apis/org/openmcf/provider/gcp/gcppubsubsubscription/v1/...

# Test
go test -v ./apis/org/openmcf/provider/gcp/gcppubsubsubscription/v1/

# Debug with manifest
cd apis/org/openmcf/provider/gcp/gcppubsubsubscription/v1/iac/pulumi
./debug.sh
```

## Outputs

| Output | Description |
|--------|-------------|
| `subscription_id` | Fully qualified subscription ID |
| `subscription_name` | Short subscription name |
