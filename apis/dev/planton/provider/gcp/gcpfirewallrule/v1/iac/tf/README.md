# GcpFirewallRule - Terraform Module

This Terraform module provisions a GCP compute firewall rule. It is the Terraform-side implementation of the Planton `GcpFirewallRule` resource kind and has feature parity with the Pulumi module.

## Usage

```bash
cd apis/dev/planton/provider/gcp/gcpfirewallrule/v1/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable schema. The primary input is `spec`, which mirrors the `GcpFirewallRuleSpec` protobuf message.

## Outputs

| Name | Description |
|------|-------------|
| `firewall_self_link` | Self-link URI of the created firewall rule |
| `firewall_name` | Name of the firewall rule |
| `creation_timestamp` | RFC3339 creation timestamp |
