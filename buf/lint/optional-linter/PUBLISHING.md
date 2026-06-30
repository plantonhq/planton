# Publishing the Optional Linter to BSR

## Publication Summary

The `optional-linter` plugin has been successfully published to the Buf Schema Registry (BSR) under the `planton` organization.

### Plugin Details

- **Name**: `buf.build/planton/optional-linter`
- **Version**: v0.1.0
- **Type**: check (linting plugin)
- **Visibility**: public
- **URL**: https://buf.build/planton/optional-linter
- **Commit**: 9c8ab12a043245c08a92d7657c0da20a

### Published Labels

The plugin was published with the following labels:
- `v0.1.0` - Exact patch version (pin to specific release)
- `v0.1` - Minor version (receive patch updates)
- `v0` - Major version (receive minor and patch updates)
- `main` - Latest version (always points to most recent publish)

## Usage

### In buf.yaml (for consumers)

```yaml
version: v2
plugins:
  - plugin: buf.build/planton/optional-linter:v0.1.0
lint:
  use:
    - STANDARD
    - DEFAULT_REQUIRES_OPTIONAL
```

### Updating the Plugin Reference

After adding the plugin to `buf.yaml`, consumers must run:

```bash
buf plugin update
```

This updates the `buf.lock` file with the specific commit digest.

## Publishing New Versions

To publish a new version of the plugin:

```bash
cd buf/lint/optional-linter
make publish version=v0.2.0
```

This will:
1. Compile the plugin to WebAssembly
2. Push to BSR with semantic version labels
3. Create the plugin if it doesn't exist (for first-time publishing)

### Version Naming Convention

Follow semantic versioning:
- **Patch** (v0.1.1): Bug fixes, no API changes
- **Minor** (v0.2.0): New features, backward compatible
- **Major** (v1.0.0): Breaking changes

## Local Development

For local development and testing, you can still use the native binary:

```bash
# Build local binary
make build

# This places the binary in $GOPATH/bin/optional-linter
# Update apis/buf.yaml to use: plugin: optional-linter
```

## Verification

To verify the published plugin:

```bash
# Check plugin info
buf registry plugin info buf.build/planton/optional-linter

# Test linting with the plugin
cd apis
buf lint
```

## Files Changed

- Renamed: `cmd/buf-plugin-planton/` → `cmd/optional-linter/`
- Created: `buf.plugin.yaml` (plugin metadata)
- Created: `Makefile` (build and publish automation)
- Updated: `go.mod` (module path)
- Updated: `apis/buf.yaml` (plugin reference)
- Updated: `apis/Makefile` (build integration)
- Updated: `README.md` (documentation)

## Benefits of BSR Publishing

1. **Remote Execution**: No need to install the plugin locally
2. **Version Pinning**: Exact version control via buf.lock
3. **Reproducible Builds**: Same plugin version across all environments
4. **Discoverability**: Plugin visible in BSR catalog
5. **Security**: WebAssembly sandboxing for safe execution

## Notes

- The plugin is published as **public**, making it available to anyone
- The WASM binary is compiled targeting `wasip1` (WASI 0.1 specification)
- Local development still works with native binaries via `make build`
- The plugin only validates the `DEFAULT_REQUIRES_OPTIONAL` rule (scoped, not extensible)

