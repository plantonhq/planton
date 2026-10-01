---
title: "Trusted Workflows"
description: "Let a GitHub repository's workflows register and deploy a service with no stored key: trust exactly one repository, copy the workflow the console fills in, and see every run the trust admitted or refused."
icon: shield
order: 37
tags: [CI/CD, GitHub Actions, Keyless, Security]
---

# Trusted Workflows

A trusted workflow is a GitHub repository whose Actions workflows may ship into your Planton organization without a stored key. Each run proves itself with the token GitHub signs for that run, and Planton exchanges that token for a short-lived credential that can register and deploy one service: the one the repository builds.

## Why No Stored Key

The usual way to deploy from CI is an API key saved as a repository secret. A key like that works from anywhere, never expires on its own, and leaks with the first copied workflow or compromised fork. A trusted workflow has nothing to leak:

- GitHub mints a fresh token for every run, naming the repository, branch or tag, and environment it ran for.
- Planton accepts the token only when a trust names exactly that repository and the token was requested for this Planton, its **audience**. The same workflow copied into another repository is refused.
- The credential Planton hands back expires in minutes, and can only register and deploy the service whose declared repository matches the token. It can never approve anything.

## Trust a Repository

Trusting a repository is an access decision, so it's open to the people who manage access in the organization. Everyone else sees who can.

1. Open **Organization Settings → Trusted Workflows** and choose **Trust a Repository**. The service page's **Ship from GitHub Actions** panel and the register dialog's CI door open the same flow.
2. Name the **Repository** as GitHub writes it, `owner/repository`. Only workflows in exactly that repository are trusted.
3. Optionally narrow the trust:
   - **Branch or Tag**: **One Branch** or **One Tag**, and runs from anywhere else in the repository are refused.
   - **GitHub Environment**: only jobs that target that environment are trusted. GitHub can make those jobs wait for a reviewer.
   - **Audience**, under **Advanced**: filled in with this Planton's own address. Change it only if you know why.
4. Pick the service the repository builds and the environment it deploys to. When no service builds from the repository yet, the workflow can register one from a `service.yaml` you write, or you register the service in the console first.
5. Copy the workflow into `.github/workflows/planton-deploy.yml` and push. The dialog waits for the first run and says when it lands.

### The Account a Trust Acts As

Every trust acts as a service account, which is the name the audit trail shows. Name none, the usual choice, and Planton provisions one for the repository, `ci-<owner>-<repo>`. Every trust of that repository shares it, and it's removed with the last trust that uses it. A trust can also act as a service account you made yourself. That account is never removed with the trust.

The account grants the workflow nothing beyond shipping its own service. The exchanged credential carries a fixed set of rules, whatever the account itself could do.

## The Workflow

The console writes the whole file with every value filled in. The repository needs no Actions secret and no variable: GitHub mints the only credentials involved, `GITHUB_TOKEN` to push the image to GHCR and the OIDC token Planton exchanges. With example values, it looks like this:

```yaml
name: planton-deploy
on:
  push:
    branches:
      - 'main'
permissions:
  contents: read
  id-token: write
  packages: write
jobs:
  ship:
    runs-on: ubuntu-latest
    env:
      PLANTON_API_ENDPOINT: planton.acme.example:443
    steps:
      - uses: actions/checkout@v4
      - name: Build and push the image to GHCR
        run: |
          echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u "${{ github.actor }}" --password-stdin
          docker build --label "org.opencontainers.image.source=${{ github.server_url }}/${{ github.repository }}" -t "ghcr.io/${{ github.repository }}:${{ github.sha }}" .
          docker push "ghcr.io/${{ github.repository }}:${{ github.sha }}"
      - name: Register and deploy with Planton
        uses: plantonhq/planton/actions/deploy@main
        with:
          org: acme
          audience: https://planton.acme.example
          service: checkout-api
          environment: prod
          image: ghcr.io/${{ github.repository }}:${{ github.sha }}
```

- **`id-token: write`** lets the action ask GitHub for the token it exchanges. **`packages: write`** lets `GITHUB_TOKEN` push the image.
- **The trigger follows the trust.** A trust narrowed to one branch gets a workflow that runs on pushes to that branch, and a tag trust gets one that runs on that tag, so the workflow never runs where its trust would refuse it. A trust open to any branch or tag runs on the service's default branch, or on every push when Planton doesn't know it.
- **The image path names no repository.** It's built from GitHub's expressions, and the source label links the package to the repository, so it inherits the repository's access.
- **Registering first.** When the workflow registers the service from a manifest, the step also carries `register: 'true'` and `service-file`. The manifest's repository is proven against the run's token, never taken as typed.

The deploy step waits for the run and fails the job unless the deployment verifiably came online. The [Deploy from GitHub Actions](/docs/ci-cd/deploy-from-github-actions) page and the [action's README](https://github.com/plantonhq/planton/tree/main/actions/deploy) cover every input.

## Pause and Remove

A trust's page shows the runs it started and what each shipped, when a workflow last used it, and its conditions. Two actions stop a trust, and both take effect at once:

- **Pause Trust** keeps the trust and refuses its runs until someone chooses **Resume Trust**. The pause asks why, and whoever resumes it reads the reason on the trust's page.
- **Remove Trust** deletes it. A provisioned service account goes with the last trust that uses it.

Either way, the credentials the trust already handed out stop working too, so a run in progress stops at its next call.

The repository a trust names never changes. To trust another repository, trust it separately.

## Refused Runs

A workflow Planton refuses gets one sentence, whatever the cause:

```
the presented token was not accepted by any workload identity binding in this organization
```

That's deliberate: the answer never tells a stranger's workflow whether the organization, a trust, or the repository exists. The real cause is recorded for the organization, one record per repository and reason, with how many runs it covers, the most recent run and its link, and what the run presented beside what the nearest trust expects:

| Cause | What it means |
|---|---|
| No trust names the repository | No trust in this organization names the repository, so its runs are refused. |
| Another branch or tag | A trust names the repository, but only for another branch or tag. |
| Another environment | A trust names the repository, but only for another GitHub environment. |
| Another audience | The run asked for a token meant for another address than the trust expects. |
| Trust paused | The trust that would admit the run is paused. |
| Service account missing | The trust matched, but the service account it acts as no longer exists. |

Only the people who manage access in the organization can read refusals, the same people who manage the trusts they explain. They read them:

- **in the console**: on **Trusted Workflows** and on each trust's page, each refusal with the change that fixes it. Runs refused from repositories no trust names are listed apart, with no one-click trust, because anyone can aim a signed token at any organization.
- **from the CLI**: `planton iam refusal list` (add `--trust <slug>` for one trust) and `planton iam refusal get <id>`.
- **through the Planton Assistant**, which reads the same records and can walk you to the fix.

## Where Trusted Workflows Work

- **GitHub's runners must reach your Planton.** The workflow exchanges its token at your Planton's public gRPC address. On hosted Planton that address is always published. A self-hosted install that publishes none says so on **Trusted Workflows**, and whoever runs the install can publish one.
- **Not on a laptop.** No CI runner can reach a Planton running on your own machine, so the local instance doesn't offer trusted workflows.
- **GitLab.** Trusts for gitlab.com projects are made and managed from the CLI.
