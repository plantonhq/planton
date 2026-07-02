---
title: "Troubleshooting"
description: "Solutions to common Planton issues - manifest validation, authentication, state management, and deployment problems"
icon: "gear"
order: 100
---

# Troubleshooting Guide

Solutions to common problems you might encounter using Planton.

---

## Manifest Validation Errors

### "kind not supported" or "Unsupported cloud resource kind"

**Symptom**: CLI doesn't recognize your resource kind.

**Common Causes**:
- Typo in `kind` field (case-sensitive)
- Kind doesn't exist in Planton
- Wrong `apiVersion` for the kind

**Solutions**:

```bash
# 1. Check spelling (kinds are PascalCase)
# Wrong: awsS3Bucket, aws_s3_bucket
# Right: AwsS3Bucket

# 2. Verify kind exists in catalog
# Browse: /docs/catalog

# 3. Check available kinds in your CLI version
# (List is shown in error message)
```

### "validation error: spec.field ..."

**Symptom**: Manifest validation fails with field-specific errors.

**Common Examples**:

```
spec.replicas: value must be >= 1 and <= 10 (got: 0)
spec.container.cpu: value must match pattern "^[0-9]+m$" (got: "2cores")
spec.region: value is required
```

**Solutions**:

```bash
# 1. Read the error message carefully - it tells you exactly what's wrong

# 2. Check field types in proto definition or docs
# Integers: replicas: 3 (not "3")
# Strings: region: "us-west-2"
# Booleans: enabled: true (not "true")

# 3. Verify value formats
# CPU: "500m" or "2" (not "2cores" or "500mcpu")
# Memory: "1Gi" or "512Mi" (not "1GB" or "512MB")

# 4. Check required fields
# If error says "value is required", you must provide it
```

### YAML Syntax Errors

**Symptom**: "failed to load yaml to json" or similar parsing errors.

**Solutions**:

```bash
# Validate YAML syntax
cat manifest.yaml | yq .

# Or use Python
python3 -c "import yaml; yaml.safe_load(open('manifest.yaml'))"

# Common issues:
# - Inconsistent indentation (use spaces, not tabs)
# - Missing colons after keys
# - Unquoted strings with special characters
```

---

## Authentication & Credentials

### AWS: "Unable to locate credentials"

**Symptom**: AWS provider can't find credentials.

**Solutions**:

```bash
# Check environment variables
env | grep AWS

# Should show:
# AWS_ACCESS_KEY_ID=AKIA...
# AWS_SECRET_ACCESS_KEY=...

# If not set:
export AWS_ACCESS_KEY_ID="your-key-id"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_DEFAULT_REGION="us-west-2"

# Or use AWS profile
export AWS_PROFILE=production

# Verify credentials work
aws sts get-caller-identity
```

### AWS: "Access Denied" or Permission Errors

**Symptom**: Credentials work but lack permissions.

**Solutions**:

```bash
# Check what identity you're using
aws sts get-caller-identity

# Check attached policies
aws iam list-attached-user-policies --user-name your-username

# Common issues:
# - Need specific IAM permissions for resource type
# - Resource locked by SCPs or permission boundaries
# - Wrong AWS account/region

# Contact your AWS administrator to grant necessary permissions
```

### GCP: "Application Default Credentials not found"

**Symptom**: GCP provider can't find credentials.

**Solutions**:

```bash
# Option 1: Use service account key
export GOOGLE_APPLICATION_CREDENTIALS=~/gcp-key.json
planton pulumi up -f resource.yaml

# Option 2: Use application default credentials (local dev)
gcloud auth application-default login
planton pulumi up -f resource.yaml

# Verify credentials
gcloud auth list
gcloud config get-value project
```

### GCP: Permission Denied Errors

**Symptom**: Credentials work but can't create resources.

**Solutions**:

```bash
# Check current project
gcloud config get-value project

# Check service account permissions
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:YOUR_SA@*"

# Common issues:
# - Service account lacks necessary roles
# - APIs not enabled (e.g., GKE API for clusters)
# - Organization policies blocking

# Enable required APIs
gcloud services enable container.googleapis.com  # For GKE
gcloud services enable sqladmin.googleapis.com   # For Cloud SQL
```

### Azure: "Failed to authenticate"

**Symptom**: Azure provider can't authenticate.

**Solutions**:

```bash
# Check environment variables
env | grep ARM_

# Should show:
# ARM_CLIENT_ID=abc-123
# ARM_CLIENT_SECRET=xyz-789
# ARM_TENANT_ID=def-456
# ARM_SUBSCRIPTION_ID=...

# Or use Azure CLI login
az login
az account show

# Verify correct subscription
az account list
az account set --subscription "My Subscription"
```

### Cloudflare: "Authentication failed"

**Symptom**: Can't authenticate with Cloudflare.

**Solutions**:

```bash
# Check API token
echo $CLOUDFLARE_API_TOKEN

# Test token
curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN"

# Common issues:
# - Token expired
# - Token has insufficient permissions
# - Wrong token type (API token vs API key)

# Create new token with correct permissions in Cloudflare dashboard
```

---

## Pulumi-Specific Issues

### "no stack named '...' found"

**Symptom**: Pulumi can't find the stack.

**Solutions**:

```bash
# Option 1: Initialize the stack first
planton pulumi init -f resource.yaml
planton pulumi preview -f resource.yaml

# Option 2: Use 'up' which auto-creates stack
planton pulumi up -f resource.yaml

# Option 3: Check stack label in manifest
# Ensure manifest has:
metadata:
  labels:
    pulumi.planton.dev/stack.name: "org/project/stack"
```

### "another update is currently in progress"

**Symptom**: Stack is locked by another operation.

**Solutions**:

```bash
# Check if operation is actually running
pulumi stack --stack <stack-fqdn>

# If no operation running, cancel the lock
pulumi cancel --stack <stack-fqdn>

# Then retry
planton pulumi up -f resource.yaml
```

### "Stack still has resources" (during delete)

**Symptom**: Can't delete stack because resources still exist.

**Solutions**:

```bash
# Destroy resources first
planton pulumi destroy -f resource.yaml

# Then delete stack
planton pulumi delete -f resource.yaml

# Or if resources are actually gone (state is wrong):
planton pulumi refresh -f resource.yaml
planton pulumi delete -f resource.yaml

# Or force delete (use with caution)
planton pulumi delete -f resource.yaml --force
```

---

## OpenTofu-Specific Issues

### "Backend initialization required"

**Symptom**: OpenTofu commands fail with initialization error.

**Solutions**:

```bash
# Run init first
planton tofu init -f resource.yaml

# Then try command again
planton tofu plan -f resource.yaml
```

### "state locked"

**Symptom**: State file is locked by another operation.

**Solutions**:

```bash
# Wait for other operation to complete

# Or force unlock (if certain no operation is running)
cd <module-directory>
tofu force-unlock <lock-id>

# The lock-id is shown in the error message
```

### "Error acquiring the state lock"

**Symptom**: Can't acquire lock on state file.

**Solutions**:

```bash
# Check backend configuration
# - Ensure backend supports locking (S3 with DynamoDB, GCS, etc.)
# - Local backend doesn't support locking

# For S3 backend, verify DynamoDB table exists
aws dynamodb describe-table --table-name terraform-locks

# If using local backend, only one operation at a time
```

---

## Kustomize Issues

### "kustomization path does not exist"

**Symptom**: Can't find kustomize directory or overlay.

**Solutions**:

```bash
# Verify directory structure
ls -R services/api/kustomize/

# Should contain:
# overlays/prod/kustomization.yaml

# Check overlay name spelling
# Wrong: --overlay production
# Right: --overlay prod (must match directory name)

# Verify full path
ls services/api/kustomize/overlays/prod/kustomization.yaml
```

### "accumulating resources: accumulation err='accumulating resources from '../../base': ..."

**Symptom**: Kustomize can't build overlay.

**Solutions**:

```bash
# 1. Verify kustomization.yaml syntax
cat overlays/prod/kustomization.yaml

# 2. Check resource paths exist
# If kustomization.yaml references "../../base"
ls -la overlays/prod/../../base

# 3. Test kustomize directly
cd services/api/kustomize
kustomize build overlays/prod

# Error messages from kustomize are more detailed
```

---

## Deployment Failures

### "resource already exists"

**Symptom**: Can't create resource because it already exists.

**Causes**:
- Resource created outside of Pulumi/OpenTofu
- State file lost/corrupted
- Deploying to wrong environment

**Solutions**:

```bash
# Option 1: Import existing resource (Pulumi)
pulumi import <type> <name> <id> --stack <stack-fqdn>

# Option 2: Import existing resource (OpenTofu)
cd <module-dir>
tofu import <resource-type>.<resource-name> <id>

# Option 3: Delete existing resource manually
# (via cloud console or CLI)

# Option 4: Use different resource name
vim manifest.yaml  # Change metadata.name
```

### Partial Deployment Failures

**Symptom**: Some resources created, others failed.

**Solutions**:

```bash
# For Pulumi: State automatically updated, can re-run 'up'
planton pulumi up -f resource.yaml
# Will continue from where it failed

# For OpenTofu: State updated, can re-run 'apply'
planton tofu apply -f resource.yaml
# Will attempt to create failed resources

# If repeatedly failing:
# 1. Check error messages
# 2. Fix underlying issue (permissions, quotas, etc.)
# 3. Retry
```

### Timeout Errors

**Symptom**: Deployment times out waiting for resource.

**Causes**:
- Resource taking longer than expected (e.g., database initialization)
- Resource stuck in provisioning state
- Provider API issues

**Solutions**:

```bash
# Check resource status in cloud console
# - Is it actually creating?
# - Any errors shown?

# Wait and retry
# Some resources (databases, clusters) can take 10-20 minutes

# For Kubernetes resources: check pod status
kubectl get pods -n <namespace>
kubectl describe pod <pod-name> -n <namespace>
```

---

## Module Issues

### "module not found"

**Symptom**: Can't find IaC module directory.

**Solutions**:

```bash
# Check module directory exists
ls -la <module-dir>

# Should contain:
# - Pulumi: main.go, go.mod, Pulumi.yaml
# - OpenTofu: *.tf files

# Verify --module-dir path
planton pulumi up \
  -f resource.yaml \
  --module-dir /full/path/to/module  # Use absolute path

# Or cd to module directory
cd /path/to/module
planton pulumi up -f ~/manifests/resource.yaml
```

### Module Compilation Errors

**Symptom**: Pulumi module fails to compile.

**Causes**:
- Missing Go dependencies
- Syntax errors in module code
- Incompatible Go version

**Solutions**:

```bash
# Verify Go installation
go version

# Update module dependencies
cd <module-dir>
go mod tidy
go mod download

# Test compilation
go build .

# Check for syntax errors
go vet ./...
```

---

## State Issues

### Drift Detection

**Symptom**: Preview/plan shows unexpected changes.

**Causes**:
- Manual changes made outside Pulumi/OpenTofu
- Provider defaults changed
- Upstream dependencies changed

**Solutions**:

```bash
# For Pulumi: refresh state
planton pulumi refresh -f resource.yaml

# For OpenTofu: refresh state
planton tofu refresh -f resource.yaml

# Then plan again to see if changes persist
planton pulumi preview -f resource.yaml
planton tofu plan -f resource.yaml

# If unexpected changes remain:
# - Investigate who made manual changes
# - Update manifest to match reality
# - Or revert changes via apply/up
```

### State File Corruption

**Symptom**: State file is corrupted or lost.

**Solutions**:

```bash
# For Pulumi: export and restore state
pulumi stack export --stack <stack-fqdn> > backup.json
# If corrupted, restore from backup
pulumi stack import --stack <stack-fqdn> < backup.json

# For OpenTofu: restore from backend backup
# (Depends on backend - S3 versioning, GCS versioning, etc.)

# If no backup available:
# - Manually recreate state by importing resources
# - Or destroy and recreate resources
```

---

## Network & Connectivity

### "connection refused" or Timeout Errors

**Symptom**: Can't connect to cloud provider APIs.

**Solutions**:

```bash
# Check internet connection
curl -I https://api.github.com

# Check if cloud provider APIs are accessible
curl -I https://api.aws.amazon.com    # AWS
curl -I https://www.googleapis.com    # GCP
curl -I https://api.cloudflare.com    # Cloudflare

# If behind corporate proxy:
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Check firewall rules
# - Corporate firewall may block cloud provider APIs
# - VPN may be required
```

---

## Installation & Prerequisites

### "pulumi: command not found"

**Symptom**: Pulumi CLI not installed.

**Solutions**:

```bash
# Install Pulumi
brew install pulumi

# Or download from pulumi.com
curl -fsSL https://get.pulumi.com | sh

# Verify installation
pulumi version
```

### "tofu: command not found"

**Symptom**: OpenTofu CLI not installed.

**Solutions**:

```bash
# Install OpenTofu
brew install opentofu

# Or download from opentofu.org

# Verify installation
tofu version
```

### "planton: command not found"

**Symptom**: Planton CLI not installed.

**Solutions**:

```bash
# Install via Homebrew
brew install plantonhq/tap/planton

# Verify installation
planton version

# If Homebrew tap not added:
brew tap plantonhq/tap
brew install planton
```

---

## Provider-Specific Issues

### Kubernetes: "Unable to connect to cluster"

**Symptom**: Can't connect to Kubernetes cluster.

**Solutions**:

```bash
# Verify kubeconfig
kubectl cluster-info

# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch to correct context
kubectl config use-context my-cluster

# Verify cluster access
kubectl get nodes
```

### AWS: "CredentialRequireScopeError"

**Symptom**: AWS credentials missing scope/region.

**Solutions**:

```bash
# Set region explicitly
export AWS_DEFAULT_REGION=us-west-2

# Or in credential file
cat > aws-cred.yaml <<EOF
accessKeyId: AKIA...
secretAccessKey: ...
region: us-west-2
EOF

planton pulumi up -f resource.yaml -p aws-cred.yaml
```

### GCP: "API not enabled"

**Symptom**: Can't create resource because API isn't enabled.

**Solutions**:

```bash
# Enable required API
gcloud services enable compute.googleapis.com       # Compute Engine
gcloud services enable container.googleapis.com     # GKE
gcloud services enable sqladmin.googleapis.com      # Cloud SQL

# Check enabled APIs
gcloud services list --enabled
```

---

## Common Error Messages

### "no such file or directory"

**Likely causes**:
- Manifest file path is wrong
- Kustomize directory doesn't exist
- Module directory not found

**Solution**: Verify all paths exist and are spelled correctly.

### "permission denied"

**Likely causes**:
- File permissions issue (can't read manifest)
- Cloud provider permission issue (can't create resource)

**Solution**: Check file permissions or cloud credentials.

### "context deadline exceeded" or Timeout

**Likely causes**:
- Resource taking longer than expected
- Network issues
- Provider API slowness

**Solution**: Wait and retry. Some resources (databases, clusters) can take 15+ minutes.

---

## Getting Additional Help

### Enable Debug Logging

```bash
# Pulumi verbose logging
export PULUMI_LOG_LEVEL=3
planton pulumi up -f resource.yaml

# OpenTofu debug logging
export TF_LOG=DEBUG
planton tofu apply -f resource.yaml

# Or trace level for maximum verbosity
export TF_LOG=TRACE
```

### Check Provider Documentation

- [AWS Provider Docs](https://www.pulumi.com/registry/packages/aws/)
- [GCP Provider Docs](https://www.pulumi.com/registry/packages/gcp/)
- [Azure Provider Docs](https://www.pulumi.com/registry/packages/azure/)
- [Kubernetes Provider Docs](https://www.pulumi.com/registry/packages/kubernetes/)

### Community Support

**GitHub Issues**: [planton/issues](https://github.com/plantonhq/planton/issues)

**GitHub Discussions**: [planton/discussions](https://github.com/plantonhq/planton/discussions)

**When reporting issues**:
- Include Planton version (`planton version`)
- Include relevant error messages (sanitize credentials!)
- Include manifest structure (sanitize sensitive data)
- Describe what you expected vs. what happened
- Include steps to reproduce

---

## Preventive Measures

### Before Deploying to Production

- [ ] Validate manifest: `planton validate -f prod.yaml`
- [ ] Preview changes: `planton pulumi preview -f prod.yaml`
- [ ] Test in lower environment first
- [ ] Verify credentials are correct (not dev/staging credentials)
- [ ] Check region/zone is correct
- [ ] Review resource names (ensure they're production-appropriate)
- [ ] Backup existing data (for databases, storage)
- [ ] Have rollback plan ready

### Regular Maintenance

- [ ] Rotate credentials every 90 days
- [ ] Update provider versions regularly
- [ ] Clean up unused stacks/state files
- [ ] Review and update manifests for best practices
- [ ] Monitor cloud provider quotas/limits
- [ ] Keep Planton CLI updated

---

## Related Documentation

- [Manifest Structure](/docs/guides/manifests) - Understanding manifests
- [Credentials Guide](/docs/guides/credentials) - Setting up credentials
- [Pulumi Commands](/docs/cli/pulumi-commands) - Pulumi troubleshooting
- [OpenTofu Commands](/docs/cli/tofu-commands) - OpenTofu troubleshooting

---

**Still stuck?** [Open an issue](https://github.com/plantonhq/planton/issues) with details, and we'll help you out!

