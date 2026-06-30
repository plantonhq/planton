# Pulumi Module: Kubernetes External Secrets

This Pulumi module deploys the External Secrets Operator (ESO) to a Kubernetes cluster with cloud-native authentication (Workload Identity, IRSA, or Managed Identity).

## Prerequisites

- Kubernetes cluster (GKE, EKS, AKS, or any K8s cluster)
- Pulumi CLI installed
- Planton CLI installed
- Cloud provider credentials configured:
  - **GKE**: Workload Identity enabled and GCP service account with Secret Manager permissions
  - **EKS**: OIDC provider configured and IAM role for IRSA
  - **AKS**: Azure AD Workload Identity enabled and Managed Identity with Key Vault permissions

## Usage

### Basic Deployment

```shell
# Set the module path
export KUBERNETES_EXTERNAL_SECRETS_MODULE=/path/to/apis/.../kubernetesexternalsecrets/v1/iac/pulumi

# Deploy
planton pulumi up --manifest manifest.yaml --module-dir ${KUBERNETES_EXTERNAL_SECRETS_MODULE}
```

### What Gets Deployed

This module creates:

1. **Kubernetes Namespace** (conditional): 
   - Created when `create_namespace: true` (default)
   - Skipped when `create_namespace: false` (uses existing namespace)
2. **ServiceAccount**: With cloud provider annotations (Workload Identity, IRSA, or Managed Identity)
3. **Helm Release**: External Secrets Operator from the official ESO Helm chart
4. **ClusterSecretStore**: Automatically configured for your cloud provider
5. **RBAC Resources**: ClusterRole and ClusterRoleBinding for ESO controller

### Namespace Management

**New in this version**: You can now control whether the component creates the namespace or uses an existing one.

**Create namespace (default):**
```yaml
spec:
  namespace:
    value: external-secrets
  create_namespace: true  # or omit this field
```

**Use existing namespace:**
```yaml
spec:
  namespace:
    value: platform-services
  create_namespace: false  # namespace must already exist
```

See the main [README](../../README.md#namespace-management) for detailed namespace management patterns.

### Cloud Provider Authentication

**GKE (Google Cloud Secret Manager):**
- Uses Workload Identity to bind Kubernetes ServiceAccount to GCP Service Account
- ServiceAccount is annotated with `iam.gke.io/gcp-service-account`
- GCP SA must have `secretmanager.secretAccessor` role

**EKS (AWS Secrets Manager):**
- Uses IRSA (IAM Roles for Service Accounts)
- ServiceAccount is annotated with `eks.amazonaws.com/role-arn`
- IAM role is auto-created with SecretsManager permissions (or you can provide your own)

**AKS (Azure Key Vault):**
- Uses Azure Workload Identity (federated credentials)
- ServiceAccount is annotated with `azure.workload.identity/client-id`
- Managed Identity must have "Key Vault Secrets User" role

## Deployment Commands

### Preview Changes

```shell
planton pulumi preview --manifest manifest.yaml --module-dir ${KUBERNETES_EXTERNAL_SECRETS_MODULE}
```

### Deploy

```shell
planton pulumi up --manifest manifest.yaml --module-dir ${KUBERNETES_EXTERNAL_SECRETS_MODULE}
```

### Destroy

```shell
planton pulumi destroy --manifest manifest.yaml --module-dir ${KUBERNETES_EXTERNAL_SECRETS_MODULE}
```

## Example Manifests

### GKE Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: external-secrets-gke-prod
spec:
  target_cluster:
    kubernetes_cluster_id:
      value: prod-gke-cluster
  
  poll_interval_seconds: 30
  
  container:
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 50m
        memory: 100Mi
  
  gke:
    project_id:
      value: my-gcp-project
    gsa_email: external-secrets@my-gcp-project.iam.gserviceaccount.com
```

### EKS Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: external-secrets-eks-prod
spec:
  target_cluster:
    kubernetes_cluster_id:
      value: prod-eks-cluster
  
  poll_interval_seconds: 30
  
  container:
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 50m
        memory: 100Mi
  
  eks:
    region: us-east-1
```

### AKS Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesExternalSecrets
metadata:
  name: external-secrets-aks-prod
spec:
  target_cluster:
    kubernetes_cluster_id:
      value: prod-aks-cluster
  
  poll_interval_seconds: 30
  
  container:
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
      requests:
        cpu: 50m
        memory: 100Mi
  
  aks:
    key_vault_resource_id: /subscriptions/.../vaults/prod-keyvault
    managed_identity_client_id: 87654321-4321-4321-4321-210987654321
```

## Outputs

The module exports the following outputs:

- `namespace`: Kubernetes namespace where ESO is deployed
- `release_name`: Helm release name (matches `metadata.name`)
- `service_account_name`: ServiceAccount name used by ESO controller
- `cluster_secret_store_name`: Name of the automatically created ClusterSecretStore

## Verification

After deployment, verify ESO is running:

```shell
# Check ESO pods (should see controller, webhook, and cert-controller)
kubectl get pods -n external-secrets

# Verify ClusterSecretStore was created
kubectl get clustersecretstore

# Check ClusterSecretStore status
kubectl describe clustersecretstore <name>

# Verify ServiceAccount has cloud provider annotation
kubectl get sa -n external-secrets <service-account-name> -o yaml
```

### Create a Test ExternalSecret

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: test-secret
  namespace: default
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: <cluster-secret-store-name>
    kind: ClusterSecretStore
  target:
    name: test-secret-data
    creationPolicy: Owner
  data:
  - secretKey: my-key
    remoteRef:
      key: test-secret-name  # Replace with actual secret name in your cloud provider
```

```shell
# Apply the test ExternalSecret
kubectl apply -f test-externalsecret.yaml

# Wait for sync
kubectl get externalsecret test-secret -w

# Verify Kubernetes Secret was created
kubectl get secret test-secret-data

# View secret data (base64 encoded)
kubectl get secret test-secret-data -o yaml
```

## Troubleshooting

### Helm Release Failed

Check Pulumi logs:
```shell
pulumi stack --show-urns
pulumi logs
```

### ESO Controller Not Running

Check ESO controller logs:
```shell
kubectl logs -n external-secrets -l app.kubernetes.io/name=external-secrets -f
```

### ClusterSecretStore Not Ready

```shell
# Check ClusterSecretStore status
kubectl describe clustersecretstore <name>

# Look for validation errors or auth failures
kubectl logs -n external-secrets -l app.kubernetes.io/name=external-secrets --tail=100
```

### Namespace Already Exists Error

```
Error: namespaces "external-secrets" already exists
```

**Fix**: Set `create_namespace: false` in your manifest to use the existing namespace instead of trying to create it.

```yaml
spec:
  namespace:
    value: external-secrets
  create_namespace: false  # Use existing namespace
```

### Authentication Failures

**GKE (Workload Identity):**
```shell
# Verify ServiceAccount annotation
kubectl get sa -n external-secrets external-secrets -o jsonpath='{.metadata.annotations.iam\.gke\.io/gcp-service-account}'

# Check GCP IAM binding
gcloud iam service-accounts get-iam-policy \
  external-secrets@my-project.iam.gserviceaccount.com

# Test secret access manually
gcloud secrets versions access latest --secret=test-secret
```

**EKS (IRSA):**
```shell
# Verify ServiceAccount annotation
kubectl get sa -n external-secrets external-secrets -o jsonpath='{.metadata.annotations.eks\.amazonaws\.com/role-arn}'

# Check IAM role trust policy
aws iam get-role --role-name <role-name>

# Verify IAM policy
aws iam list-role-policies --role-name <role-name>
aws iam get-role-policy --role-name <role-name> --policy-name <policy-name>
```

**AKS (Managed Identity):**
```shell
# Verify ServiceAccount annotation
kubectl get sa -n external-secrets external-secrets -o jsonpath='{.metadata.annotations.azure\.workload\.identity/client-id}'

# Check Managed Identity permissions
az role assignment list --assignee <managed-identity-client-id>

# Verify federated credential exists
az identity federated-credential list \
  --identity-name <identity-name> \
  --resource-group <resource-group>
```

### ExternalSecret Sync Failures

```shell
# Check ExternalSecret status
kubectl describe externalsecret <name> -n <namespace>

# Common errors:
# - "secret not found" → Verify secret exists in cloud provider
# - "permission denied" → Check IAM/RBAC permissions
# - "invalid credentials" → Verify cloud authentication is configured correctly

# Check ESO controller logs for detailed errors
kubectl logs -n external-secrets -l app.kubernetes.io/name=external-secrets --tail=200
```

## Module Structure

- `main.go`: Entrypoint that calls the module
- `module/main.go`: Core resource provisioning logic (Helm chart, ServiceAccount, ClusterSecretStore)
- `module/outputs.go`: Output constants and stack exports
- `module/vars.go`: Default values, chart versions, and constants
- `Pulumi.yaml`: Pulumi project configuration
- `Makefile`: Build automation (optional)
- `debug.sh`: Debug helper script (optional)

## Advanced Configuration

### Poll Interval Tuning

The `poll_interval_seconds` controls how often ESO checks for secret updates:

- **10-30 seconds**: Fast propagation, higher cloud API costs
- **1-4 hours**: Cost-effective for infrequently changing secrets
- **Consider**: Balance secret freshness requirements with cloud provider API pricing

### Resource Tuning

Adjust `container.resources` based on cluster size and number of secrets:

- **Small clusters** (< 50 secrets): 50m CPU / 100Mi memory
- **Medium clusters** (50-200 secrets): 100m CPU / 256Mi memory
- **Large clusters** (200+ secrets): 200m+ CPU / 512Mi+ memory

### High Availability

For production, deploy multiple ESO controller replicas:
- The module enables leader election automatically
- Use pod anti-affinity to spread replicas across nodes
- Configure PodDisruptionBudgets for resilience

### Multiple SecretStores

After deployment, create additional SecretStores for:
- Different cloud accounts/projects
- Different namespaces with scoped credentials
- Cross-region or cross-provider scenarios

## Cloud Provider Setup Guides

### GKE Prerequisites

```shell
# 1. Enable Secret Manager API
gcloud services enable secretmanager.googleapis.com --project=my-project

# 2. Create GCP service account
gcloud iam service-accounts create external-secrets \
  --display-name="External Secrets Operator" \
  --project=my-project

# 3. Grant Secret Manager access
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:external-secrets@my-project.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# 4. Create Workload Identity binding (done automatically by module during deployment)
gcloud iam service-accounts add-iam-policy-binding \
  external-secrets@my-project.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:my-project.svc.id.goog[external-secrets/external-secrets]"
```

### EKS Prerequisites

```shell
# 1. Ensure OIDC provider exists for cluster
eksctl utils associate-iam-oidc-provider \
  --cluster=prod-eks-cluster \
  --approve

# 2. IAM role is auto-created by the module with appropriate permissions
# Or, create manually if using irsa_role_arn_override
```

### AKS Prerequisites

```shell
# 1. Enable Azure AD Workload Identity on cluster
az aks update \
  --resource-group prod-rg \
  --name prod-aks-cluster \
  --enable-oidc-issuer \
  --enable-workload-identity

# 2. Create User-Assigned Managed Identity
az identity create \
  --name external-secrets-identity \
  --resource-group prod-rg

# 3. Grant Key Vault permissions
az role assignment create \
  --assignee <managed-identity-principal-id> \
  --role "Key Vault Secrets User" \
  --scope <key-vault-resource-id>

# 4. Create federated credential (done automatically during ESO setup)
# Or create manually before deploying
```

## Cost Optimization

Monitor cloud API usage:

```
Monthly API Calls = (Number of Secrets) × (86400 / poll_interval_seconds) × 30

Example: 100 secrets with 30-second poll
= 100 × (86400 / 30) × 30
= 8,640,000 API calls/month

AWS Secrets Manager: 8,640,000 / 10,000 × $0.05 = $43.20/month
GCP Secret Manager: First 6 API calls per secret per month are free
Azure Key Vault: $0.03 per 10,000 transactions = $25.92/month
```

**Recommendation:** Start with 1-hour polling (3600 seconds) for most secrets, use shorter intervals only for critical dynamic credentials.

## Notes

- ESO version is managed by the Helm chart version in the module
- Minimum supported Kubernetes version: 1.19+
- For multi-cloud scenarios, you can deploy multiple KubernetesExternalSecrets instances with different provider configs
- ClusterSecretStore is cluster-wide; use namespaced SecretStore for finer-grained access control
- Secrets synced to Kubernetes are stored in etcd (encrypted at rest if KMS is enabled)

## Additional Resources

- [External Secrets Operator Documentation](https://external-secrets.io/)
- [API Reference](../../README.md)
- [Examples](../../examples.md)
- [Research Documentation](../../docs/README.md)


