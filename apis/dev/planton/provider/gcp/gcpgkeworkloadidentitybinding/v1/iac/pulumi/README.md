# GCP GKE Workload Identity Binding - Pulumi Module

This directory contains the Pulumi implementation of the GCP GKE Workload Identity Binding component.

## Overview

The Pulumi module creates an IAM policy binding that grants `roles/iam.workloadIdentityUser` to a Kubernetes Service Account (KSA), allowing pods using that KSA to impersonate a Google Service Account (GSA) when accessing Google Cloud APIs.

## Architecture

The module performs the following operations:

1. **Constructs the Workload Identity principal**: Builds the member string in the format `serviceAccount:{project-id}.svc.id.goog[{namespace}/{ksa-name}]`
2. **Creates IAM binding**: Uses `gcp.serviceaccount.IAMMember` to grant the `roles/iam.workloadIdentityUser` role
3. **Exports outputs**: Returns the member string and GSA email for reference

## Module Structure

```
iac/pulumi/
├── main.go                           # Pulumi program entrypoint
├── Pulumi.yaml                       # Pulumi project configuration
├── Makefile                          # Build and deploy commands
├── debug.sh                          # Local debugging script
├── README.md                         # This file
├── overview.md                       # Architecture and design details
└── module/
    ├── main.go                       # Resources function (entry point)
    ├── locals.go                     # Local variables initialization
    ├── outputs.go                    # Output constant definitions
    └── workload_identity_binding.go  # IAM binding implementation
```

## Prerequisites

### Required Tools

- **Pulumi CLI**: v3.0+
- **Go**: 1.21+
- **gcloud CLI**: For GCP authentication

### GCP Prerequisites

- GKE cluster with Workload Identity enabled
- Node pools configured with `workloadMetadataConfig.mode = GKE_METADATA`
- A Google Service Account (GSA) with appropriate IAM roles for your workload
- A Kubernetes Service Account (KSA) in the target namespace

### Authentication

The module expects GCP credentials to be provided via the stack input's `provider_config`. Planton's CLI handles this automatically, but for standalone use:

```bash
# Authenticate with gcloud
gcloud auth application-default login

# Or set credentials explicitly
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

## Usage

### Via Planton CLI (Recommended)

```bash
# Create a manifest file
cat > manifest.yaml <<EOF
apiVersion: gcp.planton.dev/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-dns-binding
spec:
  projectId: "prod-project"
  serviceAccountEmail: "dns01-solver@prod-project.iam.gserviceaccount.com"
  ksaNamespace: "cert-manager"
  ksaName: "cert-manager"
EOF

# Deploy using Planton CLI
planton apply -f manifest.yaml
```

### Standalone Pulumi Deployment

For development or testing:

```bash
# Navigate to the Pulumi module
cd iac/pulumi

# Set required configuration
pulumi config set gcp:project prod-project

# Preview changes
pulumi preview

# Deploy
pulumi up
```

## Environment Variables

When running standalone, the following environment variables may be needed:

| Variable | Description | Required |
|----------|-------------|----------|
| `PULUMI_STACK_INPUT` | JSON-encoded stack input | Yes |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP service account key | No (if using gcloud auth) |
| `PULUMI_ACCESS_TOKEN` | Pulumi Cloud access token | No (if using local backend) |

## Local Development

### Debug Mode

Use the provided debug script for local testing:

```bash
# Make debug script executable
chmod +x debug.sh

# Run with environment variables
export PULUMI_STACK_INPUT='{"spec":{"projectId":{"value":"prod-project"},"serviceAccountEmail":{"value":"dns01-solver@prod-project.iam.gserviceaccount.com"},"ksaNamespace":"cert-manager","ksaName":"cert-manager"},"providerConfig":{"credential":{"valueFrom":{"env":"GOOGLE_APPLICATION_CREDENTIALS"}}}}'

./debug.sh
```

### Makefile Commands

```bash
# Install dependencies
make install

# Run unit tests
make test

# Build the module
make build

# Preview changes
make preview

# Deploy
make deploy

# Destroy resources
make destroy
```

## Stack Inputs

The module expects a `GcpGkeWorkloadIdentityBindingStackInput` protobuf message with the following structure:

```json
{
  "spec": {
    "projectId": {
      "value": "prod-project"
    },
    "serviceAccountEmail": {
      "value": "dns01-solver@prod-project.iam.gserviceaccount.com"
    },
    "ksaNamespace": "cert-manager",
    "ksaName": "cert-manager"
  },
  "providerConfig": {
    "credential": {
      "valueFrom": {
        "env": "GOOGLE_APPLICATION_CREDENTIALS"
      }
    }
  }
}
```

## Stack Outputs

After successful deployment, the following outputs are available:

| Output | Type | Description |
|--------|------|-------------|
| `member` | string | The IAM member string (e.g., `serviceAccount:prod-project.svc.id.goog[cert-manager/cert-manager]`) |
| `service_account_email` | string | The GSA email (echoed from input) |

### Accessing Outputs

```bash
# Using Pulumi CLI
pulumi stack output member
pulumi stack output service_account_email

# Using Planton CLI
# Outputs are automatically captured in the component's status
```

## Implementation Details

### Principal String Construction

The module constructs the Workload Identity member string using the formula:

```go
member := fmt.Sprintf(
    "serviceAccount:%s.svc.id.goog[%s/%s]",
    projectId,
    ksaNamespace,
    ksaName,
)
```

This eliminates manual string construction errors that are common when using low-level Terraform/Pulumi resources.

### IAM Binding Resource

```go
serviceaccount.NewIAMMember(
    ctx,
    "workload-identity-binding",
    &serviceaccount.IAMMemberArgs{
        ServiceAccountId: pulumi.String(gsaEmail),
        Role:             pulumi.String("roles/iam.workloadIdentityUser"),
        Member:           pulumi.String(member),
    },
    pulumi.Provider(gcpProvider),
)
```

The `IAMMember` resource is **idempotent** and **additive**:
- Multiple bindings can exist on the same GSA
- Deleting this resource removes only this specific binding
- No risk of removing other IAM bindings

## Troubleshooting

### "Permission denied" Errors

**Symptom**: Pods cannot access GCP APIs

**Checklist**:

1. **Verify IAM binding exists**:
   ```bash
   gcloud iam service-accounts get-iam-policy <gsa-email>
   ```
   Look for the member string and `roles/iam.workloadIdentityUser`

2. **Check Pulumi outputs**:
   ```bash
   pulumi stack output member
   ```
   Verify the member string matches your expectations

3. **Verify node pool configuration**:
   ```bash
   gcloud container node-pools describe <node-pool> --cluster <cluster>
   ```
   Ensure `workloadMetadataConfig.mode: GKE_METADATA`

### "Resource already exists" Errors

**Symptom**: Pulumi fails with "IAM binding already exists"

**Cause**: Another process or manual command created an identical binding

**Solution**: Import the existing resource:

```bash
pulumi import gcp:serviceaccount/iAMMember:IAMMember workload-identity-binding \
  <service-account-id> roles/iam.workloadIdentityUser <member-string>
```

### Debugging with Pulumi Logs

Enable verbose logging:

```bash
pulumi up --logtostderr --logflow -v=9
```

## Testing

### Unit Tests

```bash
# Run Go tests
cd module
go test -v ./...
```

### Integration Tests

```bash
# Deploy test stack
export PULUMI_STACK_INPUT='...'  # Test input
pulumi up --yes

# Verify binding
gcloud iam service-accounts get-iam-policy <gsa-email>

# Clean up
pulumi destroy --yes
```

## Performance

- **Deployment time**: 5-15 seconds (IAM binding creation)
- **Update time**: 5-10 seconds (IAM binding modification)
- **Deletion time**: 5-10 seconds (IAM binding removal)

IAM bindings are lightweight resources with fast operations.

## Security Considerations

### Least Privilege

The binding itself grants only `roles/iam.workloadIdentityUser`, which allows impersonation. The actual GCP permissions are controlled by the GSA's IAM roles.

**Best practice**: Create one GSA per application with minimal required permissions.

### Namespace Isolation

The member string explicitly includes the namespace, providing perfect isolation:

```
serviceAccount:project.svc.id.goog[namespace-a/ksa]  ≠  serviceAccount:project.svc.id.goog[namespace-b/ksa]
```

### Audit Trail

All impersonation actions are logged in Cloud Audit Logs with full KSA attribution.

## Related Documentation

- **[Component README](../../README.md)**: User-facing overview and examples
- **[Architecture Overview](overview.md)**: Design decisions and implementation details
- **[Research Documentation](../../docs/README.md)**: Deep dive into Workload Identity
- **[Examples](../../examples.md)**: Copy-paste ready examples

## Support

For issues or questions:
- Component-specific: See [main README](../../README.md)
- Pulumi-specific: [Pulumi Documentation](https://www.pulumi.com/docs/)
- Planton: [Planton Documentation](https://planton.dev)


