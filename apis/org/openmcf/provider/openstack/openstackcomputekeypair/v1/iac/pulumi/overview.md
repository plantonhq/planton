# OpenStackComputeKeypair Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> OpenStackComputeKeypairStackInput (proto)
        ├── target: OpenStackComputeKeypair
        │     ├── metadata.name → keypair name
        │     └── spec.public_key, spec.region
        └── provider_config: OpenStackProviderConfig
              └── auth_url, credentials (oneof), region, ...
```

## Module Structure

1. **main.go (entrypoint)**: Loads `OpenStackComputeKeypairStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML), then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates an OpenStack Pulumi provider from `OpenStackProviderConfig`
   - Calls `keypair()` to create the compute keypair

3. **module/locals.go**: Extracts provider config and target resource from stack input into a `Locals` struct for convenient access.

4. **module/keypair.go**: Creates the `openstack.compute.Keypair` resource:
   - Sets `Name` from `metadata.name`
   - Sets `PublicKey` if provided (import mode)
   - Sets `Region` if overridden
   - Exports outputs: name, fingerprint, public_key, region
   - Exports private_key as a Pulumi secret (only populated for generated keys)

5. **module/outputs.go**: Constants for output names, matching `stack_outputs.proto` field names.

## Provider Configuration

The `pulumiopenstackprovider.Get()` helper maps `OpenStackProviderConfig` to Pulumi provider args:
- Supports all three auth methods (password, application credential, token)
- Falls back to `OS_*` environment variables when config is nil
- Handles TLS, endpoint type, and project/domain context
