# Manifest Package - Technical Reference

Technical reference for the `internal/manifest` package, which handles loading, validation, and manipulation of OpenMCF manifests.

---

## Overview

The `manifest` package is responsible for:
1. Loading YAML manifests from files or URLs
2. Converting YAML to Protocol Buffer messages
3. Applying default values from proto field options
4. Validating manifests using proto-validate rules
5. Applying runtime value overrides (--set flags)
6. Extracting spec from manifests for validation

## Package Structure

```
internal/manifest/
├── load_manifest.go        # Core loading logic
├── manifest_validator.go   # Validation using proto-validate
├── values_setter.go         # Runtime override application
├── spec_extractor.go        # Spec extraction for validation
├── print.go                 # Manifest pretty-printing
├── protodefaults/           # Default value application
│   ├── applier.go
│   └── README.md
└── manifestprotobuf/        # Protobuf field manipulation
    └── field_setter.go
```

---

## Core Functions

### LoadManifest

**Signature**: `func LoadManifest(manifestPath string) (proto.Message, error)`

**Purpose**: Loads a manifest from a file or URL and returns it as a typed Protocol Buffer message.

**Flow**:

```
1. Check if manifestPath is a URL
   ↓ (if URL)
2. Download to temporary file
   ↓
3. Read YAML file
   ↓
4. Convert YAML → JSON
   ↓
5. Extract kind from manifest
   ↓
6. Look up proto message type for kind
   ↓
7. Unmarshal JSON into proto message
   ↓
8. Apply defaults from proto field options
   ↓
9. Return typed proto message
```

**Usage**:

```go
import "github.com/plantonhq/openmcf/internal/manifest"

// Load manifest
msg, err := manifest.LoadManifest("ops/resources/database.yaml")
if err != nil {
    return err
}

// msg is a proto.Message - type assert to specific type if needed
```

**Error Scenarios**:
- File not found
- Invalid YAML syntax
- Unsupported kind
- Failed proto unmarshaling
- Default application errors

### Validate

**Signature**: `func Validate(manifestPath string) error`

**Purpose**: Validates a manifest using proto-validate rules without deploying it.

**Flow**:

```
1. Load manifest (calls LoadManifest)
   ↓
2. Extract spec from manifest
   ↓
3. Initialize proto-validate validator
   ↓
4. Run validation on spec
   ↓
5. Format validation errors (if any)
   ↓
6. Return error or nil
```

**Usage**:

```go
// Validate before deployment
if err := manifest.Validate("database.yaml"); err != nil {
    fmt.Println(err)  // Pretty-printed validation errors
    return err
}
```

**Validation Rules**:
- Field-level constraints (e.g., `replicas` between 1-10)
- String patterns (e.g., CPU format `^[0-9]+m$`)
- Required field checks
- Cross-field validation (using CEL expressions)

### LoadWithOverrides

**Signature**: `func LoadWithOverrides(manifestPath string, valueOverrides map[string]string) (proto.Message, error)`

**Purpose**: Loads a manifest and applies runtime value overrides (from `--set` flags).

**Flow**:

```
1. Load manifest (calls LoadManifest)
   ↓
2. For each override (key=value):
   a. Parse key path (e.g., "spec.replicas")
   b. Navigate to field in proto message
   c. Convert value string to appropriate type
   d. Set field value
   ↓
3. Return modified proto message
```

**Usage**:

```go
overrides := map[string]string{
    "spec.replicas":                "5",
    "spec.container.image.tag":     "v2.0.0",
    "metadata.labels.environment":  "staging",
}

msg, err := manifest.LoadWithOverrides("deployment.yaml", overrides)
```

**Supported Override Paths**:
- Nested fields: `spec.container.resources.limits.cpu`
- Map fields: `metadata.labels.key`
- Repeated/list fields: Not directly supported (override entire list)

### ApplyOverridesToFile

**Signature**: `func ApplyOverridesToFile(manifestPath string, valueOverrides map[string]string) (string, bool, error)`

**Purpose**: Applies overrides and writes result to a new temp file (used by CLI commands).

**Returns**:
- `string`: Path to result file (original if no overrides, temp file if overrides applied)
- `bool`: True if temp file created (caller should clean up)
- `error`: Any error encountered

**Usage**:

```go
finalPath, isTemp, err := manifest.ApplyOverridesToFile(
    "deployment.yaml",
    map[string]string{"spec.replicas": "3"},
)
if err != nil {
    return err
}
if isTemp {
    defer os.Remove(finalPath)  // Clean up temp file
}

// Use finalPath for deployment
```

---

## Manifest Loading Pipeline

### Detailed Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Input: YAML file or URL                                     │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 1: URL Detection                                        │
│ - Parse manifest path as URL                                │
│ - If URL: download to temp file via http.Get                │
│ - If file: use path directly                                │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: YAML Reading                                         │
│ - Read file contents (os.ReadFile)                          │
│ - Store as []byte                                            │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: YAML → JSON Conversion                               │
│ - Convert YAML to JSON (yaml.YAMLToJSON)                    │
│ - JSON is easier to unmarshal into protos                   │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Kind Extraction                                      │
│ - Parse manifest to extract `kind` field                    │
│ - Use crkreflect to map kind string to CloudResourceKind    │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 5: Proto Message Lookup                                 │
│ - Look up proto message type for kind                       │
│ - Use crkreflect.ToMessageMap[kind]                         │
│ - Error if kind not supported                               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 6: JSON → Proto Unmarshaling                            │
│ - Unmarshal JSON into proto message                         │
│ - Uses protojson.Unmarshal                                   │
│ - Type-safe conversion                                       │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 7: Default Value Application                            │
│ - Scan proto for fields with default annotations            │
│ - Apply defaults to unset optional fields                   │
│ - Uses protodefaults.ApplyDefaults                          │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Output: Typed proto.Message                                  │
└─────────────────────────────────────────────────────────────┘
```

---

## Kind Detection and Resolution

### How Kind Resolution Works

The package uses the `crkreflect` (Cloud Resource Kind Reflection) system:

```go
// Extract kind string from YAML
kindName, err := crkreflect.ExtractKindFromTargetManifest(manifestPath)
// kindName = "AwsS3Bucket"

// Convert string to CloudResourceKind enum
cloudResourceKind := crkreflect.KindFromString(kindName)
// cloudResourceKind = CloudResourceKind_aws_s3_bucket

// Look up proto message type
manifest := crkreflect.ToMessageMap[cloudResourceKind]
// manifest = &AwsS3Bucket{}
```

**Supported kinds**: See `pkg/crkreflect/kind_map_gen.go` for complete mapping.

**Unsupported kind error**: If kind isn't in the map, returns formatted error:

```
Unsupported cloud resource kind: UnknownKind

Available kinds: AwsS3Bucket, GcpGkeCluster, PostgresKubernetes, ...
```

---

## Default Value Application

Defaults are applied from proto field options. See `protodefaults/README.md` for detailed explanation.

**Example proto definition**:

```protobuf
message ExternalDnsKubernetesSpec {
  optional string namespace = 1 [(org.openmcf.shared.options.default) = "external-dns"];
  optional string version = 2 [(org.openmcf.shared.options.default) = "v0.19.0"];
}
```

**Behavior**:
- If field is unset (nil pointer for optional fields), apply default
- If field is explicitly set (even to zero value), preserve user's value
- Only works for scalar types (string, int32, bool, etc.)

**Why optional is required**: Proto3 requires `optional` keyword for proper field presence detection. Without it, can't distinguish "user set to zero" from "user didn't set at all."

---

## Validation Architecture

### Proto-Validate Integration

Validation uses the `buf.build/go/protovalidate` library with rules defined in proto files:

```protobuf
message PostgresKubernetesSpec {
  int32 replicas = 1 [(buf.validate.field).int32 = {gte: 1, lte: 10}];
  string cpu = 2 [(buf.validate.field).string.pattern = "^[0-9]+m$"];
}
```

### Validation Error Formatting

The package formats validation errors in a user-friendly way:

```
╔═══════════════════════════════════════════════════════════╗
║                 ❌  MANIFEST VALIDATION FAILED            ║
╚═══════════════════════════════════════════════════════════╝

⚠️  Validation Errors:

spec.replicas: value must be >= 1 and <= 10 (got: 0)
spec.cpu: value must match pattern "^[0-9]+m$" (got: "invalid")
```

Color-coded output uses `github.com/fatih/color`.

---

## Runtime Value Overrides

### How --set Works

The `--set` flag allows runtime overrides:

```bash
openmcf pulumi up \
  --manifest deployment.yaml \
  --set spec.replicas=5 \
  --set spec.container.image.tag=v2.0.0
```

**Implementation**:

```go
// manifestprotobuf/field_setter.go
func SetProtoField(msg proto.Message, path string, value string) (proto.Message, error) {
    // 1. Parse path (e.g., "spec.replicas" → ["spec", "replicas"])
    // 2. Navigate proto message structure using reflection
    // 3. Convert string value to appropriate type
    // 4. Set field value
    // 5. Return modified message
}
```

**Limitations**:
- Can set scalar fields (string, int, bool, etc.)
- Can set message fields (creates if needed)
- Cannot set repeated fields directly (must override entire list)
- Cannot unset fields (set to empty/zero, not nil)

---

## Error Handling

### Error Types

**1. File Not Found**:
```go
return nil, errors.Wrapf(err, "failed to read manifest file %s", manifestPath)
```

**2. Invalid YAML**:
```go
return nil, errors.Wrap(err, "failed to load yaml to json")
```

**3. Unsupported Kind**:
```go
return nil, formatUnsupportedResourceError(kindName)
```

**4. Validation Errors**:
```go
return formatValidationError(validationErr)
```

### Error Formatting

Uses `github.com/pkg/errors` for error wrapping, providing full context:

```
failed to load manifest: failed to read manifest file ops/db.yaml: no such file or directory
```

---

## Integration Points

### With CLI Commands

All CLI commands (pulumi, tofu, validate, load-manifest) use this package:

```go
// pulumi/up.go, tofu/apply.go, etc.
import "github.com/plantonhq/openmcf/internal/manifest"

// Load manifest
manifest, err := manifest.LoadManifest(manifestPath)

// Apply overrides
manifest, err := manifest.LoadWithOverrides(manifestPath, overrides)

// Validate
if err := manifest.Validate(manifestPath); err != nil {
    // Handle validation failure
}
```

### With IaC Modules

Loaded manifests are passed to IaC modules (Pulumi, OpenTofu) as environment variables or tfvars files.

**For Pulumi**:
- Manifest marshaled to JSON
- Set as `OPENMCF_MANIFEST` environment variable
- Module reads and unmarshals

**For OpenTofu**:
- Manifest converted to Terraform variables
- Written to `variables.tf.json`
- Module reads as input variables

---

## Testing

### Unit Tests

Key test files:
- `load_manifest_test.go`: Tests manifest loading from files and URLs
- `values_setter_test.go`: Tests runtime overrides
- `manifest_validator_test.go`: Tests validation logic

### Running Tests

```bash
# Run all tests
go test ./internal/manifest/...

# Run with coverage
go test -cover ./internal/manifest/...

# Verbose output
go test -v ./internal/manifest/...
```

---

## Performance Considerations

### Caching

Manifest loading is not cached—each operation loads fresh. This ensures:
- File changes are always reflected
- No stale data issues
- Simpler implementation

For repeated operations, CLI layer handles temp file reuse.

### URL Downloads

URL manifests are downloaded once per operation:
- Temporary file created in OS temp directory
- Cleaned up after operation (unless error)
- No persistent cache

---

## Development Guidelines

### Adding Support for New Kinds

1. Define proto message in `apis/` directory
2. Add to `crkreflect` kind mapping (auto-generated via `make generate-crk-reflect`)
3. No changes needed in `manifest` package (automatic)

### Modifying Validation

1. Update proto definitions with buf-validate rules
2. No code changes needed (rules applied automatically)
3. Test with `openmcf validate`

### Adding New Override Paths

The `manifestprotobuf.SetProtoField` function supports any valid proto path:
- Nested fields: Automatically navigates message hierarchy
- New field types: Add type conversion in `field_setter.go`

---

## Related Documentation

- [Manifest Structure Guide](/docs/guides/manifests) - User-facing manifest documentation
- [Proto Defaults README](./protodefaults/README.md) - Default value system
- [CRK Reflect Package](../../pkg/crkreflect/README.md) - Kind resolution system

---

## Debugging

### Enable Debug Logging

```go
import log "github.com/sirupsen/logrus"

log.SetLevel(log.DebugLevel)

// Now manifest operations will log debug info
manifest, err := manifest.LoadManifest("resource.yaml")
```

### Common Issues

**"Unsupported kind"**: 
- Verify kind spelling (case-sensitive)
- Check `pkg/crkreflect/kind_map_gen.go` for supported kinds
- Run `make generate-crk-reflect` if adding new kind

**"Validation failed"**:
- Run `openmcf validate` for detailed errors
- Check proto definition for validation rules
- Verify field types match expected values

**"Failed to set field"**:
- Verify field path is correct (use proto field names, not YAML names)
- Check field type (string value must be convertible)
- Ensure message hierarchy exists

---

## Contributing

When contributing to this package:
- Maintain backwards compatibility
- Add tests for new functionality
- Update this README with significant changes
- Follow existing error handling patterns
- Use `github.com/pkg/errors` for error wrapping

