# KubernetesCertificate Terraform Module

## Usage

```bash
openmcf tofu apply --manifest certificate.yaml
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
| `namespace` | Namespace where the Certificate was created |
| `certificate_name` | Name of the created Certificate resource |
| `secret_name` | TLS Secret name containing the signed certificate |
