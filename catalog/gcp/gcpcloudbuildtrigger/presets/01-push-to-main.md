# Push to Main

## Use Case

Build every push to a repository's main branch with the `cloudbuild.yaml` in that repository, running as a dedicated service account and skipping documentation-only changes.

## When to Use

- Continuous integration on a repository linked through a Cloud Build connection
- Teams that keep their build definition next to their code

## What This Creates

- The Cloud Build API on the project
- A regional trigger on pushes to `main` of the linked repository

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region of the repository's connection. |
| `repositoryEventConfig.push.branch` | `^main$` | Another branch, or `tag: ^v.*` to build releases. |
| `ignoredFiles` | docs and Markdown | Paths whose changes should not start a build. |
| `filename` | `cloudbuild.yaml` | Where the build file lives in the repository. |
| `serviceAccount` | a `GcpServiceAccount` reference | The identity builds run as; its `cloudbuild.yaml` must set a log destination. |
