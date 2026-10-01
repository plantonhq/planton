# DigitalOcean Certificate -- Pulumi Module

Deploys a `digitalocean:index/certificate:Certificate` from a `DigitalOceanCertificate` stack input: the certificate's stable name plus exactly one source branch -- Let's Encrypt domains or custom PEM material. DigitalOcean's `type` argument is derived from whichever branch is set. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`, which carries the complete provider argument surface -- no PARITY-EXCEPTION guards. (This resource has no urn attribute at the provider, so there is none to export.)

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, certificate
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/certificate.go` -- the certificate resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Behavior notes

- The oneof branch derives the provider's `type` -- a mismatched type/branch combination is unrepresentable.
- Every argument is create-only; the default create-before-delete replacement order keeps name-referencing consumers (load balancers) gap-free.
- The resource id (and the `certificate_id` output) is the certificate NAME at this pin -- a Let's Encrypt certificate's UUID rotates on every auto-renewal.

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `certificate_id`, `expiry_rfc3339`.
