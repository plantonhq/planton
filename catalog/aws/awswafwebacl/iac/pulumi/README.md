# AwsWafWebAcl Pulumi Module

Pulumi IaC module for deploying an AWS WAFv2 Web ACL.

## Usage

```bash
# Set IaC input (an AwsWafWebAclIacInput: the manifest under "target", plus an optional "provider_config")
export IAC_INPUT_YAML='{"target":{"apiVersion":"aws.planton.dev/v1alpha1","kind":"AwsWafWebAcl",...}}'

# Preview
pulumi preview --stack dev

# Deploy
pulumi up --stack dev --yes

# Destroy
pulumi destroy --stack dev --yes
```

## Development

```bash
# Build
go build ./...

# Test with manifest (the CLI wraps it into the IaC input and runs this directory's module)
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/dev
```

The module reads its input from the Pulumi config key `planton:iac-input`, `IAC_INPUT_YAML` (YAML or JSON content), or `IAC_INPUT_YAML_FILE` (a path to it).

## Debug

Run Pulumi with verbose logging:

```bash
pulumi preview --stack dev --logtostderr -v=9
```
