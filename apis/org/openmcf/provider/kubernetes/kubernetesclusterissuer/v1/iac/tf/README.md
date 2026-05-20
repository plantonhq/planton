# KubernetesClusterIssuer Terraform Module

## Usage

```bash
openmcf tofu apply --manifest cluster-issuer.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification.

## Outputs

| Output | Description |
|--------|-------------|
| `cluster_issuer_name` | Name of the created ClusterIssuer |
| `acme_account_key_secret_name` | ACME account key Secret name |
