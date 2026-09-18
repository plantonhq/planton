# DigitalOcean Bucket -- Pulumi Module

Deploys a `digitalocean:index/spacesBucket:SpacesBucket` plus the per-bucket settings satellites (`SpacesBucketCorsConfiguration`, `SpacesBucketPolicy`, `SpacesBucketLogging`) from a `DigitalOceanBucket` stack input: region and canned ACL, versioning, lifecycle rules, CORS, a JSON policy, access logging, and force-destroy. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`.

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, bucket
- `module/locals.go` -- stack-input references
- `module/bucket.go` -- the bucket, its satellites, and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `bucket_id`, `endpoint`, `region`, `bucket_domain_name`, `urn`.

The provider's `urn` attribute is `BucketUrn` in the SDK (`URN()` is Pulumi's own resource URN, a different thing).

## Behavior notes

- `region` is sent only when set (the zero enum value never becomes a slug); unset lets the provider apply its own default (`nyc3`).
- CORS, policy, and logging are created as child resources only when configured. Their `Bucket` is the created bucket's id; their `Region` is the spec's region (the spec requires it whenever a satellite is set).
- CORS uses the standalone `SpacesBucketCorsConfiguration` resource. The bucket's deprecated inline `CorsRules` argument is never written.
- There is no Pulumi SDK gap on this surface at the pin (re-verified on disk 2026-09-18 at v4.79.1: `SpacesBucketArgs`, `SpacesBucketCorsConfigurationArgs`, `SpacesBucketPolicyArgs`, and `SpacesBucketLoggingArgs` are all present) — every modeled argument is wired.
- See the kind [GUIDE](../../GUIDE.md).
