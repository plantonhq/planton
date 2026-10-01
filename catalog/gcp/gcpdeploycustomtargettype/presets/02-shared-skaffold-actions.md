# Shared Skaffold Actions

## Use Case

Render and deploy through Skaffold custom actions kept in one shared Git repository, so every application that deploys this way uses the same, versioned actions.

## When to Use

- Teams already standardized on Skaffold
- A platform team publishing deploy actions for many applications

## What This Creates

- The Cloud Deploy API on the project
- A custom target type that renders with `terraform-plan` and deploys with `terraform-apply`, both from the `terraform-actions` config in a shared repository at tag `v1`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `customActions.renderAction` / `deployAction` | `terraform-plan` / `terraform-apply` | The names of your actions. |
| `customActions.includeSkaffoldModules[].git` | a placeholder repository | Your actions repository, path, and tag; or use `googleCloudBuildRepo` with a `GcpCloudBuildRepository` reference, or `googleCloudStorage` with a `gs://` path. |
| `customActions.includeSkaffoldModules[].configs` | `terraform-actions` | The config names in that Skaffold file; empty takes them all. |
