# Pulumi Backend Config Package

This package provides functionality to extract Pulumi backend configuration from Planton resource manifests using standardized annotations.

## Overview

The `backendconfig` package implements the logic to read and parse Pulumi backend configuration from manifest annotations. It supports both a simplified single-annotation approach (stack FQDN) and a detailed multi-annotation approach, with intelligent prioritization between them.

## Core Types

### PulumiBackendConfig

```go
type PulumiBackendConfig struct {
    StackFqdn    string  // Full stack identifier: "org/project/stack"
    Organization string  // Pulumi organization name
    Project      string  // Pulumi project name  
    StackName    string  // Pulumi stack name
}
```

## Key Functions

### ExtractFromManifest

```go
func ExtractFromManifest(manifest proto.Message) (*PulumiBackendConfig, error)
```

Extracts Pulumi backend configuration from a manifest's metadata annotations.

**Priority Logic:**
1. If `stack.fqdn` annotation is present, it takes precedence
2. If not, all three component annotations must be present
3. Returns an error if neither approach provides complete configuration

**Example Usage:**

```go
import (
    "github.com/plantonhq/planton/pkg/iac/pulumi/backendconfig"
)

// Extract backend config from a manifest
config, err := backendconfig.ExtractFromManifest(awsVpcManifest)
if err != nil {
    // Handle error - no valid backend config in manifest
    return err
}

// Use the extracted configuration
fmt.Printf("Stack: %s\n", config.StackFqdn)
```

## Label Processing

### Stack FQDN Parsing

When a `stack.fqdn` annotation is provided, it's automatically parsed into its components:

```
"demo-org/aws-infrastructure/production" 
    ↓
Organization: "demo-org"
Project:      "aws-infrastructure"
StackName:    "production"
```

The parser:
- Validates the format (must have exactly 3 components)
- Trims whitespace from each component
- Ensures no component is empty

### Validation Rules

1. **Stack FQDN Format**: Must be `organization/project/stack`
2. **Required Annotations**: Either stack.fqdn OR all three component annotations
3. **Non-Empty Values**: All annotation values must be non-empty strings
4. **No Partial Config**: Cannot specify only some component annotations

## Error Handling

The package provides detailed error messages for common issues:

```go
// Missing annotations
"no annotations found in manifest"

// Invalid FQDN format
"invalid stack.fqdn format: stack FQDN must be in format 'organization/project/stack'"

// Missing required annotations
"missing required Pulumi backend annotations: need either pulumi.planton.dev/stack.fqdn or all of (organization, project, stack.name)"

// Empty values
"Pulumi backend annotations cannot be empty"
```

## Testing

The package includes comprehensive tests covering:
- Stack FQDN precedence over component annotations
- FQDN parsing with various formats
- Error cases (missing annotations, invalid formats, empty values)
- Edge cases (spaces in FQDN, empty components)

Run tests:
```bash
go test ./pkg/iac/pulumi/backendconfig -v
```

## Integration Example

Here's how the CLI might use this package:

```go
// Load manifest
manifest, err := loadManifest(manifestPath)
if err != nil {
    return err
}

// Extract backend config from manifest
manifestConfig, err := backendconfig.ExtractFromManifest(manifest)
if err != nil {
    // No backend config in manifest, fall back to CLI flags
    manifestConfig = nil
}

// Determine final stack FQDN
var stackFqdn string
if manifestConfig != nil {
    stackFqdn = manifestConfig.StackFqdn
} else if flagStackFqdn != "" {
    stackFqdn = flagStackFqdn
} else {
    return errors.New("no stack configuration provided")
}

// Use stackFqdn for Pulumi operations
```

## Design Decisions

1. **Proto-Agnostic**: Uses `proto.Message` interface to work with any manifest type
2. **Clear Precedence**: Stack FQDN always wins over component annotations
3. **Fail-Fast Validation**: Returns errors immediately for invalid configurations
4. **Nil-Safe**: Returns nil for manifests without metadata or annotations

## Related Packages

- `pkg/iac/pulumi/pulumiannotationkeys`: Defines the annotation key constants
- `pkg/reflection/metadatareflect`: Provides annotation extraction from protobuf messages
- `pkg/iac/pulumi/pulumistack`: Consumes the extracted configuration
