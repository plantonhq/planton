# Kustomize Builder Package - Technical Reference

Technical reference for the `pkg/kustomize/builder` package, which handles building kustomize manifests for Planton CLI.

---

## Overview

The `builder` package provides a simple wrapper around kustomize's API to build manifests from a base + overlay structure. It's used by CLI commands when the `--kustomize-dir` and `--overlay` flags are provided.

## Core Function

### BuildManifest

**Signature**: `func BuildManifest(kustomizeDir, overlay string) (string, error)`

**Purpose**: Builds a kustomize manifest and writes it to a temporary file.

**Parameters**:
- `kustomizeDir`: Base directory containing kustomize structure (e.g., `./backend/services/api/kustomize`)
- `overlay`: The overlay environment to build (e.g., `prod`, `dev`, `staging`)

**Returns**:
- `string`: Path to temporary manifest file
- `error`: Any error encountered during build

**How it works**:

```
1. Validate inputs (kustomizeDir and overlay must be non-empty)
   ↓
2. Construct kustomization path: kustomizeDir/overlays/overlay
   ↓
3. Verify path exists
   ↓
4. Run kustomize build
   ↓
5. Convert result to YAML
   ↓
6. Write to temp file
   ↓
7. Return temp file path
```

**Usage**:

```go
import "github.com/plantonhq/planton/pkg/kustomize/builder"

// Build manifest
tempFile, err := builder.BuildManifest("services/api/kustomize", "prod")
if err != nil {
    return err
}
defer builder.Cleanup(tempFile)  // Clean up temp file

// Use tempFile for deployment...
```

### Cleanup

**Signature**: `func Cleanup(tempFilePath string) error`

**Purpose**: Removes the temporary manifest file created by BuildManifest.

**Usage**:

```go
// Defer cleanup to ensure temp file is removed
tempFile, err := builder.BuildManifest(kustomizeDir, overlay)
if err != nil {
    return err
}
defer builder.Cleanup(tempFile)
```

---

## Integration with CLI

### How CLI Commands Use This Package

```go
// internal/cli/manifest/resolve.go
import "github.com/plantonhq/planton/pkg/kustomize/builder"

func ResolveManifestPath(cmd *cobra.Command) (string, bool, error) {
    kustomizeDir, _ := cmd.Flags().GetString("kustomize-dir")
    overlay, _ := cmd.Flags().GetString("overlay")
    
    if kustomizeDir != "" && overlay != "" {
        // Build kustomize manifest
        tempPath, err := builder.BuildManifest(kustomizeDir, overlay)
        return tempPath, true, err  // true = is temp file
    }
    
    // ... other manifest sources
}
```

### Directory Structure Expected

The package expects this structure:

```
<kustomizeDir>/
├── base/
│   ├── kustomization.yaml
│   └── resource.yaml
└── overlays/
    ├── dev/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        └── kustomization.yaml
```

The build target is: `<kustomizeDir>/overlays/<overlay>`

---

## Error Handling

### Validation Errors

**Empty kustomizeDir**:
```go
errors.New("kustomize-dir cannot be empty")
```

**Empty overlay**:
```go
errors.New("overlay cannot be empty")
```

**Path doesn't exist**:
```go
errors.Errorf("kustomization path does not exist: %s", kustomizationPath)
```

### Build Errors

**Invalid kustomization**:
```go
errors.Wrapf(err, "failed to run kustomize build for path: %s", kustomizationPath)
```

Common causes:
- Syntax errors in kustomization.yaml
- Referenced files don't exist
- Circular dependencies
- Invalid patches

### File Operation Errors

**Temp file creation**:
```go
errors.Wrap(err, "failed to create temporary file")
```

**Write failure**:
```go
errors.Wrap(err, "failed to write manifest to temporary file")
```

---

## Temporary File Management

### Naming Pattern

Temp files are created with pattern: `kustomize-manifest-*.yaml`

Example: `kustomize-manifest-1234567890.yaml`

### Location

Files are created in the OS temp directory:
- Linux/macOS: `/tmp/`
- Windows: `C:\Users\<user>\AppData\Local\Temp\`

### Lifecycle

1. Created by `BuildManifest()`
2. Used by CLI for deployment
3. Cleaned up by:
   - `Cleanup()` function
   - Defer statements in CLI commands
   - OS cleanup on reboot (if leaked)

### Best Practices

```go
// Always defer cleanup
tempFile, err := builder.BuildManifest(dir, overlay)
if err != nil {
    return err
}
defer builder.Cleanup(tempFile)

// Or use explicit cleanup
tempFile, err := builder.BuildManifest(dir, overlay)
if err != nil {
    return err
}

// ... use tempFile ...

if err := builder.Cleanup(tempFile); err != nil {
    log.Warnf("failed to cleanup temp file: %v", err)
}
```

---

## Dependencies

**External**:
- `sigs.k8s.io/kustomize/api/krusty`: Core kustomize build logic
- `sigs.k8s.io/kustomize/kyaml/filesys`: Virtual filesystem for kustomize

**Internal**:
- `github.com/pkg/errors`: Error wrapping

---

## Testing

### Testing the Builder

```go
// Example test
func TestBuildManifest(t *testing.T) {
    tempFile, err := builder.BuildManifest("testdata/example", "prod")
    if err != nil {
        t.Fatalf("build failed: %v", err)
    }
    defer builder.Cleanup(tempFile)
    
    // Verify temp file exists and contains expected YAML
    content, err := os.ReadFile(tempFile)
    if err != nil {
        t.Fatalf("failed to read temp file: %v", err)
    }
    
    // Assert content matches expectations
}
```

### Manual Testing

```bash
# Create test structure
mkdir -p testdata/kustomize/{base,overlays/test}

# Create base
cat > testdata/kustomize/base/kustomization.yaml <<EOF
resources:
  - deployment.yaml
EOF

cat > testdata/kustomize/base/deployment.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: MicroserviceKubernetes
metadata:
  name: test
EOF

# Create overlay
cat > testdata/kustomize/overlays/test/kustomization.yaml <<EOF
resources:
  - ../../base
EOF

# Test the builder
cd pkg/kustomize/builder
go test -v
```

---

## Performance Considerations

### Build Time

Kustomize builds are fast (typically <100ms) because:
- No network calls
- In-memory processing
- Efficient YAML manipulation

### Temp File Size

Manifests are typically small (< 100KB), so temp file creation is fast and has minimal disk impact.

### Cleanup Importance

Always cleanup temp files to avoid:
- Disk space accumulation
- Clutter in temp directory
- Potential information leakage (if manifests contain sensitive data)

---

## Debugging

### Enable Kustomize Debug Logging

```go
// In your code
import "sigs.k8s.io/kustomize/api/konfig"

konfig.BuiltinPluginLoadingOptions = konfig.BuiltinPluginLoadingOptions.Debug()
```

### Verify Build Output

```bash
# Build manually to see result
cd services/api/kustomize
kustomize build overlays/prod

# Or use Planton's load-manifest command
planton load-manifest \
  --kustomize-dir services/api/kustomize \
  --overlay prod
```

---

## Related Documentation

- [Kustomize Integration Guide](/docs/guides/kustomize) - User-facing guide
- [Official Kustomize Docs](https://kustomize.io/) - Complete kustomize documentation
- [Internal Manifest Package](../../../internal/manifest/README.md) - How manifests are loaded

---

## Contributing

When modifying this package:
- Maintain backwards compatibility with existing kustomize structures
- Add tests for new functionality
- Clean up temp files in all code paths
- Update this README with significant changes
- Follow existing error handling patterns

