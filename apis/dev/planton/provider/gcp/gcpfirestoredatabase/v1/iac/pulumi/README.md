# GcpFirestoreDatabase - Pulumi Implementation

## Quick Start

```bash
# From the component directory
cd apis/dev/planton/provider/gcp/gcpfirestoredatabase/v1/iac/pulumi

# Preview changes
pulumi preview --stack dev

# Apply changes
pulumi up --stack dev
```

## Local Development

```bash
# Build the module
go build ./module/...

# Run tests
go test -v ./module/...
```

## Module Overview

See [overview.md](overview.md) for architecture and design details.
