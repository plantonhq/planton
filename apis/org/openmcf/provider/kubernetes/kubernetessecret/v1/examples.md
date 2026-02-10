# Kubernetes Secret Examples

This document provides practical examples for deploying Kubernetes Secrets using OpenMCF.

## Understanding Target Cluster

All examples include the `target_cluster` field, which specifies which Kubernetes cluster the secret should be created in. This field consists of:

- `cluster_kind`: The type of Kubernetes cluster (e.g., `gcp_gke_cluster`, `aws_eks_cluster`, `azure_aks_cluster`, etc.)
- `cluster_name`: The name (or slug) of the Kubernetes cluster in your environment

**Supported cluster kinds:**
- `azure_aks_cluster` (400) - Azure AKS
- `aws_eks_cluster` (207) - AWS EKS
- `gcp_gke_cluster` (607) - GCP GKE
- `digital_ocean_kubernetes_cluster` (1208) - DigitalOcean Kubernetes
- `civo_kubernetes_cluster` (1507) - Civo Kubernetes

The cluster specified must exist in the same environment as the secret resource.

## Example 1: Minimal Opaque Secret

The simplest possible secret with a single key-value pair:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: my-api-key
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: my-api-key
  opaque:
    data:
      api-key: "sk-abc123def456"
```

**CLI Commands:**

```bash
# Validate the manifest
openmcf validate --manifest my-api-key.yaml

# Deploy with Pulumi
openmcf pulumi up \
  --manifest my-api-key.yaml \
  --stack myorg/myproject/dev

# Deploy with Terraform
openmcf tofu apply \
  --manifest my-api-key.yaml \
  --auto-approve
```

## Example 2: Opaque Secret with Multiple Keys

Database credentials stored as an Opaque secret with multiple key-value pairs:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: myapp-db-credentials
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: myapp-db-credentials
  namespace: myapp-production
  labels:
    team: backend
    environment: production
    purpose: database-credentials
  immutable: true
  opaque:
    data:
      DB_HOST: "postgres.internal.example.com"
      DB_PORT: "5432"
      DB_NAME: "myapp_production"
      DB_USER: "myapp_svc"
      DB_PASSWORD: "s3cur3-p@ssw0rd"
```

**What this creates:**

- An immutable Opaque secret in the `myapp-production` namespace
- Five key-value pairs accessible as environment variables or volume mounts
- Labels for governance and cost tracking

## Example 3: TLS Certificate Secret

Store a TLS certificate and private key for use with Ingress controllers:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: myapp-tls-cert
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: myapp-tls-cert
  namespace: myapp-production
  labels:
    purpose: tls-termination
    domain: app.example.com
  tls:
    tls_crt: |
      -----BEGIN CERTIFICATE-----
      MIIFjTCCA3WgAwIBAgIUK+example+certificate+data...
      -----END CERTIFICATE-----
    tls_key: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBgkqhkiG9w0BAQE+example+key+data...
      -----END PRIVATE KEY-----
```

**What this creates:**

- A `kubernetes.io/tls` type secret with `tls.crt` and `tls.key` data keys
- Can be referenced by Ingress resources for TLS termination

**Usage in Ingress:**

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-ingress
spec:
  tls:
    - hosts:
        - app.example.com
      secretName: myapp-tls-cert
```

## Example 4: Docker Registry Credentials

Authenticate with a private container registry for image pulls:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: ghcr-registry-creds
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: ghcr-registry-creds
  namespace: myapp-production
  labels:
    purpose: image-pull
    registry: ghcr.io
  docker_config_json:
    registry_server: "ghcr.io"
    username: "myorg-bot"
    password: "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    email: "bot@myorg.com"
```

**What this creates:**

- A `kubernetes.io/dockerconfigjson` type secret
- The IaC module constructs the `.dockerconfigjson` JSON automatically from the structured fields
- Can be referenced as an `imagePullSecret` in pods or service accounts

**Usage in Pod:**

```yaml
apiVersion: v1
kind: Pod
spec:
  imagePullSecrets:
    - name: ghcr-registry-creds
  containers:
    - name: myapp
      image: ghcr.io/myorg/myapp:latest
```

## Example 5: GCR Service Account Key for Image Pulls

Authenticate with Google Container Registry using a service account:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: gcr-registry-creds
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: gcr-registry-creds
  namespace: myapp-production
  docker_config_json:
    registry_server: "gcr.io"
    username: "_json_key"
    password: '{"type":"service_account","project_id":"my-project"}'
```

## Example 6: Basic Authentication Secret

Username/password credentials for services that use HTTP Basic Auth:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: grafana-admin-creds
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: monitoring-cluster
  name: grafana-admin-creds
  namespace: monitoring
  labels:
    purpose: admin-credentials
    app: grafana
  basic_auth:
    username: "admin"
    password: "s3cur3-gr@fana-p@ss"
```

**What this creates:**

- A `kubernetes.io/basic-auth` type secret with `username` and `password` data keys

## Example 7: SSH Authentication Secret

SSH private key for Git operations or SSH-based authentication:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: deploy-ssh-key
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: deploy-ssh-key
  namespace: argocd
  labels:
    purpose: git-ssh
    app: argocd
  annotations:
    argocd.argoproj.io/secret-type: repository
  ssh_auth:
    ssh_private_key: |
      -----BEGIN OPENSSH PRIVATE KEY-----
      b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAA+example+key+data...
      -----END OPENSSH PRIVATE KEY-----
```

**What this creates:**

- A `kubernetes.io/ssh-auth` type secret with `ssh-privatekey` data key
- Annotations for ArgoCD repository configuration

## Example 8: Immutable Secret with Custom Namespace

Immutable configuration secret that cannot be modified after creation:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSecret
metadata:
  name: payment-api-config
spec:
  target_cluster:
    cluster_kind: gcp_gke_cluster
    cluster_name: prod-cluster
  name: payment-api-config
  namespace: payment-services
  labels:
    team: payments
    environment: production
    compliance: pci-dss
    criticality: critical
  annotations:
    description: "Immutable payment gateway credentials"
  immutable: true
  opaque:
    data:
      STRIPE_API_KEY: "sk_live_xxxxxxxxxxxxxxxxxxxx"
      STRIPE_WEBHOOK_SECRET: "whsec_xxxxxxxxxxxxxxxxxxxx"
      PAYMENT_GATEWAY_URL: "https://api.stripe.com/v1"
```

## Verification

After deployment, verify the secret was created correctly:

```bash
# Check secret exists
kubectl get secret <secret-name> -n <namespace>

# Check secret type
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.type}'

# Check secret keys (without revealing values)
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.data}' | jq 'keys'

# Check labels and annotations
kubectl get secret <secret-name> -n <namespace> -o yaml | head -20

# Check immutability
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.immutable}'

# Check stack outputs (Pulumi)
openmcf pulumi stack output --manifest <file> --stack <stack>
```

## Common Patterns

### Pattern 1: Environment-Specific Secrets

Deploy the same secret structure across environments with different values:

```yaml
metadata:
  name: myapp-secrets-{env}
spec:
  namespace: myapp-{env}
  labels:
    app: myapp
    environment: {env}
  immutable: true
  opaque:
    data:
      API_KEY: "{env-specific-value}"
```

### Pattern 2: Registry Credentials per Namespace

Deploy image pull secrets to namespaces that need private registry access:

```yaml
metadata:
  name: registry-creds
spec:
  namespace: {target-namespace}
  docker_config_json:
    registry_server: "{registry-url}"
    username: "{username}"
    password: "{token}"
```

### Pattern 3: TLS Certificates for Ingress

Pair with an Ingress resource for TLS termination:

```yaml
spec:
  namespace: {ingress-namespace}
  tls:
    tls_crt: "{pem-certificate}"
    tls_key: "{pem-private-key}"
```

## Troubleshooting

**Issue: Secret not visible to pods**

```bash
# Verify namespace matches
kubectl get secret -n <namespace>

# Check pod's namespace
kubectl get pod <pod-name> -o jsonpath='{.metadata.namespace}'
```

**Issue: Image pull fails with registry secret**

```bash
# Verify secret type
kubectl get secret <name> -n <ns> -o jsonpath='{.type}'
# Should be: kubernetes.io/dockerconfigjson

# Check if secret is referenced in pod or service account
kubectl get pod <pod> -o jsonpath='{.spec.imagePullSecrets}'
kubectl get sa default -n <ns> -o jsonpath='{.imagePullSecrets}'
```

**Issue: Cannot update immutable secret**

```bash
# Immutable secrets cannot be updated, only deleted and recreated
# Delete first, then redeploy
openmcf pulumi destroy --manifest <file> --stack <stack>
openmcf pulumi up --manifest <file> --stack <stack>
```

## Next Steps

1. Review the [README.md](README.md) for component documentation
2. Check the [research documentation](docs/README.md) for architecture deep-dive
3. Deploy your first secret using one of the examples above
4. Set up CI/CD pipelines to inject secret values at deploy time
