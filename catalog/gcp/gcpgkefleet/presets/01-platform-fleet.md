# Platform Fleet

## Use Case

Create the fleet a platform team runs its clusters in, before any cluster joins, with Google's standard security posture checks on every cluster by default.

## When to Use

- Setting up fleet team management (scopes) before clusters exist
- Giving every cluster that joins the fleet configuration auditing and OS vulnerability scanning

## What This Creates

- The Fleet API on the project
- The project's fleet, with basic security posture and vulnerability scanning as the cluster defaults

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId` | `my-gcp-project` | The fleet host project (or a `GcpProject` reference). |
| `displayName` | `Platform fleet` | The name teams see in the console (4-30 characters). |
| `defaultClusterConfig.securityPostureConfig.vulnerabilityMode` | `VULNERABILITY_BASIC` | `VULNERABILITY_ENTERPRISE` adds language-package scanning. |
