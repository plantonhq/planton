# KubernetesIssuer Terraform Module

## Usage

```hcl
module "issuer" {
  source = "./iac/tf"

  metadata = {
    name = "my-ca-issuer"
  }

  spec = {
    namespace = "my-namespace"
    ca = {
      ca_secret_name = "my-ca-keypair"
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, labels, etc.) | object | yes |
| `spec` | Issuer specification (namespace + exactly one of ca/self_signed) | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | Namespace where the Issuer was created |
| `issuer_name` | Name of the created Issuer |
