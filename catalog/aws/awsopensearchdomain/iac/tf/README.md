## Terraform Module to Deploy AwsOpenSearchDomain

Provisions the OpenSearch domain (`domain.tf`) plus its folded satellites
(`satellites.tf`): SAML Dashboards sign-in when `samlOptions` is
configured, and one cross-account VPC endpoint grant per
`authorizedVpcEndpointAccessAccounts` entry. Encryption at rest and
node-to-node TLS default to ON; the encryption, software-update, and
off-peak blocks are always sent explicitly so a `false` genuinely turns
the setting off.

Run the module via the Planton CLI (tofu) using the default local backend.

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

- Credentials are provided via IaC input (by the CLI), not in the manifest `spec`.
- Manifest file: `../../e2e/manifest.yaml`
