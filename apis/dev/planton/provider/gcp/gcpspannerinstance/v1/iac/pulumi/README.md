# GcpSpannerInstance Pulumi Module

This Pulumi module provisions a Google Cloud Spanner instance.

## Structure

```
module/
  main.go                # Entry point - orchestrates resource creation
  locals.go              # Variable transformations and label computation
  outputs.go             # Output constant definitions
  spanner_instance.go    # Spanner instance resource creation
```

## Development

```bash
# Build
cd ~/scm/github.com/plantonhq/planton
go build ./apis/dev/planton/provider/gcp/gcpspannerinstance/v1/...

# Test
go test -v ./apis/dev/planton/provider/gcp/gcpspannerinstance/v1/

# Debug with local manifest
cd apis/dev/planton/provider/gcp/gcpspannerinstance/v1/iac/pulumi
./debug.sh
```

## Outputs

| Name | Description |
|---|---|
| `instance_id` | Fully qualified instance ID |
| `instance_name` | Short instance name |
| `state` | Instance state (CREATING or READY) |
