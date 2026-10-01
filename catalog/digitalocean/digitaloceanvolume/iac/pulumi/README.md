# DigitalOcean Volume -- Pulumi Module

Deploys a `digitalocean:index/volume:Volume` from a `DigitalOceanVolume` stack input: the full provider argument surface -- name, region, size, description, one-time filesystem formatting with an optional label, snapshot source, and tags. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete argument surface -- no PARITY-EXCEPTION guards. (The SDK renames the volume's `urn` attribute to `VolumeUrn`; the module exports it under the contract's `urn` key. The SDK's `FilesystemType` input maps to the provider's DEPRECATED attribute -- this module wires `InitialFilesystemType`, never that one.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, volume
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/volume.go` -- the volume resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- The spec enum's value names ARE the provider's filesystem strings (`ext4`/`xfs`); `unformatted` stays unset. The label is sent only when a filesystem type is being formatted.
- Size can only be EXPANDED (provider plan-time rule); description is create-only at this pin -- a change plans a REPLACEMENT.
- Tags are the union of `spec.tags` and the standard Planton labels rendered as `key:value` -- the exact set the Terraform module applies.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `volume_id`, `urn`.
