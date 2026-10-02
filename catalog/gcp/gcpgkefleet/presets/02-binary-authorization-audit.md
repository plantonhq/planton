# Binary Authorization Audit

## Use Case

Audit every workload in the fleet's clusters against a GKE platform policy, so non-compliant images show up in the fleet's security findings.

## When to Use

- A fleet whose clusters must run only images that meet an organization policy
- Rolling out continuous validation before enforcing at admission

## What This Creates

- The Fleet API on the project
- The project's fleet, evaluating workloads against one platform policy by default

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `defaultClusterConfig.binaryAuthorizationConfig.policyBindings` | one placeholder policy | Your platform policy names, `projects/{number}/platforms/gke/policies/{id}`. |
| `projectId` | `my-gcp-project` | The fleet host project. |
