# Pull Request Checks

## Use Case

Run checks on every pull request opened against main, show the build log on the pull request, and make outside contributors wait for a maintainer's `/gcbrun` comment before their code runs.

## When to Use

- Required status checks on a GitHub repository linked through a Cloud Build connection
- Open-source repositories that accept pull requests from forks

## What This Creates

- The Cloud Build API on the project
- A regional trigger on pull requests against `main` of the linked repository

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `repositoryEventConfig.pullRequest.branch` | `^main$` | The base branches whose pull requests are checked. |
| `repositoryEventConfig.pullRequest.commentControl` | external contributors only | `COMMENTS_DISABLED` to build every pull request, `COMMENTS_ENABLED` to always wait for `/gcbrun`. |
| `includeBuildLogs` | with status | Remove for GitLab or Bitbucket repositories (Google accepts it only on GitHub). |
| `filename` | `cloudbuild.pr.yaml` | The check build's file. |
