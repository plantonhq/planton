# Cert Manager Issuer

Creates one cert-manager Issuer — a NAMESPACE-SCOPED certificate authority front-end. Only Certificate resources in the same namespace can request certificates from it, and every Secret it needs (credentials, CA keypairs) lives in that same namespace. The namespace scope is the point: a team's CA keypair and DNS credentials stay readable only inside the team's namespace instead of being trusted cluster-wide. The same four signing backends as the cluster-scoped variant: ACME, CA, SelfSigned, and Vault.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Issuer** -- the cert-manager custom resource, named after `metadata.name`, in the specified namespace, configured with the selected signing backend
- **Credential Secrets** -- wherever the backend needs a credential (a DNS token, Vault token, TSIG key), the module materializes it as a Kubernetes Secret in the Issuer's own namespace and wires the CR's secretRef to it
- **ACME Account Key Secret** -- created by cert-manager itself on first ACME registration (ACME backend only)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **Cert Manager must be installed first** -- the Issuer is a cert-manager custom resource.
- The target namespace must exist (or be managed by a Kubernetes Namespace resource this one references).
- For the CA backend: a Secret holding the CA keypair (`tls.crt` + `tls.key`) in the SAME namespace — typically the output of a Cert Manager Certificate with `isCa: true`.

## Deploy

### Console

Open the deployment store, find **Cert Manager Issuer**, and click **Deploy**. The creation wizard walks you through the namespace, the signing backend, and — for ACME — the challenge solvers. Start from the **Self-Signed Issuer** preset to bootstrap internal PKI in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIssuer
metadata:
  name: selfsigned-bootstrap
  org: acme-corp
  env: prod
spec:
  namespace:
    value: team-payments
  config:
    selfSigned: {}
```

```shell
planton apply -f issuer.yaml
```

This creates a self-signed Issuer in `team-payments` — the starting point of the standard CA-chain bootstrap. `selfSigned: {}` is a complete, meaningful configuration: presence selects the backend. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the CA bootstrap by reference — the CA-backend Issuer consumes the CA Certificate's Secret output:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: payments-namespace
      fieldPath: spec.name
  config:
    ca:
      caSecretName:
        valueFrom:
          kind: KubernetesCertificate
          name: payments-root-ca
          fieldPath: status.outputs.secret_name
```

The InfraPipeline deploys the namespace and the CA Certificate first, then creates the Issuer against them.

## Key Configuration

These are the most important decisions when configuring an Issuer. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Issuer vs ClusterIssuer** -- Identical signing capabilities (the config message is shared), different scope. Use the namespace-scoped Issuer when a namespace needs its own CA or its own credential blast radius; use a Cert Manager Cluster Issuer for the platform-wide CA serving every team.

**The CA bootstrap is a four-step composition** -- (1) a self-signed Issuer, (2) a Cert Manager Certificate with `isCa: true` referencing it — producing a root CA Secret, (3) a CA-backend Issuer pointing `config.ca.caSecretName` at that Secret's output, (4) leaf Certificates referencing the CA Issuer. Each step references the previous one's outputs — the standard mTLS bootstrap, entirely FK-wired.

**Everything stays in the namespace** -- Credential Secrets, the CA keypair, the ACME account key, and every Certificate that uses this Issuer live in the Issuer's namespace. If a certificate is needed in another namespace, that namespace needs its own Issuer (or a ClusterIssuer).

**Team-owned ACME** -- An ACME backend on a namespace Issuer keeps DNS credentials readable only inside the team's namespace. The same solver surface as the ClusterIssuer applies (DNS-01 for wildcards and non-public services, HTTP-01 for public reachability).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesCertificate** | `config.ca.caSecretName` | `status.outputs.secret_name` |
| **KubernetesServiceAccount** | `config.vault.kubernetesAuth.serviceAccountName` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `issuer_name` | Name of the created Issuer (equals `metadata.name`) | Cert Manager Certificate `issuerRef.issuer.name` (same namespace only) |
| `namespace` | Namespace where the Issuer was created | Verifying Certificate co-location |
| `acme_account_key_secret_name` | The ACME account private key Secret cert-manager creates in the Issuer's namespace (empty for non-ACME backends) | Migrating the ACME account |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bootstrap a root CA** -- A self-signed Issuer that a CA Certificate (`isCa: true`) references. Start from the **Self-Signed Issuer** preset.

**Sign service certificates** -- A CA-backend Issuer referencing the root CA Secret, issuing mTLS leaf certificates inside the namespace. Start from the **CA Issuer** preset.

## Works With

- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- must be installed first; provides the controller and CRDs.
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- requests certificates from this Issuer (same namespace); also produces the CA Secret a CA-backend Issuer consumes.
- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) -- the cluster-scoped alternative for platform-wide public TLS.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference it so infra charts create the namespace and this Issuer in dependency order.
