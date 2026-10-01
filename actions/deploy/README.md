# Planton Deploy Action

Deploy from GitHub Actions through Planton — in either of two modes, inferred from the inputs, with **zero stored Planton secrets** in both:

- **Offline** — no Planton backend anywhere. The action renders the checked-out repository's own kustomize declaration and deploys it to **your** cloud through the open-source engine, in dependency order, behind a preflight report that verifies everything verifiable before the first IaC handoff. Your runner authenticates to your cloud with the provider's own official OIDC action — keyless end to end.
- **Connected** — deploy through your Planton backend. The action exchanges the workflow's own OIDC token for a short-lived Planton credential, optionally registers the service from its manifest, deploys your container image into one environment, and fails the job unless the deployment verifiably succeeded.

Set `org` and `audience` and you are connected; leave both out and you are offline. Any half-state fails immediately with the exact line to add or remove.

## Switching modes is a few lines, both directions

The same action serves a repo through its whole journey — start offline with nothing but a cloud account, connect to a backend later (or the reverse) without changing anything else about the workflow:

| Direction | Change in the `with:` block |
|---|---|
| **Offline → connected** | Add `org`, `audience`, and `service`. Drop the `PLANTON_BACKEND_*` state env — the backend holds state. `image` stays exactly as it is. |
| **Connected → offline** | Remove `org`, `audience`, and `service`, and give the job a remote state backend (`PLANTON_BACKEND_*` env or manifest annotations). `image` stays exactly as it is. |

Everything else — the checkout, the image build, the provider OIDC auth your job already does — stays exactly as it is.

## Quick start — offline (no Planton backend)

The repository carries its own deployment declaration (a `_kustomize/` tree with one overlay per environment). The job authenticates to the cloud with the provider's official OIDC action, and state lives in your own bucket — nothing here stores a long-lived key:

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write   # lets the provider's auth action mint an OIDC token
      contents: read
    steps:
      - uses: actions/checkout@v4

      # Your cloud, your trust rule: the provider's own official OIDC action.
      # (aws-actions/configure-aws-credentials and azure/login work the same way.)
      - uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: projects/000000/locations/global/workloadIdentityPools/github/providers/github
          service_account: deployer@my-project.iam.gserviceaccount.com

      # ... build and push your image ...

      - uses: plantonhq/planton/actions/deploy@main
        env:
          # CI runners are ephemeral — state must outlive them. Point the
          # engine at your own bucket (or carry these as annotations in the
          # manifests themselves).
          PLANTON_BACKEND_TYPE: gcs
          PLANTON_BACKEND_BUCKET: my-project-tofu-state
        with:
          environment: prod
          image: ghcr.io/acme/storefront@${{ steps.build.outputs.digest }}
```

The job's log shows one preflight report — schema, references, state backend reachability, credential validity, module availability, all checked before anything is handed to an IaC engine — then the resources deploy in the order their own references require, and the job's exit code tells the truth: refused at preflight (nothing ran) and a mid-deploy failure (re-run continues as no-ops) are distinct.

## Quick start — connected (through your Planton backend)

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write   # lets the job mint its OIDC token
      contents: read
    steps:
      - uses: actions/checkout@v4

      # ... build and push your image ...

      - uses: plantonhq/planton/actions/deploy@main
        with:
          org: acme
          service: checkout-api
          environment: prod
          image: ghcr.io/acme/checkout-api@${{ steps.build.outputs.digest }}
          audience: https://acme.example/planton   # must match your trust's audience
```

### One-time setup for connected mode (keyless trust)

The keyless exchange works because your organization **trusts the repository**: "workflows from this repository may ship into this organization". Someone who manages access in the organization sets it up once, and the console hands back the finished workflow.

**In the console** — **Organization Settings → Trusted Workflows → Trust a Repository**:

1. Name the repository (`owner/repository`). Only workflows in exactly that repository are trusted; a fork or a copy elsewhere is refused.
2. Optionally narrow it: **Branch or Tag** (one branch or one tag) and a **GitHub Environment**, so only jobs that target that environment are trusted and GitHub can make them wait for a reviewer.
3. Pick the service the repository builds and the environment it deploys to, then copy the workflow. Every value in it — `org`, `audience`, `service`, `environment`, the Planton address — is already filled in, and the repository needs no Actions secret or variable.

The trust acts as a service account. Name none and Planton provisions one for the repository, `ci-<owner>-<repo>`, shared by every trust of that repository and removed with the last trust that uses it. The trust's page shows each run it admitted and what the run shipped. **Pause Trust** and **Remove Trust** take effect at once: the credentials a trust already handed out stop working, not only the next exchange.

**From the CLI or git** — the same trust as a manifest, applied with `planton apply -f`:

```yaml
# workload-identity-binding.yaml — applied with: planton apply -f workload-identity-binding.yaml
apiVersion: iam.planton.ai/v1alpha1
kind: WorkloadIdentityBinding
metadata:
  org: acme
  name: checkout-api-github
spec:
  audience: https://acme.example/planton
  githubActions:
    repository: acme/checkout-api      # matched exactly against the token's own claim
  # serviceAccount: ci-deployer        # optional: an existing service account; omit it and Planton provisions ci-acme-checkout-api
```

There is no API key to create, rotate, or leak — in either mode. Each connected workflow run proves itself with GitHub's own token; the exchanged credential expires in minutes and can act only on the service this repository's token names. Offline mode never talks to Planton at all: the only credentials in play are the ones your provider's own OIDC action minted for the job.

### When a run is refused

A workflow Planton refuses gets one sentence, whatever the cause:

```
the presented token was not accepted by any workload identity binding in this organization
```

That is deliberate: the answer never tells a stranger's workflow whether the organization, a trust, or the repository exists. The real cause is recorded for the organization — one record per repository and reason (no trust names the repository, the run built another branch or tag, targeted another environment, asked for another audience, the trust is paused, or its service account is gone), with what the run presented beside what the trust expects. People who manage access read it:

- on the trust's page in the console (and on **Trusted Workflows**, for repositories no trust names), each refusal with the fix;
- with `planton iam refusal list` and `planton iam refusal get <id>`;
- by asking the Planton Assistant, which reads the same records.

## Inputs

| Input | Mode | Required | Default | Description |
|---|---|---|---|---|
| `environment` | both | yes | | Environment to deploy into. Connected: must be among the service's declared deploy environments. Offline: names the overlay of the working tree's kustomize declaration to render. |
| `org` | connected | yes | | Planton organization slug. Setting `org` AND `audience` selects connected mode. |
| `audience` | connected | yes | | The OIDC token audience — exactly the value your trust declares. The console fills it in: this Planton's own address. |
| `service` | connected | yes | | Service slug or id to deploy. |
| `image` | both | connected: yes | | Container image reference (`host/path:tag` or `@digest`). Connected, the backend injects it per the service's declared targets. Offline, the CLI injects it into the tree's annotated image slots — blank container images receive it, authored images (sidecars) are untouched. |
| `register` | connected | no | `false` | Apply the service manifest before deploying. The manifest's declared repository is **proven** against this workflow's token, never trusted as typed. |
| `service-file` | connected | no | `service.yaml` | Manifest path, used with `register: true`. |
| `follow` | connected | no | `true` | Wait for the run and fail the job honestly (see below). Set `false` to fire-and-forget. Offline deploys run inside this job and are always followed. |
| `set` | offline | no | | Node-addressed field overrides, one per line: `<kind>/<name>:<fieldPath>=<value>`. A tree holds many documents, so the override names its document — the field-exact escape hatch when `image` alone is not enough. |
| `working-directory` | offline | no | `.` | Directory holding the service's kustomize declaration. |
| `cli-version` | both | no | `latest` | Planton CLI version to install. The download is checksum-verified before it is ever executed. |

## Outputs

| Output | Description |
|---|---|
| `run-id` | The delivery run a connected deploy started (`svcpipe_...`). Empty in offline mode — the deploy runs inside this job and has no backend run to point at. |

## What makes the job red

**Both modes fail loudly at the inputs**: a half-state (`org` without `audience`, a `service` in offline mode, a missing `image` in connected mode) fails before any network call, naming the exact line to add or remove — all problems at once, never one at a time.

**Offline**, the exit code is a contract:

- **Refused at preflight** — the report names every problem (schema, unresolvable references, unreachable state backend, invalid credentials, missing modules) and its fix. Nothing was handed to an IaC engine; nothing changed anywhere.
- **A resource failed to deploy** — the engine's own error is in the log, verbatim. Completed resources are safe in state: re-running the job re-applies them as no-ops and continues from the failure.
- **Green** means every resource in the tree is live.

**Connected**, with `follow: true` (the default), green means something:

- The **run failed** (a resource failed to apply) → the job fails with the failed resources named.
- The run succeeded but **rollout verification reported `failed`** — Planton watched the workload on your own cloud and it demonstrably never came online (crash loops, unreachable endpoints) → the job fails with each failed check and the provider's own diagnostic words.
- Verification was honestly **unverifiable** (no declared endpoint, no runner to probe from) → the job passes with a note; honesty is not an alarm.

## Behavior worth knowing

- **Offline needs remote state on CI runners**: runners are ephemeral, so point the engine at your own bucket — `PLANTON_BACKEND_TYPE` + `PLANTON_BACKEND_BUCKET` (+ `PLANTON_BACKEND_REGION`/`PLANTON_BACKEND_ENDPOINT` where the backend needs them) as job env, or the equivalent annotations in the manifests themselves. The preflight report states where state lives and probes that it is reachable with the credentials in hand.
- **Offline cloud auth is the provider's business, deliberately**: use `aws-actions/configure-aws-credentials`, `google-github-actions/auth`, or `azure/login` before this action — the world already trusts them — and the preflight verifies the ambient result honestly.
- **Protected environments (connected)**: a deploy into a protected environment pauses at that environment's approval gate, and the deployer can never be the approver. With `follow: true` the job waits for the approval — bound your job with `timeout-minutes` if that wait should not be open-ended.
- **Connected deploys render from the service's inline deploy declaration**: a service using the kustomize path refuses with the working alternatives named (the git-push lane deploys it fully; promote moves an existing deployment) — or deploy that tree offline, where the kustomize declaration is exactly what renders.
- **Provenance rides along (connected)**: the action passes the workflow's commit and branch, so the run header and history read honestly for an image Planton never built.
- **Runners**: Linux and macOS runners, x64 and arm64. The CLI download is checksum-verified against the release's published checksums before it is ever executed.
