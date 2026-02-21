# AliCloudPrivateDnsZone -- Terraform Module

This directory contains the Terraform (HCL) implementation for the AliCloudPrivateDnsZone deployment component.

## Structure

```
tf/
├── provider.tf    # AliCloud provider configuration
├── variables.tf   # Input variables (from proto spec)
├── locals.tf      # Tag computation, record map
├── main.tf        # Zone + VPC attachment
├── records.tf     # Zone records with for_each
└── outputs.tf     # Output values matching stack_outputs.proto
```

## Resources Created

1. **alicloud_pvtz_zone** -- the private DNS hosted zone
2. **alicloud_pvtz_zone_attachment** -- binds the zone to VPCs using dynamic blocks
3. **alicloud_pvtz_zone_record** -- one per record, managed via `for_each`

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=test.tfvars
```
