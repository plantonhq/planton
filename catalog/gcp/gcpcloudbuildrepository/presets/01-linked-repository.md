# Linked Repository

## Use Case

Link one repository through an existing connection so Cloud Build triggers can build on its pushes and pull requests.

## When to Use

- A team onboarding its repository to Cloud Build
- Before declaring a trigger on `repositoryEventConfig`

## What This Creates

- A repository link under the `acme-github` connection, in the connection's project and region

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `parentConnection` | `acme-github` reference | The connection your repository is reachable through. |
| `remoteUri` | `https://github.com/acme/orders.git` | Your repository's HTTPS clone URI. |
| `repositoryId` | `metadata.name` | A different ID in Cloud Build. |
