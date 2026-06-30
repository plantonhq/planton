# OpenFGA Authorization Model - Pulumi Module

## Important: Pass-Through Placeholder

**OpenFGA does not have a Pulumi provider.** This module is a pass-through placeholder that exists only for structural consistency with other Planton deployment components.

**This module will NOT create any resources.**

## Use Terraform/Tofu Instead

To deploy OpenFGA Authorization Models, you must use Terraform/Tofu as the provisioner:

```bash
planton apply --manifest authorization-model.yaml \
  --openfga-provider-config openfga-creds.yaml \
  --provisioner tofu
```

## What This Module Does

When executed, this module:

1. Logs a warning that OpenFGA has no Pulumi provider
2. Exports a notice explaining the situation
3. Exits successfully without creating any resources

## Why Keep This Module?

This placeholder module exists for:

1. **Structural Consistency**: All Planton deployment components have both Pulumi and Terraform modules
2. **Future Compatibility**: If a Pulumi provider is ever created, we have the structure ready
3. **Clear Error Messaging**: Running with Pulumi gives clear guidance to use Terraform instead

## References

- [terraform-provider-openfga](https://github.com/openfga/terraform-provider-openfga) - The only available IaC provider for OpenFGA
- [OpenFGA Documentation](https://openfga.dev/docs)
