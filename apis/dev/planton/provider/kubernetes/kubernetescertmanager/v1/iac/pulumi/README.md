# Pulumi Module: Kubernetes Cert-Manager

This Pulumi module deploys kubernetes-cert-manager to a Kubernetes cluster with automatic ClusterIssuer creation based on DNS provider configurations.

## Prerequisites

- Kubernetes cluster with sufficient resources
- Pulumi CLI installed
- Planton CLI installed
- DNS provider credentials (Cloudflare API token, GCP/AWS/Azure workload identity configured)

## Usage

### Basic Deployment

```shell
# Set the module path
export KUBERNETES_CERT_MANAGER_MODULE=/path/to/apis/.../kubernetescertmanager/v1/iac/pulumi

# Deploy
planton pulumi up --manifest hack/manifest.yaml --module-dir ${KUBERNETES_CERT_MANAGER_MODULE}
```

### What Gets Deployed

This module creates:

1. **Kubernetes Namespace**: Created if `createNamespace: true`, otherwise looks up existing namespace (default: `kubernetes-cert-manager`)
2. **ServiceAccount**: With workload identity annotations for GCP/AWS/Azure providers
3. **Helm Release**: kubernetes-cert-manager chart from Jetstack
4. **Kubernetes Secrets**: For Cloudflare API tokens (one per Cloudflare provider)
5. **ClusterIssuers**: One per domain across all DNS providers

**Note**: By default (`createNamespace: false`), the namespace must exist before deployment. Set `createNamespace: true` in the manifest to have the module create it automatically.

### Environment Variables

No environment variables are required. All configuration comes from the manifest file.

### DNS Provider Credentials

**Cloudflare:**
- API token is stored in Kubernetes Secret automatically
- Secret name: `kubernetes-cert-manager-<provider-name>-credentials`

**GCP Cloud DNS:**
- Requires Workload Identity binding between GCP SA and Kubernetes SA
- ServiceAccount is annotated with `iam.gke.io/gcp-service-account`

**AWS Route53:**
- Requires IRSA (IAM Roles for Service Accounts) configured
- ServiceAccount is annotated with `eks.amazonaws.com/role-arn`

**Azure DNS:**
- Requires Managed Identity and workload identity federation
- ServiceAccount is annotated with `azure.workload.identity/client-id`

## Deployment Commands

### Initialize

```shell
planton pulumi preview --manifest manifest.yaml --module-dir ${KUBERNETES_CERT_MANAGER_MODULE}
```

### Deploy

```shell
planton pulumi up --manifest manifest.yaml --module-dir ${KUBERNETES_CERT_MANAGER_MODULE}
```

### Destroy

```shell
planton pulumi destroy --manifest manifest.yaml --module-dir ${KUBERNETES_CERT_MANAGER_MODULE}
```

## Outputs

The module exports the following outputs:

- `namespace`: Kubernetes namespace where cert-manager is deployed
- `release_name`: Helm release name (`cert-manager`)
- `cluster_issuer_names`: List of ClusterIssuer names (one per domain)

## Verification

After deployment, verify cert-manager is running:

```shell
# Check pods
kubectl get pods -n kubernetes-cert-manager

# Verify ClusterIssuers were created
kubectl get clusterissuers

# Check a specific ClusterIssuer
kubectl describe clusterissuer <domain-name>
```

## Troubleshooting

### Helm Release Failed

Check Pulumi logs:
```shell
pulumi stack --show-urns
pulumi logs
```

### ClusterIssuer Not Ready

Check cert-manager controller logs:
```shell
kubectl logs -n kubernetes-cert-manager -l app=cert-manager -f
```

### DNS Provider Authentication Failed

**Cloudflare:**
```shell
# Verify secret was created
kubectl get secret -n kubernetes-cert-manager | grep credentials

# Check secret content (base64 encoded)
kubectl get secret kubernetes-cert-manager-<provider-name>-credentials -n kubernetes-cert-manager -o yaml
```

**GCP/AWS/Azure:**
```shell
# Verify ServiceAccount annotations
kubectl get sa cert-manager -n kubernetes-cert-manager -o yaml

# Check workload identity is configured correctly
kubectl describe sa cert-manager -n kubernetes-cert-manager
```

## Module Structure

- `main.go`: Entrypoint that calls the module
- `module/main.go`: Core resource provisioning logic
- `module/outputs.go`: Output constants
- `module/vars.go`: Default values and constants
- `Pulumi.yaml`: Pulumi project configuration
- `Makefile`: Build automation
- `debug.sh`: Debug helper script

## Advanced Configuration

### Custom Helm Values

The module uses opinionated Helm values:
- `installCRDs: true` - Installs CRDs automatically
- `dns01-recursive-nameservers-only` - Uses public DNS for propagation checks
- `dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53` - Uses Cloudflare and Google DNS

### Multi-Provider Setup

The module supports multiple DNS providers in a single deployment. Each provider's zones get their own ClusterIssuer resources.

### ClusterIssuer Naming

ClusterIssuers are named after the domain they manage (e.g., `example.com`). This makes it easy to reference them in Certificate resources.

## Notes

- Minimum cert-manager version: v1.16.4
- DNS propagation can take 60-120 seconds
- Use Let's Encrypt staging for testing to avoid rate limits
- ClusterIssuers are cluster-wide resources (not namespaced)

