# GitLab Connection

## Use Case

Connect Cloud Build to gitlab.com or a GitLab Enterprise server with a read-write and a read-only personal access token.

## When to Use

- Your code lives on GitLab
- You want Cloud Build to install webhooks on your GitLab projects

## What This Creates

- The Cloud Build API on the project
- A connection to the GitLab server, with both tokens and the webhook secret named from Secret Manager

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `gitlabConfig.hostUri` | `https://gitlab.example.com` | Your GitLab Enterprise server; remove it for gitlab.com. |
| `gitlabConfig.authorizerCredential` | `gitlab-api-token` reference | A token with the `api` scope. |
| `gitlabConfig.readAuthorizerCredential` | `gitlab-read-token` reference | A token with at least `read_api`. |
| `gitlabConfig.webhookSecretSecretVersion` | `gitlab-webhook-secret` reference | Immutable: changing it replaces the connection. |
