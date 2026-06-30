---
title: "Configuration & Utilities"
description: "CLI configuration, manifest validation, manifest loading, and version management"
icon: "gear"
order: 80
---

# Configuration & Utilities

This page covers the utility commands that support the deployment workflow: configuration management, manifest validation, manifest loading, and version checking.

## config

Manage CLI configuration stored at `~/.planton/config.yaml`.

### config set

Set a configuration value.

```bash
planton config set backend-url https://api.example.com
```

| Argument | Description |
|----------|-------------|
| `key` | Configuration key to set |
| `value` | Value to assign |

Available keys:

| Key | Validation | Description |
|-----|-----------|-------------|
| `backend-url` | Must start with `http://` or `https://` | Backend API URL for remote operations |

### config get

Retrieve a configuration value.

```bash
planton config get backend-url
```

Prints the value to stdout. Exits with code 1 if the key is not set.

### config list

List all configuration values.

```bash
planton config list
```

Prints all set configuration values in `key=value` format. If no values are set, prints "No configuration values set".

## validate-manifest

Validate a manifest against its Protocol Buffer schema without deploying. This runs the same validation that deployment commands perform before execution, allowing you to catch errors early.

**Aliases**: `validate`

```bash
planton validate -f database.yaml
planton validate database.yaml                                # positional arg
planton validate --clipboard                                  # from clipboard
planton validate --kustomize-dir services/api --overlay prod  # from kustomize
```

The command accepts a manifest from any source: file path (as a flag or positional argument), clipboard, or kustomize build. See [Manifest Source Flags](./cli-reference#manifest-source-flags) in the CLI Reference for all input options.

Validation checks the manifest against the component's protobuf schema using `protovalidate`. This catches:

- Missing required fields
- Invalid field values (wrong types, out-of-range values)
- Constraint violations defined in `buf.validate` annotations
- Unrecognized `apiVersion` or `kind`

On success, prints a confirmation message and exits with code 0. On failure, prints detailed validation errors and exits with code 1.

## load-manifest

Load a manifest, apply defaults and overrides, and print the resolved result. This is useful for inspecting what the CLI will actually see after all processing is applied.

**Aliases**: `load`

```bash
planton load -f database.yaml
planton load database.yaml                                    # positional arg
planton load --clipboard                                      # from clipboard
planton load -f api.yaml --set spec.container.replicas=5      # with overrides
planton load --kustomize-dir services/api --overlay prod      # from kustomize
```

The command:

1. Resolves the manifest from the specified source
2. Applies `--set` overrides if provided
3. Fills in protobuf default values
4. Prints the fully resolved manifest as YAML to stdout

This is the same processing pipeline that deployment commands use before handing off to the IaC engine. Use it to verify that your overrides, kustomize builds, and manifest sources produce the expected result.

### Flags

Accepts all [manifest source flags](./cli-reference#manifest-source-flags) (`-f`, `-c`, `-i`, `--kustomize-dir`, `--overlay`) plus:

| Flag | Description |
|------|-------------|
| `--set` | Override manifest values using `key=value` pairs (repeatable) |

## Manifest Source Resolution

Both `validate-manifest` and `load-manifest` use the same manifest resolution logic as deployment commands. When multiple source flags are provided, the CLI uses the first match in priority order:

| Priority | Flag | Description |
|----------|------|-------------|
| 1 | `--clipboard` (`-c`) | Read manifest YAML from the system clipboard |
| 2 | `--stack-input` (`-i`) | Extract manifest from the `target` field of a stack input YAML file |
| 3 | `--manifest` (`-f`) | Read manifest from a file path or URL |
| 4 | `--input-dir` | Use `{dir}/target.yaml` as the manifest |
| 5 | `--kustomize-dir` + `--overlay` | Build manifest from kustomize configuration (both flags required) |

A positional argument (e.g., `planton validate database.yaml`) is also accepted for backward compatibility. When a positional argument is provided, it takes precedence over all flags.

URL manifests (paths starting with `http://` or `https://`) are downloaded to `~/.planton/downloads/` and cached locally.

## version

Display the current CLI version and check for available updates.

```bash
planton version
planton -v
planton --version
```

The command prints the current version, fetches the latest available version from GitHub releases, and indicates whether an update is available. If a newer version exists, it suggests running `planton upgrade`.

## What's Next

- **[CLI Reference](./cli-reference)** — Complete flag reference for all commands
- **[Module Management](./module-management)** — Module versioning, staging area, and CLI upgrades
- **[Unified Commands](./unified-commands)** — Provisioner-agnostic deployment commands
