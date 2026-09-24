# DigitalOcean Container Registry -- Pulumi Module

Deploys a `digitalocean:index/containerRegistry:ContainerRegistry` from a `DigitalOceanContainerRegistry` stack input, plus a `digitalocean:index/containerRegistryDockerCredentials:ContainerRegistryDockerCredentials` when (and only when) the spec's `docker_credentials` block is set. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface for both resources -- no PARITY-EXCEPTION guards. (The SDK does NOT flag the credential token as secret, so the module wraps the export in `pulumi.ToSecret`.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, registry
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/registry.go` -- the registry resource, the conditional credentials resource, and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- The proto enum value names ARE the DigitalOcean tier slugs; an unspecified region stays nil so DigitalOcean chooses and reports the slug back.
- `server_url` and `endpoint` are read from the SDK's real resource attributes, never string-formatted.
- The credentials resource exists only when the spec block does; its knobs are unrecoverable from the API and its importer is DEFECTIVE at this pin -- see `../import-map.yaml`. Unconfigured credentials export as empty strings, identical to the Terraform module.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `registry_name`, `server_url`, `endpoint`, `region`, `docker_credentials` (secret), `credential_expiration_time`.
