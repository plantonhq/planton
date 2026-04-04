# KubernetesCertManager

> Automated TLS certificate management for your Kubernetes clusters with multi-provider DNS support

## Overview

Think of kubernetes-cert-manager as your cluster's personal certificate authority liaison. It watches your Certificate resources, talks to Let's Encrypt (or other ACME servers) on your behalf, proves you own the domains using DNS challenges, and keeps your TLS certificates fresh and valid—all automatically.

The `KubernetesCertManager` addon deploys and configures kubernetes-cert-manager with **automatic ClusterIssuer creation** based on your DNS provider configurations. Whether you're managing domains across Cloudflare, Google Cloud DNS, AWS Route53, or Azure DNS—or even a mix of multiple providers—this addon handles the entire DNS provider integration and ClusterIssuer setup for you.

### Why DNS-01 Challenges?

While HTTP-01 challenges are simpler (just serve a file over HTTP), DNS-01 challenges unlock powerful capabilities:

- **Wildcard certificates**: Get `*.example.com` coverage in a single cert
- **Internal services**: No need to expose services publicly for validation
- **Multi-cloud flexibility**: Use any DNS provider regardless of where your cluster runs
- **Centralized DNS**: Manage all domains in one place (e.g., Cloudflare) across multiple clusters

The tradeoff? You need to give kubernetes-cert-manager permission to create DNS TXT records. This addon handles that credential plumbing for you.

### What's New: Multi-Provider Support

Unlike traditional kubernetes-cert-manager setups where you manually create ClusterIssuers, this addon:

- ✅ **Automatically creates a ClusterIssuer** with multiple DNS solvers
- ✅ **Supports multiple DNS providers** in a single deployment
- ✅ **Manages DNS provider credentials** as Kubernetes Secrets
- ✅ **Configures optimal DNS propagation** using public recursive nameservers
- ✅ **Enforces minimum kubernetes-cert-manager version** (v1.16.4+) for stability

You simply specify your ACME email and list your DNS providers—the addon handles the rest.

## Quick Start

Here's the minimal configuration to get kubernetes-cert-manager running with Cloudflare:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: kubernetes-cert-manager
spec:
  targetCluster:
    clusterName: "my-gke-cluster"
  namespace:
    value: "kubernetes-cert-manager"
  
  # Global ACME configuration
  acme:
    email: "admin@example.com"
    server: "https://acme-v02.api.letsencrypt.org/directory"
  
  # DNS provider configurations
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
        - example.org
      cloudflare:
        apiToken: "your-cloudflare-api-token"
```

Deploy it:

```bash
export KUBERNETES_CERT_MANAGER_MODULE=/path/to/apis/.../kubernetescertmanager/v1/iac/pulumi

openmcf pulumi up \
  --manifest kubernetes-cert-manager.yaml \
  --module-dir ${KUBERNETES_CERT_MANAGER_MODULE}
```

**That's it!** The addon automatically creates:
- kubernetes-cert-manager Helm deployment
- Kubernetes Secret with your Cloudflare API token
- ClusterIssuer named `letsencrypt-cluster-issuer` with Cloudflare DNS-01 solver

You can now immediately request certificates—no manual ClusterIssuer creation needed.

## Namespace Management

The addon provides control over namespace creation through the `createNamespace` field, giving you flexibility in how namespaces are managed.

### Default Behavior (createNamespace: false)

By default, the addon **expects the namespace to already exist** before deployment:

- The namespace must be created separately (either manually or by another resource)
- Useful when namespaces are managed centrally or require specific configurations
- Follows GitOps best practices where infrastructure is explicitly declared
- Deployment will fail if the namespace doesn't exist (fail-fast behavior)

**Example with existing namespace**:
```yaml
spec:
  namespace:
    value: "cert-manager"  # Must already exist
  createNamespace: false   # Default value
  # ... rest of configuration
```

### Automatic Creation (createNamespace: true)

When set to `true`, the addon creates the namespace during deployment:

- Namespace is created with basic metadata only
- Convenient for quick deployments and testing
- Namespace will be deleted if the addon is destroyed

**Example with automatic creation**:
```yaml
spec:
  namespace:
    value: "cert-manager"  # Will be created
  createNamespace: true
  # ... rest of configuration
```

### When to Use Each Approach

**Use `createNamespace: false` (default) when**:
- Namespaces have specific labels, annotations, or resource quotas
- Following GitOps practices with namespace-as-code
- Multiple components share the same namespace
- Namespace lifecycle should be independent of the addon

**Use `createNamespace: true` when**:
- Quick testing or proof-of-concept deployments
- Namespace is dedicated to this addon only
- Simple deployments without special namespace requirements
- You want the addon to manage the complete lifecycle

## Configuration Reference

### Common Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster` | object | required | Identifies which Kubernetes cluster to install on |
| `namespace` | string | `"kubernetes-cert-manager"` | Kubernetes namespace where kubernetes-cert-manager will be deployed |
| `createNamespace` | bool | `false` | Whether to create the namespace (false=use existing, true=create new) |
| `kubernetesCertManagerVersion` | string | `"v1.19.1"` | kubernetes-cert-manager version (minimum v1.16.4) |
| `helmChartVersion` | string | `"v1.19.1"` | Helm chart version to deploy |
| `skipInstallSelfSignedIssuer` | bool | `false` | Skip creating a self-signed ClusterIssuer |

### ACME Configuration

Global ACME settings used for all DNS providers:

```yaml
acme:
  email: "admin@example.com"  # Required: ACME account email for notifications
  server: "https://acme-v02.api.letsencrypt.org/directory"  # Optional: ACME server URL
```

**ACME Server Options**:
- **Production**: `https://acme-v02.api.letsencrypt.org/directory` (default)
- **Staging**: `https://acme-staging-v02.api.letsencrypt.org/directory` (for testing)

### DNS Provider Configurations

The `dnsProviders` field is an array of DNS provider configurations. Each provider can manage multiple DNS zones, and you can mix different provider types in a single deployment.

#### Cloudflare

Perfect for multi-cloud setups or when you want DNS management separate from your cloud provider.

```yaml
dnsProviders:
  - name: cloudflare-prod  # Unique identifier for this provider config
    dnsZones:
      - example.com
      - example.org
    cloudflare:
      apiToken: "cloudflare-api-token-here"  # Required
```

**What you need**:
- A Cloudflare API token with `Zone:Zone:Read` and `Zone:DNS:Edit` permissions
- Token scoped to the specific zones listed in `dnsZones`

**What gets created**:
- Kubernetes Secret: `kubernetes-cert-manager-cloudflare-prod-credentials` (in kubernetes-cert-manager namespace)
- ClusterIssuer solver configured for example.com and example.org

**How to get the token**:
1. Cloudflare Dashboard → Profile → API Tokens → Create Token
2. Use "Edit zone DNS" template
3. Scope to specific zones
4. Copy the token (shown only once)

#### Google Cloud DNS

For domains managed in Google Cloud DNS (works on any Kubernetes cluster, not just GKE):

```yaml
dnsProviders:
  - name: gcp-prod
    dnsZones:
      - internal.example.net
    gcpCloudDns:
      projectId: "my-gcp-project"
      serviceAccountEmail: "kubernetes-cert-manager@my-project.iam.gserviceaccount.com"
```

**What you need**:
- GCP service account with `dns.admin` role on the DNS zones
- Workload Identity binding between the GSA and kubernetes-cert-manager KSA (for GKE)
- For non-GKE clusters, service account key configured via workload identity federation

**What gets created**:
- ServiceAccount annotation: `iam.gke.io/gcp-service-account`
- ClusterIssuer solver configured for internal.example.net

#### AWS Route53

For domains managed in AWS Route53:

```yaml
dnsProviders:
  - name: aws-prod
    dnsZones:
      - aws.example.com
    awsRoute53:
      region: "us-east-1"
      roleArn: "arn:aws:iam::123456789012:role/kubernetes-cert-manager-role"
```

**What you need**:
- IAM role with Route53 permissions for the specified zones
- IRSA (IAM Roles for Service Accounts) configured for EKS
- For non-EKS clusters, appropriate AWS authentication configured

**What gets created**:
- ServiceAccount annotation: `eks.amazonaws.com/role-arn`
- ClusterIssuer solver configured for aws.example.com

#### Azure DNS

For domains managed in Azure DNS:

```yaml
dnsProviders:
  - name: azure-prod
    dnsZones:
      - azure.example.com
    azureDns:
      subscriptionId: "12345678-1234-1234-1234-123456789012"
      resourceGroup: "my-rg"
      clientId: "87654321-4321-4321-4321-210987654321"
```

**What you need**:
- Azure Managed Identity with DNS Zone Contributor role
- Workload identity federation configured for AKS
- For non-AKS clusters, appropriate Azure authentication configured

**What gets created**:
- ServiceAccount annotation: `azure.workload.identity/client-id`
- ClusterIssuer solver configured for azure.example.com

### Multi-Provider Example

The real power comes from combining multiple providers in a single deployment:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: kubernetes-cert-manager-multi
spec:
  targetCluster:
    clusterName: "my-gke-cluster"
  namespace:
    value: "kubernetes-cert-manager"
  
  acme:
    email: "admin@example.com"
    server: "https://acme-v02.api.letsencrypt.org/directory"
  
  dnsProviders:
    # Public domains managed by Cloudflare
    - name: cloudflare-public
      dnsZones:
        - example.com
        - example.org
      cloudflare:
        apiToken: "cloudflare-token-here"
    
    # Internal domains managed by Google Cloud DNS
    - name: gcp-internal
      dnsZones:
        - internal.example.net
      gcpCloudDns:
        projectId: "my-gcp-project"
        serviceAccountEmail: "kubernetes-cert-manager@my-project.iam.gserviceaccount.com"
    
    # AWS-hosted domains
    - name: aws-services
      dnsZones:
        - aws.example.com
      awsRoute53:
        region: "us-east-1"
        roleArn: "arn:aws:iam::123456789012:role/kubernetes-cert-manager"
```

This creates a single ClusterIssuer with three DNS solvers, automatically routing certificate requests to the correct DNS provider based on the domain name.

## Using kubernetes-cert-manager

### Understanding Auto-Created Resources

After deploying the addon, the following resources are automatically created:

| Resource | Name | Purpose |
|----------|------|---------|
| Namespace | `kubernetes-cert-manager` | Isolates kubernetes-cert-manager components |
| ServiceAccount | `kubernetes-cert-manager` | For running kubernetes-cert-manager pods (with cloud provider annotations) |
| Helm Release | `kubernetes-cert-manager` | The kubernetes-cert-manager controller, webhook, cainjector |
| Secrets | `kubernetes-cert-manager-<provider-name>-credentials` | For each Cloudflare provider |
| **ClusterIssuers** | **One per domain** (e.g., `planton.ai`, `planton.live`) | **Each domain gets its own ClusterIssuer for better visibility** |

**The key difference**: You **DO NOT** need to manually create ClusterIssuers. The addon creates **one ClusterIssuer per domain**, each named after the domain itself for easy identification.

### Requesting Certificates

Simply create Certificate resources referencing the domain-named ClusterIssuer:

```yaml
apiVersion: kubernetes-cert-manager.io/v1
kind: Certificate
metadata:
  name: planton-wildcard
  namespace: default
spec:
  secretName: planton-tls  # Where the cert will be stored
  issuerRef:
    name: planton.ai  # Use the domain name - auto-created by the addon
    kind: ClusterIssuer
  dnsNames:
    - planton.ai
    - "*.planton.ai"
```

Each domain gets its own ClusterIssuer named after the domain itself. This makes it crystal clear which issuer to use for each domain and provides better visibility when troubleshooting.

Within a few minutes, you'll have a Secret named `planton-tls` containing your certificate.

### Separate Certificates Per Domain

Since each domain has its own ClusterIssuer, you'll typically create separate certificates for each domain:

```yaml
# Certificate for planton.ai
apiVersion: kubernetes-cert-manager.io/v1
kind: Certificate
metadata:
  name: planton-cert
  namespace: default
spec:
  secretName: planton-tls
  issuerRef:
    name: planton.ai  # Domain-specific issuer
    kind: ClusterIssuer
  dnsNames:
    - planton.ai
    - "*.planton.ai"
---
# Certificate for planton.live
apiVersion: kubernetes-cert-manager.io/v1
kind: Certificate
metadata:
  name: planton-live-cert
  namespace: default
spec:
  secretName: planton-live-tls
  issuerRef:
    name: planton.live  # Domain-specific issuer
    kind: ClusterIssuer
  dnsNames:
    - planton.live
    - "*.planton.live"
```

This approach provides better isolation and makes it easier to track which certificates belong to which domain.

### Using Certificates in Ingress

Reference the certificate secret in your Ingress resources:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    # Optional: Let kubernetes-cert-manager auto-create the certificate
    kubernetes-cert-manager.io/cluster-issuer: planton.ai  # Use the domain name
spec:
  tls:
    - hosts:
        - app.planton.ai
      secretName: planton-tls
  rules:
    - host: app.planton.ai
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

## How It Works

### DNS-01 Challenge Flow

Think of the DNS-01 challenge as a proof-of-ownership test. Let's Encrypt says "prove you control `example.com`," and kubernetes-cert-manager responds by temporarily creating a specific TXT record that only the domain owner could create.

Here's the flow:

1. **You create a Certificate**: kubectl apply -f certificate.yaml
2. **kubernetes-cert-manager creates an ACME order**: Talks to Let's Encrypt, gets challenge details
3. **DNS TXT record appears**: kubernetes-cert-manager creates `_acme-challenge.example.com` TXT record via your DNS provider API
4. **Propagation check**: kubernetes-cert-manager queries public DNS (1.1.1.1, 8.8.8.8) to verify the record exists
5. **Let's Encrypt validates**: Queries public DNS to verify the TXT record exists
6. **Certificate issued**: Let's Encrypt signs your certificate
7. **Cleanup**: TXT record deleted, certificate stored in Kubernetes Secret
8. **Auto-renewal**: kubernetes-cert-manager repeats this process before expiration (default: 30 days before)

The entire process takes 1-3 minutes for most DNS providers. Cloudflare is typically fastest (~30 seconds), while Route53 can take up to 2 minutes due to DNS propagation.

### What Makes This Addon Reliable

The addon includes several best practices and fixes for common issues:

**DNS Propagation Reliability**:
- Configures kubernetes-cert-manager to use public recursive nameservers (Cloudflare 1.1.1.1 and Google 8.8.8.8)
- Avoids reliance on in-cluster DNS which may cache stale records
- Significantly reduces "waiting for DNS propagation" failures

**Version Enforcement**:
- Minimum kubernetes-cert-manager version v1.16.4
- Fixes critical Cloudflare API compatibility issues in older versions
- Prevents DNS record cleanup failures that cause stuck challenges

**Automatic Solver Selection**:
- kubernetes-cert-manager automatically routes certificate requests to the correct DNS provider
- Based on `dnsZones` configuration in your `dnsProviders`
- No manual configuration needed per certificate

## Security Best Practices

### Least Privilege DNS Access

Each provider has different scoping mechanisms:

**Cloudflare**:
- Scope API tokens to specific zones (not "All zones")
- Use only required permissions: `Zone:Zone:Read` + `Zone:DNS:Edit`
- Set token expiration and rotate quarterly

**GCP/AWS/Azure**:
- Grant DNS permissions only for specific zones (not wildcard)
- Use separate service accounts/roles per environment
- Enable audit logging for DNS changes

### Secret Management

Cloudflare API tokens live in Kubernetes Secrets. Protect them:

- **Never commit secrets to Git**: Use sealed secrets, external-secrets operator, or Vault
- **Limit RBAC access**: Only kubernetes-cert-manager needs read access to the secrets
- **Rotate credentials**: Update tokens/credentials regularly (quarterly minimum)
- **Use secret scanning**: Enable GitHub secret scanning or similar tools

For cloud providers (GCP/AWS/Azure), credentials are tied to workload identity—no long-lived secrets to manage.

### Certificate Issuance Hygiene

- **Start with staging**: Use Let's Encrypt staging environment until you're confident
- **Watch rate limits**: Let's Encrypt allows 50 certs per domain per week (production)
- **Consolidate domains**: Use SANs (Subject Alternative Names) to get multiple domains in one cert
- **Monitor expiration**: Set up alerts for certs expiring soon (though auto-renewal should handle it)

## Troubleshooting

### Verify Addon Deployment

Start here if certificates aren't being issued:

```bash
# Check kubernetes-cert-manager pods are healthy
kubectl get pods -n kubernetes-cert-manager

# Should see 3 pods running:
# - kubernetes-cert-manager-*
# - kubernetes-cert-manager-cainjector-*
# - kubernetes-cert-manager-webhook-*

# Verify the auto-created ClusterIssuers (one per domain)
kubectl get clusterissuer
kubectl describe clusterissuer planton.ai
kubectl describe clusterissuer planton.live

# Check for provider-specific resources
kubectl get secret -n kubernetes-cert-manager  # Look for kubernetes-cert-manager-*-credentials
kubectl get sa kubernetes-cert-manager -n kubernetes-cert-manager -o yaml  # Check cloud provider annotations
```

### Debug Certificate Flow

When a certificate isn't being issued, follow the resource chain:

```bash
# 1. Check the Certificate
kubectl describe certificate <name> -n <namespace>
# Look for: "The certificate has been successfully issued"

# 2. If not issued, check the CertificateRequest
kubectl get certificaterequest -n <namespace>
kubectl describe certificaterequest <name> -n <namespace>

# 3. Check the Order
kubectl get order -n <namespace>
kubectl describe order <name> -n <namespace>

# 4. Check the Challenge (most important for DNS-01)
kubectl get challenge -n <namespace>
kubectl describe challenge <name> -n <namespace>
# This shows DNS-01 challenge status and which solver was selected

# 5. Check kubernetes-cert-manager logs
kubectl logs -n kubernetes-cert-manager deployment/kubernetes-cert-manager -f
```

The status messages in `describe` output tell you exactly where things are stuck.

### Common Issues

#### "Waiting for DNS propagation"

**What's happening**: kubernetes-cert-manager created the DNS TXT record but can't verify it yet.

**Why**:
- DNS propagation takes time (30 seconds to 2 minutes typically)
- DNS provider API delays
- Cached DNS responses

**Fix**: Usually resolves itself. If stuck > 5 minutes:

```bash
# Verify TXT record exists manually
dig _acme-challenge.example.com TXT @1.1.1.1 +short

# Check kubernetes-cert-manager can talk to DNS provider
kubectl logs -n kubernetes-cert-manager deployment/kubernetes-cert-manager | grep -i dns
```

#### Authentication/Permission Errors

**Cloudflare**: "Error 403: Forbidden" or "Invalid API token"
- Token expired or invalid
- Missing `Zone:Zone:Read` or `Zone:DNS:Edit` permissions
- Token not scoped to the domain's zone
- Check: `kubectl get secret kubernetes-cert-manager-<provider-name>-credentials -n kubernetes-cert-manager -o yaml`

**GCP**: "PermissionDenied"
- Service account missing `dns.admin` role
- Workload Identity not configured properly
- Wrong GCP project specified
- Check: `kubectl get sa kubernetes-cert-manager -n kubernetes-cert-manager -o yaml` (verify annotation)

**AWS**: "AccessDenied"
- IAM role missing Route53 permissions
- IRSA trust relationship broken
- Wrong zone ID or region
- Check: ServiceAccount annotation for role ARN

**Azure**: "AuthorizationFailed"
- Managed Identity missing DNS Zone Contributor role
- Workload identity federation misconfigured
- Check: ServiceAccount annotation for client ID

**Fix approach** (all providers):
1. Check kubernetes-cert-manager logs for specific error messages
2. Verify DNS provider credentials/permissions in their respective consoles
3. For cloud providers, test the ServiceAccount annotation manually
4. Regenerate credentials if needed and redeploy the addon

#### "No configured challenge solvers can be used"

**Symptom**: Certificate stuck in "Pending" with error about no matching solver

**Why**: The domain in your Certificate doesn't match any `dnsZones` in your `dnsProviders` configuration.

**Fix**: Either:
- Add the domain to a provider's `dnsZones` list in your KubernetesCertManager spec
- Create a new `dnsProvider` entry for that domain's DNS provider

Example:
```yaml
# If requesting cert for "newdomain.com" but it's not in any dnsZones:
dnsProviders:
  - name: cloudflare-new
    dnsZones:
      - newdomain.com  # Add it here
    cloudflare:
      apiToken: "..."
```

#### Rate Limit Hit

**Symptom**: "too many certificates already issued"

**Why**: Let's Encrypt limits you to 50 certificates per registered domain per week.

**Fix**:
- Use staging environment while testing: Update `spec.acme.server` to staging URL
- Consolidate domains using SANs instead of separate certificates
- Wait for the weekly limit to reset (rolling 7-day window)

## Common Patterns

### Multi-Domain Certificates with SANs

Get multiple domains in a single certificate using Subject Alternative Names (SANs):

```yaml
apiVersion: kubernetes-cert-manager.io/v1
kind: Certificate
metadata:
  name: multi-domain-cert
  namespace: default
spec:
  secretName: multi-domain-tls
  issuerRef:
    name: letsencrypt-cluster-issuer
    kind: ClusterIssuer
  dnsNames:
    - example.com
    - www.example.com
    - api.example.com
    - "*.staging.example.com"
```

This counts as one certificate against Let's Encrypt rate limits and simplifies cert management.

### Automatic Certificate Creation with Ingress

Cert-manager can watch Ingress resources and automatically create Certificates:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    kubernetes-cert-manager.io/cluster-issuer: letsencrypt-cluster-issuer  # Triggers auto-creation
spec:
  tls:
    - hosts:
        - app.example.com
      secretName: app-example-com-tls  # kubernetes-cert-manager creates this
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

No need to manually create the Certificate resource—kubernetes-cert-manager detects the annotation and handles it.

### Migrating DNS Providers

If changing DNS providers (e.g., GCP → Cloudflare):

1. **Move DNS first**: Update nameservers, wait for full propagation (24-48h)
2. **Update addon config**: Modify your KubernetesCertManager spec to add new provider
3. **Redeploy addon**: The ClusterIssuer is automatically updated with new solver
4. **Force certificate renewal**: Delete existing certificates to trigger reissuance with new provider
5. **Cleanup**: Remove old provider from `dnsProviders` list

The ClusterIssuer automatically updates when you change the addon configuration.

## Frequently Asked Questions

**Q: Do I still need to create ClusterIssuers manually?**

A: No! This addon automatically creates **one ClusterIssuer per domain**, each named after the domain itself (e.g., `planton.ai`, `planton.live`). Just reference the domain name in your Certificate resources.

**Q: Can I use multiple DNS providers for different domains?**

A: Yes! This is the primary use case. Configure multiple entries in the `dnsProviders` array, each with their own `dnsZones`. Each domain gets its own ClusterIssuer.

**Q: What if I need staging AND production issuers?**

A: Deploy two instances of the KubernetesCertManager addon with different names and different `acme.server` URLs. Each creates its own set of domain-named ClusterIssuers.

**Q: Can I use the same Cloudflare API token for multiple provider entries?**

A: Yes, but create separate provider entries if you want different `dnsZones` selectors. Each entry creates a separate solver in the ClusterIssuer.

**Q: How do I handle certificate renewal?**

A: kubernetes-cert-manager automatically renews certificates 30 days before expiration. You don't need to do anything. Monitor the kubernetes-cert-manager logs if you want to see renewal events.

**Q: Can I use HTTP-01 challenges instead of DNS-01?**

A: This addon is specifically designed for DNS-01 challenges and automatically configures the ClusterIssuer for DNS-01. For HTTP-01, you'd need to manually create a separate ClusterIssuer.

**Q: What happens if DNS provider credentials expire?**

A: Certificate renewals will fail. kubernetes-cert-manager will log errors. For Cloudflare, update the token in your KubernetesCertManager spec and redeploy. For cloud providers, fix the workload identity configuration.

## References

### Official Documentation
- [kubernetes-cert-manager Official Docs](https://kubernetes-cert-manager.io/docs/) - Comprehensive kubernetes-cert-manager documentation
- [Let's Encrypt](https://letsencrypt.org/docs/) - Free, automated certificate authority
- [ACME Protocol](https://datatracker.ietf.org/doc/html/rfc8555) - Standard for automated certificate management

### DNS Provider Guides
- [Cloudflare DNS-01 Solver](https://kubernetes-cert-manager.io/docs/configuration/acme/dns01/cloudflare/)
- [Google Cloud DNS Solver](https://kubernetes-cert-manager.io/docs/configuration/acme/dns01/google/)
- [AWS Route53 Solver](https://kubernetes-cert-manager.io/docs/configuration/acme/dns01/route53/)
- [Azure DNS Solver](https://kubernetes-cert-manager.io/docs/configuration/acme/dns01/azuredns/)

### API Token Setup
- [Cloudflare API Tokens](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/)
- [GCP Service Accounts](https://cloud.google.com/iam/docs/service-accounts)
- [AWS IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [Azure Workload Identity](https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview)

### Troubleshooting Resources
- [kubernetes-cert-manager Troubleshooting](https://kubernetes-cert-manager.io/docs/troubleshooting/) - Official troubleshooting guide
- [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/) - Understand production limits
- [DNS-01 Challenge Types](https://letsencrypt.org/docs/challenge-types/) - When to use DNS-01 vs HTTP-01
