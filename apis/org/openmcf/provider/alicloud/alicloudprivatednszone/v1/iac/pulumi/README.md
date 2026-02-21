# AlicloudPrivateDnsZone -- Pulumi Module

This directory contains the Pulumi (Go) implementation for the AlicloudPrivateDnsZone deployment component.

## Structure

```
pulumi/
├── main.go          # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml      # Pulumi project configuration
├── module/
│   ├── main.go      # Creates Zone + ZoneAttachment, exports outputs
│   ├── records.go   # Creates ZoneRecord resources for each spec.records entry
│   ├── locals.go    # Tag computation and helper functions
│   └── outputs.go   # Output key constants
```

## Resources Created

1. **pvtz.Zone** -- the private DNS hosted zone
2. **pvtz.ZoneAttachment** -- binds the zone to specified VPCs (supports cross-region)
3. **pvtz.ZoneRecord** -- one per record in `spec.records`

## Local Development

```bash
cd module/
go build ./...
go vet ./...
```

## Debug

```bash
export ALICLOUD_ACCESS_KEY="your-key"
export ALICLOUD_SECRET_KEY="your-secret"
pulumi up --stack dev
```
