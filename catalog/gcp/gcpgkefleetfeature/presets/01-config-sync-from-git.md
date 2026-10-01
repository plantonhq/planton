# Config Sync from Git

## Use Case

Have every cluster in the fleet sync its Kubernetes configuration from one Git repository, with Google keeping Config Sync upgraded.

## When to Use

- GitOps for platform configuration (namespaces, RBAC, quotas, policies) across all clusters
- Clusters that join the fleet later should pick up the same configuration automatically

## What This Creates

- The Fleet API and the Config Management API on the fleet host project
- The configmanagement feature, with Config Sync as the fleet-wide default

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `fleetDefaultMemberConfig.configmanagement.configSync.git.syncRepo` | example repository | Your configuration repository. |
| `fleetDefaultMemberConfig.configmanagement.configSync.git.secretType` | `none` | `gcpserviceaccount` (with `gcpServiceAccountEmail`), `ssh`, `token`, or `githubapp` for a private repository. |
| `fleetDefaultMemberConfig.configmanagement.configSync.git.policyDir` | `clusters/all` | The directory to sync. |
