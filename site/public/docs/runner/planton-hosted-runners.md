---
title: "Planton-Hosted Runners"
description: "On Planton's hosted product, your deploys, live cloud operations, and builds run on runners Planton operates until you add a runner of your own"
icon: server
order: 20
tags:
  - Runner
  - Security
  - Deployments
  - Builds
---

# Planton-Hosted Runners

On Planton's hosted product, your work runs on runners Planton operates until you add a runner of your own. A new organization can connect a cloud account, deploy, and build without installing anything: a Planton-hosted runner picks up the work.

## Why They Exist

A runner is what executes your work: the infrastructure-as-code behind a deploy, the live calls behind a connection's verify or `planton kubectl`, the container build behind a service. Asking you to install one before your first deploy puts setup between you and the result. Planton-hosted runners remove that step. You add a runner of your own when you need what only a runner inside your environment can give you, not before.

## What a Planton-Hosted Runner Can Access

A Planton-hosted runner holds no standing access to your organization. For each job it runs, it exchanges its own identity for a short-lived credential scoped to that one job and your organization. That credential can read only what the job needs: the secrets and variables the job references and the connections it deploys through. It cannot change anything in your organization. It lasts an hour at most, and a new one is issued only while the job is still running.

The job's cloud credentials come from the connection you configured, exactly as they would on a runner of your own. Planton-managed state storage is available only to Planton-hosted runners (see [State Backends](/docs/connections/state-backends)).

## Where Your Work Runs

Each kind of work picks its runner in its own order. The first step that names a runner wins.

| Work | 1 | 2 | 3 |
|------|---|---|---|
| **Deploys** (deploy, update, destroy) | The runner the connection names | A Planton-hosted runner | — |
| **Live cloud operations** (verify, `planton kubectl`, resource browsing) | The runner the connection names | Your organization's default runner | A Planton-hosted runner |
| **Builds** | The build connection the service names | Your organization's default build connection | A Planton-hosted runner |

Your organization's default runner carries live cloud operations only. A deploy does not consult it: to run deploys on your own runner, name that runner on the connection.

## When to Add a Runner of Your Own

Add a runner inside your environment when:

- **Your resources are private.** A Kubernetes API server, a database, or a Vault that is reachable only from inside your network needs a runner inside that network.
- **Your credentials must stay in your cloud.** With the Self-Hosted Runner connection mode, the runner discovers credentials from its own environment (an IAM role, a workload identity), and nothing is stored in Planton.
- **Your policies require it.** Some compliance regimes require every change to originate from infrastructure you operate.

See [Deployment](/docs/runner/deployment) to start one.

## Running Deploys on Your Own Runner

1. Start a runner with a runner token (see [Deployment](/docs/runner/deployment)).
2. Keep your organization's state in a backend your runner can reach: S3, GCS, Azure Blob, Cloudflare R2, Terraform Cloud, or Pulumi Cloud. Planton-managed state is available only to Planton-hosted runners, so a deploy on your own runner needs a backend of your own.
3. Move each existing resource's state onto your backend:

   ```bash
   planton tofu state migrate-backend AwsS3Bucket my-data-bucket \
     --destination-state-backend my-s3-backend
   ```

   Use `planton pulumi state migrate-backend` for resources deployed with Pulumi. Add `--dry-run` to preview a migration.
4. Name the runner on the connection. Every deploy and live cloud operation through that connection runs on it from then on.

## Choosing a Default Runner

Your organization's default runner serves live cloud operations for every connection that names no runner:

```bash
# Show where live cloud operations run for your organization
planton runner get-default

# Choose one of your organization's runners as the default
planton runner set-default prod-runner

# Clear it; live cloud operations return to a Planton-hosted runner
planton runner unset-default
```

## Related Documentation

- [Runner Overview](/docs/runner) — What a runner is and how it connects
- [Deployment](/docs/runner/deployment) — Starting and deploying a runner of your own
- [Security Model](/docs/runner/security-model) — Credential isolation and trust boundaries
- [State Backends](/docs/connections/state-backends) — Where deployment state lives
