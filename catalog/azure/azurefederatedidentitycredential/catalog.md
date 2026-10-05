# Azure Federated Identity Credential

Deploys a federated identity credential: a keyless trust rule on a user-assigned managed identity that lets an external workload exchange its own OIDC token for Azure credentials — no client secret, nothing to rotate, nothing to leak. It unlocks two headline flows: CI pipelines (GitHub Actions and friends) that deploy to Azure without stored service-principal secrets, and AKS workload identity, where pods reach RBAC-granted resources through a Kubernetes service account's projected token.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Federated Identity Credential** -- one trust rule written on the parent user-assigned managed identity, visible in the portal's "Federated credentials" tab

The trust model is a three-way match. An external workload presents a token; Azure AD accepts the exchange only when the token's `iss` claim matches this credential's issuer, its `sub` claim matches the subject, and its `aud` claim matches the audience -- each byte for byte, no wildcards. An identity carries one credential per external workload that should act as it (Azure allows up to 20 per identity).

A credential conveys no permissions by itself -- it only authenticates the external workload AS the identity. What that identity may do is granted separately through **AzureRoleAssignment** resources.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A user-assigned managed identity** the credential is written on. Reference an AzureUserAssignedIdentity Infra Component via ValueFromRef, or pass the identity's full ARM resource ID as a literal.

## Deploy

### Console

Open the deployment store, find **Azure Federated Identity Credential**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields -- with quick-picks for the well-known issuers and subject-format templates so the byte-for-byte match rules are hard to get wrong. Start from the **GitHub Actions Keyless CI** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFederatedIdentityCredential
metadata:
  name: github-main-branch
  org: acme-corp
  env: prod
spec:
  name: github-main-branch
  userAssignedIdentity:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: ci-deployer
      fieldPath: status.outputs.identity_id
  issuer:
    value: "https://token.actions.githubusercontent.com"
  subject: "repo:acme/platform:ref:refs/heads/main"
```

```shell
planton apply -f credential.yaml
```

This lets workflows on the `main` branch of `acme/platform` deploy to Azure as the `ci-deployer` identity -- with no stored service-principal secret anywhere. The audience is omitted, so Azure applies `api://AzureADTokenExchange` (what every standard client requests). An Infra Job tracks the provisioning in real time.

### InfraChart

The AKS workload-identity composition references the cluster's issuer -- it only exists once the cluster is deployed:

```yaml
spec:
  name: aks-payments-serviceaccount
  userAssignedIdentity:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: payments-api
      fieldPath: status.outputs.identity_id
  issuer:
    valueFrom:
      kind: AzureAksCluster
      name: prod-aks
      fieldPath: status.outputs.oidc_issuer_url
  subject: "system:serviceaccount:payments:payments-api"
```

The InfraPipeline resolves the dependency graph, deploys the identity and cluster first, then provisions the credential with the resolved issuer URL.

## Key Configuration

These are the most important decisions when configuring a federated identity credential. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The byte-for-byte match** -- Azure applies no normalization: no trailing-slash forgiveness on the issuer, no wildcards in the subject. The classic failure is a near-miss string; decode a real token's claims when unsure.

**Subject specificity** -- The trust is exactly as narrow as the subject string. Prefer the most specific form the issuer offers: a branch-scoped GitHub subject (`repo:{owner}/{repo}:ref:refs/heads/{branch}`) over a repository-wide one, one Kubernetes service account (`system:serviceaccount:{namespace}:{serviceaccount}`) over a namespace. Need to trust several branches or accounts? Add one credential each -- rules stay individually reviewable and removable.

**Audience** -- Leave it unset for `api://AzureADTokenExchange`, the audience Azure AD's token-exchange endpoint expects and the value every standard client (azure-sdk, azure/login, the AKS workload-identity webhook) requests. Override only when a provider genuinely mints a different audience.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureUserAssignedIdentity** | `userAssignedIdentity` | `status.outputs.identity_id` |
| **AzureAksCluster** (workload identity) | `issuer` | `status.outputs.oidc_issuer_url` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream tooling can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `federated_identity_credential_id` | Full ARM ID of the credential | Automation and auditing |
| `name` | The credential's ARM resource name | Portal cross-reference |
| `user_assigned_identity_id` | ARM ID of the parent identity | Wiring the external side to the right identity |
| `issuer` | The issuer as deployed | CI configuration generators, exchange-failure diagnosis |
| `subject` | The subject as deployed | CI configuration generators, exchange-failure diagnosis |
| `audience` | The audience as deployed | Non-standard client configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI without secrets** -- One credential per pipeline trust boundary (branch, environment, or tag), all on a `ci-deployer` identity whose role assignments scope to exactly what the pipeline deploys. Start from the **GitHub Actions Keyless CI** preset.

**AKS workload identity** -- The issuer references the cluster's `oidc_issuer_url` output; the subject names one Kubernetes service account; the pod's service account carries the `azure.workload.identity/client-id` annotation pointing at the identity's client ID. Start from the **AKS Workload Identity** preset.

**Any other OIDC issuer** -- The same trust rule works for GitLab, HashiCorp Vault, or any provider that mints OIDC tokens: a literal issuer URL and the provider's own subject format. Start from the **Generic External OIDC Issuer** preset.

## Works With

- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the parent identity this trust rule is written on
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants the identity the permissions the trusted workload will exercise
- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- provides the OIDC issuer for the workload-identity composition
