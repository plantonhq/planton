# OpenFGA Store - Pulumi Module

> ⚠️ **IMPORTANT**: OpenFGA does not have a Pulumi provider. This module is a **pass-through placeholder** that does not create any resources.

## Use Terraform Instead

To deploy OpenFGA resources, you **must** use Terraform/Tofu as the provisioner:

```bash
planton apply --manifest openfga-store.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## Why This Module Exists

This module exists only to maintain consistency with the Planton deployment component structure. All deployment components have both Pulumi and Terraform modules, but OpenFGA only has a Terraform provider.

When this module is executed:
1. It logs a warning that no Pulumi provider is available
2. It exports a notice indicating no resources were created
3. It exits successfully without creating any infrastructure

## References

- [OpenFGA Terraform Provider](https://github.com/openfga/terraform-provider-openfga)
- [OpenFGA Documentation](https://openfga.dev/docs)
- [Terraform Module README](../tf/README.md)
