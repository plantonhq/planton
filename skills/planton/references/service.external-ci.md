---
title: Keyless CI — Workload Identity Bindings, CI-Step Registration, and the Deploy Step
description: External CI (GitHub Actions, gitlab.com) calls Planton with zero stored secrets — the org trusts one repository (a workload identity binding, "Trusted Workflows" in the console), the CI job exchanges its provider's OIDC token via planton iam federate, and the credential can register and deploy exactly the service its token proves. Read when someone asks how CI authenticates, how to deploy from a CI step, how to use the Planton GitHub Action, or why a run was refused (the workflow's answer is deliberately identical; the organization's refusal records say which condition missed).
---

# Keyless CI — Workload Identity Bindings, CI-Step Registration, and the Deploy Step

External CI (GitHub Actions, gitlab.com CI) can call Planton without any stored secret: the CI job presents its provider's own short-lived OIDC token, and Planton exchanges it for a short-lived scoped credential that can register and deploy the one service the token proves. Read this when someone asks how CI authenticates to Planton, how to register or deploy a service from a CI step, how to use the Planton GitHub Action, why a run was refused, or how to stop trusting a repository.

## The trust binding

A **workload identity binding** is the org-registered rule that makes the exchange possible: "tokens from THIS provider, matching THESE conditions, may act as THIS service account." The console calls it a **trust**: **Organization Settings → Trusted Workflows → Trust a Repository** makes one and hands back the finished GitHub workflow, with every value filled in. It is also declarative YAML (`WorkloadIdentityBinding`, applied like any resource), and reads/writes require **manage-access on the organization** — trusting an external identity is an access-management act, so ordinary org write access is deliberately not enough. Hold these facts when explaining one:

- The binding acts as a **service account** — the identity CI acts as, the name audit trails show. `serviceAccount` is optional, and leaving it out is the usual choice: Planton provisions `ci-<owner>-<repo>` (display name `CI: <owner>/<repo>`). Every trust of that repository shares it, and it is removed with the last trust that uses it. A named account must already exist, and an account a person made is never removed with a trust. An account a trust uses cannot be deleted; the refusal names the trusts.
- Conditions are **structured**, matched against the provider's own token claims: the repository (GitHub `owner/name`, GitLab `group/project`, always required, matched exactly — one binding trusts one repository), and optionally a ref or environment. There are no patterns or wildcards to author.
- The **audience is required** and should be a value your CI explicitly requests the token with (GitHub's `audience` parameter). It is what makes Planton the token's only valid consumer.
- Each provider's **issuer is pinned in Planton's code** — a binding never carries an issuer URL. GitHub Actions and gitlab.com are supported; self-hosted GitLab instances are not yet (their issuers are customer URLs, which need a hardened key-fetch design — say so plainly rather than improvising).
- A binding **never widens authority**: exchanged credentials carry the external-CI class's fixed rule set (registering the service, deploying it, and reading the runs and receipts those verbs produce — never approving anything), and binding a powerful service account grants CI nothing extra — work-credential authorization never consults the account's standing grants.
- **Pausing** (`disabled: true`, **Pause Trust** in the console, which asks why) suspends the trust without deleting the record; **deleting** the binding (**Remove Trust**) is the off-switch. Both take effect at once: every live credential the trust admitted is revoked, so a run already holding one stops at its next call. The repository, provider and service account never change; to trust another repository, make another trust.

## The exchange, in a CI step

```bash
export PLANTON_API_KEY=$(planton iam federate --org acme --token "$OIDC_TOKEN")
planton service register -f service.yaml
```

`planton iam federate` prints the raw credential to stdout and nothing else, so capture works exactly as above; the token comes from `--token`, stdin (`--token -`), or `$PLANTON_OIDC_TOKEN`. In GitHub Actions the job needs `permissions: id-token: write` and requests its token with the binding's audience. The exchanged credential is a standard bearer — `PLANTON_API_KEY` accepts it.

The CI token is **exchange-only**: it can never be used directly as a Planton credential. Only the exchanged `pwk_` credential calls APIs.

## Reading a refused run

The workflow's answer is **identical by design**: "the presented token was not accepted by any workload identity binding in this organization", so the endpoint never reveals whether the org, a binding, or a repository exists. The cause is not lost. For every token that verified (GitHub or gitlab.com really signed it), the organization records the refusal: one record per repository and reason, with the platform's own `explanation` sentence, what the token presented (`observed`) beside what the nearest trust expects (`expected`), that trust's id (empty when no trust names the repository), how many runs it groups, and the most recent run with its link.

Only people with **manage-access on the organization** can read refusals, the same people who manage the trusts. Read them before touching any trust:

- MCP: `list_workload_identity_refusals` (the org's records, newest first) and `get_workload_identity_refusal` (one, by its `wir_` id).
- CLI: `planton iam refusal list` (`--trust <slug-or-id>` for one trust; `-o json` for the records) and `planton iam refusal get <id>`.
- Console: **Trusted Workflows** and each trust's page show them, each with the change that fixes it.

Render `explanation` verbatim, then fix by reason:

| Reason | What happened | The fix |
|---|---|---|
| `no_trust_for_repository` | No trust in the org names the repository | If the repository is theirs, trust it (create a binding). If not, nothing to do: anyone can aim a signed token at any organization, so never trust a stranger's repository to make a refusal go away. |
| `ref_mismatch` | The trust admits only another branch or tag | Either widen the trust's ref (update the binding), or trust the run's ref as well with a second trust, or run the workflow from the ref the trust names. Say what each choice lets ship. |
| `environment_mismatch` | The trust admits only another GitHub environment | Point the job at the trusted environment, or change the trust's environment. |
| `audience_mismatch` | The token was requested for another audience | Make the workflow's `audience` exactly the trust's (the console's default is this Planton's origin, `https://<host>`). |
| `trust_paused` | The trust that would admit the run is paused | Resume it (clear `disabled`), after reading the pause reason on the trust's page. |
| `service_account_missing` | The trust acts as a service account that no longer exists | Point the trust at an existing account by recreating the trust (the account never changes on update). A trust that names no account gets a provisioned one. |

When the records show nothing for a run the person says was refused, the token itself never verified: the wrong issuer, an expired token (they expire in minutes), or a request for another organization. Walk those by hand.

## The deploy step

A CI job that built its own image deploys it with the same credential:

```bash
export PLANTON_API_KEY=$(planton iam federate --org acme --token "$OIDC_TOKEN")
planton service deploy checkout-api --org acme --env prod \
  --image ghcr.io/acme/checkout-api@sha256:... \
  --commit "$GITHUB_SHA" --branch "$GITHUB_REF_NAME" --follow
```

The deploy renders manifests from the service's CURRENT inline deploy declaration with the image injected, and rides the same engine as every deploy: protection gates, separation of duties (the deployer can never approve), URLs, and rollout verification. Facts to hold:

- **`--follow` tells the truth in CI**: the command exits non-zero when the run fails AND when rollout verification reports `failed` (the workload demonstrably never came online) — the failed checks print with the provider's own words. An honestly `unverifiable` rollout passes with a note. This is why a green deploy step means something.
- **Kustomize-source services refuse** the deploy verb (their rendering runs on the build lane) — the working paths are the git-push lane or promoting an existing deployment.
- **The environment must be declared** in the service's deploy configuration; the refusal names the declared set.
- **A protected environment parks the run at its approval gate** — a followed CI job waits for the approval; set the job's own timeout accordingly.

## The GitHub Action

`plantonhq/planton/actions/deploy` wraps the whole story — install the CLI (checksum-verified), mint the job's OIDC token, federate, optionally register, deploy, and wait honestly. The job needs `permissions: id-token: write`, and the `audience` input must be exactly the binding's audience. The console writes the whole workflow file for a trust (`.github/workflows/planton-deploy.yml`: build and push to `ghcr.io/<repository>:<sha>` with `GITHUB_TOKEN`, then this action), with its trigger matching the trust's branch or tag. Hand that file over, not a hand-written one. `register: true` applies the service manifest first (its repository proven by the same token). See the action's own README for the full input table. The same action also runs fully backendless: with `org` and `audience` absent it deploys the repository's own kustomize declaration through the open-source engine — offline mode, covered by the action's README.

## Proven repository identity

A service registered with a federated credential may only declare the repository the credential's own verified token names — a mismatch is refused, naming both repositories. This is the point of CI-step registration: a catalog entry whose repository was typed can drift or lie; one registered from the repository's own CI is **proven**. A service that declares no repository (a pure catalog entry) registers fine from any lane.

The same law narrows the delivery verbs, and MORE strictly: a federated credential may deploy, promote, or roll back **only the service whose declared repository matches its token's** — and here a repo-less service refuses (with no declared repository there is no proof relation at all, so the answer is to set the service's repository or deliver from a lane with standing access). The deploy rules themselves are org-wide by grammar; this per-service narrowing is what keeps one repository's workflow from deploying every service in the organization.
