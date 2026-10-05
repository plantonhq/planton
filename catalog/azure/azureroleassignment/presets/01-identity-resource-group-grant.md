# Managed Identity Grant on a Resource Group

This preset grants a user-assigned managed identity a built-in role on a
resource group -- the most common authorization pattern in composed Azure
environments: the identity a workload runs as gets exactly the access it needs
on the environment it lives in.

The template uses literal placeholders so it deploys standalone. In an infra
chart, replace either literal with a `valueFrom` block to wire the grant to
deployed resources -- both fields resolve through their default kinds, so only
the resource names are needed:

```yaml
spec:
  scope:
    valueFrom:
      name: platform-rg          # resolves to the AzureResourceGroup's ARM ID
  principalId:
    valueFrom:
      name: app-identity         # resolves to the AzureUserAssignedIdentity's principal ID
```

`skipServicePrincipalAadCheck` is pre-set for the composed case where the
identity is created in the same deployment (freshly created principals
replicate through Azure AD asynchronously; the flag avoids minutes of
PrincipalNotFound retries).

## When to Use

- Granting a workload identity access to the resource group its dependencies live in
- Granting a CI/CD deploy identity scoped rights on an environment
- Any identity + grant pair created together in one Infra Chart

## Key Configuration Choices

- **Role** (`roleDefinitionName`) -- prefer the narrowest built-in role that
  satisfies the need: "Reader" for observation, "Contributor" for resource
  management (never grants RBAC rights), or a service-specific data-plane role
- **Scope** -- a resource group inherits the grant to everything inside it;
  switch the reference to a single resource (explicit `kind` + `fieldPath`)
  for tighter scoping

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-arm-id>` | ARM ID of the resource group being granted on (`/subscriptions/{sub}/resourceGroups/{name}`) | `az group show --name <rg> --query id` |
| `<built-in-role-name>` | Built-in role, e.g. `Reader`, `Contributor`, `AcrPull` | [Azure built-in roles](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles) |
| `<identity-principal-object-id>` | The managed identity's principal (object) ID -- not its client ID | `az identity show --name <mi> --resource-group <rg> --query principalId` |
| `<why-this-grant-exists>` | Audit note shown in the portal's IAM blade | Your runbook / change ticket |
