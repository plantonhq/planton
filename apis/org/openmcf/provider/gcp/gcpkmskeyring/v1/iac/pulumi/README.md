# GcpKmsKeyRing — Pulumi Module

This directory contains the Pulumi Go implementation for the GcpKmsKeyRing component.

## Module Structure

```
module/
  main.go        — Entry point: initializes locals, provider, and orchestrates resources
  locals.go      — Computes derived values from stack input (labels, config)
  key_ring.go    — Creates the google_kms_key_ring resource
  outputs.go     — Output name constants matching stack_outputs.proto
```

## What It Creates

- `kms.KeyRing` — A Cloud KMS key ring in the specified project and location

## Outputs

| Output | Description |
|--------|-------------|
| `key_ring_id` | Fully qualified resource path (`projects/{project}/locations/{location}/keyRings/{name}`) |
| `key_ring_name` | Short name of the key ring |

## Local Development

```bash
# Build the module
make build

# Preview changes
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
make preview

# Apply changes
make up

# Destroy (removes from state only — key ring persists in GCP)
make destroy
```

## Debugging

```bash
./debug.sh ../hack/manifest.yaml
```

Then attach a Delve-compatible debugger to `:2345`.

## Notes

- **Key rings cannot be deleted from GCP.** The `destroy` command only removes the resource from Pulumi state.
- **No labels**: GCP KMS key rings do not support resource labels. Labels are computed in `locals.go` for internal tracking but not applied to the resource.
