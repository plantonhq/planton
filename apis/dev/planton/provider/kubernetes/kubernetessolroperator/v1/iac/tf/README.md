# Terraform Module for KubernetesSolrOperator

## Status: Not Implemented

⚠️ **The Terraform implementation for KubernetesSolrOperator is currently not available.**

This component currently supports deployment via **Pulumi only**. The Terraform module skeleton files exist but the implementation is incomplete.

## Recommended Approach

For deploying the Apache Solr Operator, please use the **Pulumi module**:

```bash
cd ../pulumi
pulumi up
```

See [../pulumi/README.md](../pulumi/README.md) for complete Pulumi deployment instructions.

## Why Pulumi Only?

The Pulumi implementation provides:
- Dynamic resource creation and updates
- Strong type safety with Go
- Better handling of Helm chart dependencies
- CRD installation before Helm chart deployment
- Comprehensive output management

## Future Plans

A Terraform implementation may be added in a future release. The implementation would need to:

1. Install Solr Operator CRDs
2. Deploy operator Helm chart
3. Handle proper dependency ordering
4. Export relevant outputs (namespace, CRD status)

## Alternative: Direct Terraform + Helm

If you must use Terraform, you can deploy the operator directly using the Terraform Helm provider:

```hcl
provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}

# Install CRDs first
resource "null_resource" "solr_operator_crds" {
  provisioner "local-exec" {
    command = <<-EOT
      kubectl apply -f https://solr.apache.org/operator/downloads/crds/v0.7.0/all-with-dependencies.yaml
    EOT
  }
}

# Deploy operator
resource "helm_release" "solr_operator" {
  name       = "solr-operator"
  repository = "https://solr.apache.org/charts"
  chart      = "solr-operator"
  version    = "0.7.0"
  namespace  = "solr-operator"
  
  create_namespace = true
  
  depends_on = [null_resource.solr_operator_crds]
}
```

**Note**: This approach lacks the Planton integration for credential management and standardized outputs.

## Support

For deployment questions:
- **Pulumi users**: See [../pulumi/README.md](../pulumi/README.md)
- **General questions**: Refer to [../../README.md](../../README.md)
- **Upstream documentation**: https://apache.github.io/solr-operator/

## Contributing

If you're interested in implementing the Terraform module, contributions are welcome! The implementation should match the functionality provided by the Pulumi module in `../pulumi/module/`.

