# Keyless GCP Deploys on Pulumi Again

**Date**: October 5, 2026
**Type**: Fix
**Components**: Pulumi GCP provider builder, GcpGlobalForwardingRule catalog page

## Summary

**A keyless (workload identity federation) GCP connection deploys on Pulumi again.** Since the GCP catalog moved to pulumi-gcp v9.37 on October 2, every keyless Pulumi infra job failed before it reached Google:

```
cannot encode provider configuration to call ValidateProviderConfig:
objectEncoder failed on property "external_credentials": Expected an Object PropertyValue, found []
```

The provider builder sent `external_credentials` as a single-element list. At v9.29 that list was the only shape the provider-config encoder accepted: the typed object failed there (pulumi-gcp#3869). v9.37 flipped the encoder, so the workaround became the failure. Static service-account keys, access tokens and the ambient credential chain were unaffected. Dev's and production's environment verdicts both failed `gcp-keyless-connect-and-deploy` on it.

**The keyless arm now passes the typed object** (`gcp.ProviderExternalCredentialsArgs`) through `gcp.NewProvider`, like every other credential mode. The identity token stays secret-wrapped, so it never reaches Pulumi state in plain text.

## What Changed

- `pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider/provider.go`: one typed provider path for every credential mode. The raw property map and its explicit plugin-version stamp are gone: `gcp.NewProvider` stamps the version itself, and the provider's resource name, and so its identity in existing stacks, is unchanged. The package doc records the history so the next SDK bump knows what to check.
- `provider_test.go`: pins the typed object shape and the secret-wrapped token (`TestBuildProviderInputs_WebIdentity_TypedObjectShape`). The two plugin-version tests went with the helper they tested.
- **The forwarding rule's display name is now "GCP Forwarding Rule".** The kind builds global and regional rules, and the console's create header read "Create GCP Global Forwarding Rule" above a Regional choice. Only the catalog page's H1 changed (the bundle reads its title from there), plus the link text in the pages that name it. The kind name `GcpGlobalForwardingRule` and the slug stay.

## How It Was Proven

A local Pulumi program built the provider through this package with a fake keyless token and previewed one bucket.
- **Before the change:** the exact production error, at `ValidateProviderConfig`.
- **After it:** the provider configured and the preview planned the bucket. A malformed token is now rejected by the provider's own JWT check, which reads `externalCredentials.identityToken`, so the object reaches the provider intact.

No GCP call was made. The live proof is the environments' own `gcp-keyless-connect-and-deploy` promise once a platform pins this release.
