# GitHub Connection

## Use Case

Connect Cloud Build to a GitHub organization through Cloud Build's GitHub App, so triggers can build on pushes and pull requests.

## When to Use

- Your code lives on github.com
- You want pull-request and push triggers through Cloud Build's second-generation repositories

## What This Creates

- The Cloud Build API on the project
- A connection to the GitHub installation, authorized by an OAuth token kept in Secret Manager

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region your repositories and triggers use. |
| `githubConfig.appInstallationId` | `12345678` | Your installation of Cloud Build's GitHub App (in the app's settings URL). |
| `githubConfig.authorizerCredential.oauthTokenSecretVersion` | `github-oauth-token` reference | The secret holding the authorizing account's OAuth token; grant Cloud Build's service agent `secretAccessor` on it. |
