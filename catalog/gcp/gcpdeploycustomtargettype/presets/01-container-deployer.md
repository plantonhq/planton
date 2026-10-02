# Container Deployer

## Use Case

Deploy releases to a system Cloud Deploy does not support natively by running your own deployer container, with Cloud Deploy's normal rendering.

## When to Use

- Deploying to a vendor API, an internal platform, or a non-Kubernetes runtime
- Teams that want a custom target without learning Skaffold custom actions

## What This Creates

- The Cloud Deploy API on the project
- A custom target type whose deploy task runs the deployer image with one argument and one environment variable

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region of the targets that use the type. |
| `tasks.deploy.container.image` | a placeholder image | Your deployer image; the targets' execution service account must be able to pull it. |
| `tasks.deploy.container.args` / `env` | `--wait`, `VENDOR_REGION` | What your deployer needs. |
| `tasks.render` | unset | Add a render container to render releases yourself. |
