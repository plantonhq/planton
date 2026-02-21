# AlicloudContainerRegistry - Research Documentation

## Provider Resource Analysis

### Terraform Resources

**alicloud_cr_ee_instance** (Primary)
- Creates an ACR Enterprise Edition instance via BSS (Business Support Service) API
- Creation time: ~6 minutes (asynchronous provisioning)
- Instance enters `RUNNING` state when ready
- ForceNew fields: `instance_name`, `instance_type`, `payment_type`

**alicloud_cr_ee_namespace** (Bundled)
- Creates a namespace within an Enterprise Edition instance
- ForceNew fields: `instance_id`, `name`
- Updatable fields: `auto_create`, `default_visibility`

### Pulumi Resources

- Instance: `cr.RegistryEnterpriseInstance` from `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cr`
- Namespace: `cs.RegistryEnterpriseNamespace` from `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cs`

Note: The namespace resource is in the `cs` (Container Service) package, not `cr`. This is a Pulumi SDK organizational quirk where Enterprise Edition namespace/repo/sync-rule resources were originally placed under Container Service.

## Provider Fields Not Exposed

The following provider fields were intentionally excluded from the spec:

| Field | Reason |
|-------|--------|
| `namespace_quota` | Quota tuning; provider defaults are sufficient for 80%+ of users |
| `repo_quota` | Same as above |
| `vpc_quota` | Deprecated field; VPC endpoint management done separately |
| `image_scanner` | Defaults to "ACR"; SAS (Security Advisor Service) integration is niche |
| `custom_oss_bucket` | Custom storage bucket; rarely needed |
| `default_oss_bucket` | Whether to use default OSS bucket; implied by not setting custom |
| `kms_encrypted_password` | KMS encryption for password; users can use plain password field |
| `kms_encryption_context` | Related to KMS encryption |
| `renewal_status` | Auto-renewal configuration; manageable via console |
| `renew_period` | Auto-renewal period; manageable via console |

## Endpoint Architecture

The `instance_endpoints` computed attribute returns a list of network access endpoints:

```
instance_endpoints = [
  {
    endpoint_type = "internet"
    enable = true
    domains = [
      { type = "SYSTEM", domain = "myregistry-registry.cn-hangzhou.cr.aliyuncs.com" }
    ]
  },
  {
    endpoint_type = "vpc"
    enable = true
    domains = [
      { type = "SYSTEM", domain = "myregistry-registry-vpc.cn-hangzhou.cr.aliyuncs.com" }
    ]
  }
]
```

The component extracts the first domain from each endpoint type for the `public_endpoint` and `vpc_endpoint` outputs.

## Tags Limitation

ACR Enterprise Edition instances do not support tags in either the Terraform or Pulumi provider. This is because the instance is provisioned through the BSS (Business Support Service) API, which has a different lifecycle model than regular resource APIs. The `locals.go` in the Pulumi module does not compute tags, unlike other Alibaba Cloud components.

## Personal Edition vs Enterprise Edition

Alibaba Cloud offers two container registry products:

| Feature | Personal Edition | Enterprise Edition |
|---------|-----------------|-------------------|
| Terraform Resources | `alicloud_cr_namespace`, `alicloud_cr_repo` | `alicloud_cr_ee_instance`, `alicloud_cr_ee_namespace`, `alicloud_cr_ee_repo` |
| Cost | Free | Paid (Basic/Standard/Advanced tiers) |
| Instance Required | No | Yes |
| Image Scanning | No | Yes |
| Geo-Replication | No | Advanced tier |
| Helm Charts | No | Yes |

This component targets Enterprise Edition only, following the same pattern as Azure Container Registry (which covers Basic/Standard/Premium SKUs in one component without creating separate components for each tier).

## Bundling Rationale (DD07)

Namespaces are bundled into this component because:
1. A registry instance without namespaces cannot store images
2. Namespaces have a mandatory dependency on the instance (require `instance_id`)
3. The combination represents a single logical unit: "a container registry ready to use"

Repositories (`alicloud_cr_ee_repo`) are NOT bundled because they are typically created dynamically by CI/CD pipelines, not statically in infrastructure definitions.
