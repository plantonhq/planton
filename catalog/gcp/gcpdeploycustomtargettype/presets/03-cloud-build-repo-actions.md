# Cloud Build Repository Actions

## Use Case

Keep Skaffold custom actions in a private repository already linked to Cloud Build, so Cloud Deploy reads them through the Cloud Build connection instead of a public clone URL.

## When to Use

- Deploy actions in a private GitHub, GitLab, or Bitbucket repository
- Organizations that already link repositories through Cloud Build connections

## What This Creates

- The Cloud Deploy API on the project
- A custom target type that deploys with `platform-deploy` from the Skaffold file in a linked Cloud Build repository; Cloud Deploy renders as usual

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `customActions.deployAction` | `platform-deploy` | The name of your deploy action. |
| `customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository` | a `GcpCloudBuildRepository` reference | Your linked repository, or its full name as a literal. |
| `customActions.includeSkaffoldModules[].googleCloudBuildRepo.path` / `ref` | `platform/skaffold.yaml` / `main` | Where the Skaffold file lives and which branch or tag to read. |
