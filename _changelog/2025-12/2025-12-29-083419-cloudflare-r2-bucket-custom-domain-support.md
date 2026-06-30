# Cloudflare R2 Bucket Custom Domain Support

**Date**: December 29, 2025
**Type**: Feature
**Components**: Cloudflare Provider, API Definitions, Pulumi CLI Integration, Terraform Module

## Summary

Added support for configuring custom domains on Cloudflare R2 buckets, enabling production-grade access via user-owned domains instead of rate-limited r2.dev URLs. The feature supports both literal zone ID values and foreign key references to `CloudflareDnsZone` resources, with conditional CEL validation ensuring fields are only required when the custom domain is enabled.

## Problem Statement / Motivation

Cloudflare R2 buckets can be made publicly accessible via `r2.dev` subdomains, but these URLs are rate-limited and not suitable for production workloads. The recommended approach for production is to configure a custom domain, which provides:

- No rate limits
- Full Cloudflare CDN caching
- Professional branding with your own domain
- HTTPS with your certificate

### Pain Points

- Users had to manually configure custom domains via Cloudflare Dashboard
- No IaC support for automating custom domain attachment to R2 buckets
- Hardcoding zone IDs was required when referencing zones
- No validation to ensure proper configuration

## Solution / What's New

### Custom Domain Configuration

Added a new `CloudflareR2BucketCustomDomainConfig` message to the R2 bucket spec with:

- **`enabled`** (bool): Toggle to enable/disable custom domain
- **`zone_id`** (StringValueOrRef): Zone ID with foreign key support
- **`domain`** (string): Full domain name (e.g., "media.example.com")

### Foreign Key Reference Support

The `zone_id` field uses the `StringValueOrRef` pattern, allowing users to:

1. Provide a literal zone ID value directly
2. Reference a `CloudflareDnsZone` resource's outputs

### Conditional Validation

CEL message-level validation ensures:
- `zone_id` is only required when `enabled` is true
- `domain` is only required when `enabled` is true

## Implementation Details

### Proto Schema (`spec.proto`)

```protobuf
message CloudflareR2BucketCustomDomainConfig {
  option (buf.validate.message).cel = {
    id: "custom_domain_zone_id_required"
    message: "zone_id is required when custom domain is enabled"
    expression: "!this.enabled || has(this.zone_id)"
  };

  option (buf.validate.message).cel = {
    id: "custom_domain_domain_required"
    message: "domain is required when custom domain is enabled"
    expression: "!this.enabled || this.domain != ''"
  };

  bool enabled = 1;
  
  dev.planton.shared.foreignkey.v1.StringValueOrRef zone_id = 2 [
    (dev.planton.shared.foreignkey.v1.default_kind) = CloudflareDnsZone,
    (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.zone_id"
  ];

  string domain = 3 [(buf.validate.field).string.max_len = 253];
}
```

### Pulumi Module (`bucket.go`)

```go
if locals.CloudflareR2Bucket.Spec.CustomDomain != nil && 
   locals.CloudflareR2Bucket.Spec.CustomDomain.Enabled {
    customDomain := locals.CloudflareR2Bucket.Spec.CustomDomain
    zoneId := customDomain.ZoneId.GetValue()

    _, err := cloudflare.NewR2CustomDomain(ctx, "custom-domain", 
        &cloudflare.R2CustomDomainArgs{
            AccountId:  pulumi.String(spec.AccountId),
            BucketName: createdBucket.Name,
            ZoneId:     pulumi.String(zoneId),
            Domain:     pulumi.String(customDomain.Domain),
        }, pulumi.Provider(cloudflareProvider), 
           pulumi.DependsOn([]pulumi.Resource{createdBucket}))

    ctx.Export(OpCustomDomainUrl, 
        pulumi.Sprintf("https://%s", customDomain.Domain))
}
```

### Terraform Module (`main.tf`)

```hcl
resource "cloudflare_r2_custom_domain" "main" {
  count       = local.custom_domain_enabled ? 1 : 0
  account_id  = local.account_id
  bucket_name = cloudflare_r2_bucket.main.name
  zone_id     = local.custom_domain_zone_id
  hostname    = local.custom_domain_name
}
```

### Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Added `CloudflareR2BucketCustomDomainConfig` message and field |
| `stack_outputs.proto` | Added `custom_domain_url` output |
| `bucket.go` | Added R2CustomDomain resource creation |
| `outputs.go` | Added `OpCustomDomainUrl` constant |
| `variables.tf` | Added custom_domain object with nested zone_id |
| `locals.tf` | Added custom domain local variables |
| `main.tf` | Added cloudflare_r2_custom_domain resource |
| `outputs.tf` | Added custom_domain_url output |
| `spec_test.go` | Added validation test cases for custom domain |
| `README.md` | Added custom domain documentation |
| `examples.md` | Added Examples 9 and 10 for custom domain |
| `overview.md` | Updated architecture diagram and implementation details |

## Benefits

### For Users

- **Production-ready R2 access**: Configure custom domains via IaC
- **No manual configuration**: Automate what previously required dashboard clicks
- **Reference-based configuration**: Use zone references instead of hardcoded IDs
- **Validation at schema level**: Get immediate feedback on misconfiguration

### For Developers

- **Consistent foreign key pattern**: Uses the same `StringValueOrRef` pattern as other components
- **CEL conditional validation**: Modern approach to conditional field requirements
- **Full IaC support**: Both Pulumi and Terraform implementations

## Impact

### User Experience

Users can now declare custom domains directly in their R2 bucket manifests:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareR2Bucket
metadata:
  name: media-bucket
spec:
  bucketName: media-assets
  accountId: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  location: WEUR
  customDomain:
    enabled: true
    zoneId:
      valueFrom:
        name: my-dns-zone  # References CloudflareDnsZone
    domain: "media.example.com"
```

### Provider Coverage

The Cloudflare R2 Bucket component now provides:
- ✅ Bucket creation with location hints
- ✅ Public access configuration
- ✅ Custom domain attachment
- ✅ Versioning flag (with warning - R2 doesn't support it)

## Related Work

- Uses the `StringValueOrRef` foreign key pattern from `shared/foreignkey/v1`
- Similar to `CloudflareLoadBalancer` zone_id reference pattern
- Extends the R2 bucket component following existing research documentation

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

