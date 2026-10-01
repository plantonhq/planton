# CI Deployer Identity

This preset creates the identity at the center of keyless CI/CD -- the
identity a pipeline authenticates as when it deploys to Azure. It is the
first of three composable pieces; the other two attach to its outputs:

**1. What CI may do** -- an `AzureRoleAssignment` granting the identity its
deployment rights:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleAssignment
metadata:
  name: ci-deployer-contributor
spec:
  scope:
    valueFrom:
      name: platform-rg
  roleDefinitionName: Contributor
  principalId:
    valueFrom:
      name: ci-deployer-identity
```

**2. Who may act as it** -- an `AzureFederatedIdentityCredential` trusting
the pipeline's OIDC token:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFederatedIdentityCredential
metadata:
  name: ci-deployer-trust
spec:
  name: github-main-branch
  userAssignedIdentity:
    valueFrom:
      name: ci-deployer-identity
  issuer: https://token.actions.githubusercontent.com
  subject: repo:<owner>/<repo>:ref:refs/heads/main
```

Replace `<owner>/<repo>` in `subject` with the repository allowed to deploy.

No client secret exists at any point in this composition.

## When to Use

- Replacing a stored service-principal secret in GitHub Actions, GitLab CI,
  or any OIDC-issuing pipeline
- Establishing per-pipeline identities whose rights and trusted sources are
  reviewable as code

## Key Configuration Choices

- **One identity per pipeline duty** -- a deploy identity, a
  registry-push identity, etc., each with narrow grants, beats one broad
  shared identity
- **Scope grants to what CI touches** -- a resource group by reference, not
  the subscription

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<org-prefix>` | Your organization's naming prefix | Your naming convention |
