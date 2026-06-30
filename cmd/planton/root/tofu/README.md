# OpenTofu Commands - Technical Reference

> **📖 For the best reading experience**, view the comprehensive guide on the [Planton website](https://planton.dev/docs/cli/tofu-commands).

Technical reference for the `planton tofu` command group.

---

## Overview

The `tofu` command group provides a manifest-driven interface to OpenTofu/Terraform operations. It wraps standard OpenTofu commands (`init`, `plan`, `apply`, `refresh`, `destroy`) with Planton's consistent workflow: manifest loading, validation, and credential injection.

## Command Structure

```
planton tofu
├── init         Initialize backend and providers
├── plan         Preview infrastructure changes
├── apply        Deploy infrastructure changes
├── refresh      Sync state with cloud reality
└── destroy      Teardown infrastructure
```

## Architecture

### Request Flow

```
1. CLI parses flags (--manifest, --set, --kustomize-dir, etc.)
   ↓
2. Manifest resolution (priority: --manifest > --kustomize-dir + --overlay)
   ↓
3. Value overrides applied (--set flags modify manifest)
   ↓
4. Manifest validation (proto-validate checks)
   ↓
5. Credential extraction (from flags or manifest)
   ↓
6. Manifest conversion to tfvars (YAML → variables.tf.json)
   ↓
7. Environment variable injection (provider credentials)
   ↓
8. OpenTofu execution (via tofumodule.RunCommand)
   ↓
9. Stream output to user
```

### Key Implementation Details

**Manifest to TFVars Conversion**:
The CLI converts Planton manifests (YAML) into Terraform variables (`variables.tf.json`) that modules consume. This happens in `pkg/iac/tofu/tofumodule/`.

**Credential Injection**:
Provider credentials can be provided via:
1. CLI flags (`--aws-credential`, `--gcp-credential`, etc.)
2. Environment variables (AWS_ACCESS_KEY_ID, etc.)

Credentials are injected as environment variables before OpenTofu execution.

**Module Discovery**:
The `--module-dir` flag specifies where OpenTofu code lives. Defaults to current directory but typically points to IaC modules in the Planton repository.

## Command Details

### `init`

**Purpose**: Initialize backend and download provider plugins.

**Handler**: `init.go -> initHandler()`

**Key Operations**:
- Loads and validates manifest
- Sets up module directory
- Runs `tofu init` to:
  - Initialize backend (S3, GCS, local, etc.)
  - Download provider plugins
  - Create `.terraform` directory

**Idempotent**: Yes - safe to run multiple times.

### `plan`

**Purpose**: Create execution plan showing proposed changes.

**Handler**: `plan.go -> planHandler()`

**Key Operations**:
- Loads, validates, and applies overrides to manifest
- Converts manifest to tfvars
- Injects credentials
- Runs `tofu plan`
- Optionally creates destroy plan (`--destroy` flag)

**Modifies Resources**: No
**Modifies State**: No

### `apply`

**Purpose**: Deploy infrastructure changes.

**Handler**: `apply.go -> applyHandler()`

**Key Operations**:
- Loads, validates, and applies overrides to manifest
- Converts manifest to tfvars
- Injects credentials
- Runs `tofu apply` (shows plan unless `--auto-approve`)
- Updates state file with changes

**Modifies Resources**: Yes
**Modifies State**: Yes

**Auto-approval**: Use `--auto-approve` flag for non-interactive mode (CI/CD).

### `refresh`

**Purpose**: Update state file to match cloud reality.

**Handler**: `refresh.go -> refreshHandler()`

**Key Operations**:
- Loads and validates manifest
- Queries cloud provider for current resource state
- Updates state file to reflect actual values
- Does NOT modify resources

**Modifies Resources**: No
**Modifies State**: Yes (synchronizes with reality)

### `destroy`

**Purpose**: Delete all managed infrastructure.

**Handler**: `destroy.go -> destroyHandler()`

**Key Operations**:
- Loads, validates, and applies overrides to manifest
- Creates destroy plan
- Waits for confirmation (unless `--auto-approve`)
- Deletes resources in reverse dependency order
- Updates state to reflect deletion

**Modifies Resources**: Yes (deletes them)
**Modifies State**: Yes
**Reversible**: No

## Shared Flags

All commands inherit flags from `root/tofu.go`:

- `--manifest`: Path to manifest file
- `--kustomize-dir` + `--overlay`: Kustomize-based manifest generation
- `--module-dir`: Override module directory
- `--set`: Runtime manifest value overrides (key=value pairs)
- `--aws-credential`, `--gcp-credential`, etc.: Provider credential files

Command-specific flags:
- `plan`: `--destroy` (create destroy plan)
- `apply`: `--auto-approve` (skip confirmation)
- `destroy`: `--auto-approve` (skip confirmation)

## Integration Points

### With Internal Packages

- `internal/manifest`: Manifest loading, validation, and override application
- `internal/cli/manifest`: Manifest path resolution (kustomize, direct, input-dir)
- `internal/cli/cliprint`: Consistent CLI output formatting

### With PKG Packages

- `pkg/iac/tofu/tofumodule`: Core OpenTofu execution logic
- `pkg/iac/stackinput/stackinputproviderconfig`: Credential extraction and injection
- `pkg/iac/tofu/tfvars`: Manifest-to-tfvars conversion
- `pkg/kustomize/builder`: Kustomize manifest building

## Error Handling

All handlers follow the pattern:
1. Parse and validate flags
2. Load and validate manifest
3. Prepare execution (credentials, tfvars)
4. Hand off to OpenTofu
5. Exit with error code on failure

**Exit codes**:
- `0`: Success
- `1`: General error (manifest validation, OpenTofu execution, etc.)

## Development Notes

### Adding New Commands

To add a new `tofu` subcommand:

1. Create `cmd/planton/root/tofu/<command>.go`
2. Define cobra.Command with Use, Short, and Run handler
3. Add command to `root/tofu.go` init function
4. Implement handler following existing patterns

### Testing Commands Locally

```bash
# Build CLI
make build

# Run command with local module
./bin/planton tofu plan \
  --manifest test-manifest.yaml \
  --module-dir /path/to/local/module
```

### Debugging

Enable OpenTofu verbose logging:

```bash
export TF_LOG=DEBUG
planton tofu plan --manifest resource.yaml
```

## Dependencies

- OpenTofu CLI (`tofu` binary must be in PATH)
- Git (for module cloning, if using external modules)
- Provider-specific CLIs (for credential validation, optional)

## Related Documentation

- [OpenTofu Commands Guide](https://planton.dev/docs/cli/tofu-commands) - User-facing comprehensive guide
- [Manifest Structure](/docs/guides/manifests) - Understanding manifests
- [IaC Module Structure](../../pkg/iac/tofu/tofumodule/README.md) - How modules are executed

## Contributing

When modifying tofu commands:
- Maintain consistent error handling
- Use `cliprint` for output formatting
- Validate manifests before OpenTofu execution
- Add integration tests for new functionality
- Update both code README and website documentation

