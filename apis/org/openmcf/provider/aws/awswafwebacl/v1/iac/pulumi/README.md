# AwsWafWebAcl Pulumi Module

Pulumi IaC module for deploying an AWS WAFv2 Web ACL.

## Usage

```bash
# Set stack input
export STACK_INPUT='{"target":{"apiVersion":"aws.openmcf.org/v1","kind":"AwsWafWebAcl",...}}'

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

# Test with manifest
export STACK_INPUT=$(cat iac/hack/manifest.yaml | yq -o json)
pulumi preview --stack dev
```

## Debug

Use `debug.sh` to run with verbose logging:

```bash
./debug.sh
```
