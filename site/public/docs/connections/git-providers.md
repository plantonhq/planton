---
title: "Git Providers"
description: "Connect GitHub and GitLab organizations to enable automated builds, webhooks, and pull request deployments in Service Hub"
icon: connect
order: 30
tags:
  - Connect
  - Git Providers
  - GitHub
  - GitLab
---

# Git Providers

Service Hub builds and deploys code from your Git repositories. For that to work, Planton needs read access to your repositories and a way to learn about pushes. Git provider connections establish that access at the organization level, so every service you create can pull code without engineers managing individual access tokens.

## Why Git Connections Matter

Without a git connection, Service Hub cannot:

- Clone your repository to run builds
- Learn about pushes to start pipelines (from GitHub's webhooks with an App, or by watching the repository with a sign-in)
- Report build status back to pull requests (an App identity)
- Access private repositories

A single git connection covers your entire GitHub or GitLab organization. Every service that references a repository under that organization automatically uses the connection — no per-repository configuration needed.

## GitHub

GitHub is the primary git provider in Planton. A GitHub connection signs in as one of two identities, and which one is right depends on where Planton runs:

- **The sign-in on this machine** — when Planton runs on your own laptop or on a self-hosted server, the machine is usually already signed in to GitHub (`gh auth login`, or a `GH_TOKEN` / `GITHUB_TOKEN` in its environment). The connection names that account and stores no token: Planton reads the machine's sign-in each time it reaches GitHub, and signing out stops the connection the same second.
- **A GitHub App** — Planton's own App, or one you register yourself (including on GitHub Enterprise Server). This is the identity for hosted Planton and for anything only an App can do: GitHub starting runs on push through webhooks, and check runs reported back to pull requests.

Planton never uses a pasted personal access token as a connection. An App is GitHub's integration identity; a sign-in is something the machine already holds. Nothing is copied.

| | Sign-in on this machine | GitHub App |
|---|---|---|
| Clones and reads repositories | Yes | Yes |
| GitHub starts runs on push (webhooks) | No — Planton watches the repository for pushes instead | Yes |
| Check runs on pull requests | No | Yes |
| Where it fits | A laptop; a self-hosted server with a token in its environment | Hosted Planton; GitHub Enterprise Server with your own App |

### The Sign-In on This Machine

On a local instance, GitHub is a detected card on the Connections page, beside your cloud accounts. Click **GitHub** and Planton shows the account this machine is signed in as — *GitHub Account priya-dev · github.com · Signed in with gh* — and offers **Use This Account**. One click writes a connection that holds no token, then proves it: Planton reads the sign-in once and asks GitHub who it belongs to. The connection is called ready only when GitHub has answered.

From the terminal, the same connection is one command:

```bash
# Which GitHub account is this machine signed in as?
planton connect github detect

# Save it as a connection (a specific account, when the machine holds several)
planton connect github detect --account priya-dev --yes
```

If the machine is not signed in yet, the card and the command say exactly what to do: `gh auth login`, or export `GH_TOKEN` in your shell profile and restart Planton (a shell token is read when Planton starts; `gh auth login` takes effect at once).

### The GitHub App

The App path is an installation flow, never a credential you paste:

1. You initiate the connection from the Planton console and choose **Planton GitHub App** (or **Your Own GitHub App**).
2. Planton sends you to GitHub to install the App.
3. You choose which GitHub organization (or personal account) to install it on.
4. GitHub returns you to Planton with the installation, and the connection is created.

Advantages of the App identity:

- **Scoped permissions** — the App requests only what it needs (repository read, webhooks, commit status).
- **Organization-level** — one installation covers every repository in the organization.
- **No token rotation** — GitHub mints short-lived installation tokens; nothing to rotate by hand.
- **Audit trail** — GitHub logs every API call the installation makes.

<!-- SCREENSHOT: GitHub App installation page
  Page: GitHub.com App installation flow
  Action: Show the GitHub App installation approval screen
  Focus: The organization selector and permission list
  Alt: GitHub App installation page showing organization selection and permission approval
-->

### Connecting via the Web Console

1. Navigate to **Connections** and click the **GitHub** card.
2. On a local instance, the detected card offers the account this machine is signed in as — **Use This Account**. Otherwise (or with **Set Up Manually**), the wizard's **Identity** step asks how Planton should sign in: **Sign-In on This Machine**, **Planton GitHub App**, or **Your Own GitHub App**. An instance only offers the identities it can honor.
3. For an App, GitHub asks you to install it; for the sign-in, you confirm the account.
4. The connection's page shows **Verify Sign-In**: press it and read **Confirmed — GitHub attributes this sign-in to priya-dev**, or the exact sentence explaining what to fix.

### Your Own GitHub App: Webhook URL and Secret

With **Your Own GitHub App**, GitHub sends the App's pushes and pull requests to whatever Webhook URL the App holds, so the App's settings must point at Planton. The wizard's App step shows the **Webhook URL** to copy into them.

Where GitHub can reach Planton (hosted Planton, or a self-hosted install with a public address), a webhook is the only way pushes arrive. Planton acts only on a delivery whose signature proves it came from your App, since an unsigned "push" could come from anyone who can reach the address. So the connection needs a **webhook secret**:

- The wizard's **Webhook Security** step generates one and stores it as a Planton secret. Paste the same value into the App's **Webhook secret** field on GitHub, beside the Webhook URL, with **Active** ticked.
- A connection made without one (through `planton apply`, say) is refused, naming the receiver and the fix.
- A delivery Planton cannot verify is refused whole: no build starts, and no `service.yaml` registers a service.

**Verify** on the connection's page also reports where GitHub sends the App's pushes. It says whether the App's Webhook URL points at this Planton and whether the App subscribes to push events, and names the setting to change if not. GitHub never reveals whether the webhook is ticked **Active** or which secret it signs with, so only the App's **Recent Deliveries** page on GitHub proves those.

Where GitHub cannot reach Planton (a laptop, or a server behind a firewall), no secret is asked for: Planton reads pushes from GitHub's API instead, and the wizard skips the webhook step.

### Connecting via the CLI

```bash
# Detect the sign-in on this machine and save it as a connection
planton connect github detect

# List repositories accessible through any GitHub connection
planton connect github list-repos my-github-org
```

### GitHub Enterprise (Self-Hosted)

If your organization uses GitHub Enterprise Server, both identities work against it. For the sign-in, the machine's `gh` login (or `GH_ENTERPRISE_TOKEN`) for that host is what the connection names, and the connection records the server URL. For an App, choose **Your Own GitHub App** and register it on your server.

### Account Types

GitHub connections support both organization accounts and personal accounts. Organization accounts are the typical choice for teams — they cover all repositories under the organization. Personal accounts are useful for individual developers or for connecting repositories that aren't under an organization.

---

## GitLab

GitLab connections use an access token and refresh token pair to authenticate. Unlike GitHub's OAuth App model, GitLab connections require you to provide a group ID and tokens.

### Connecting via the Web Console

1. Navigate to **Connections** and click the **GitLab** card under DevOps Pipeline.
2. **Name your connection** and provide the required credentials.
3. **Create the connection**.

> **Note**: GitLab integration is currently being finalized in the web console. The connection type is available in the API, but the web console wizard may show limited functionality. Check back for updates as GitLab support is expanded.

### What You Need

| Field | Description |
|-------|-------------|
| GitLab Host | Your GitLab instance URL (defaults to `https://gitlab.com`) |
| Group ID | The GitLab group whose repositories Planton will access |
| Access Token | A GitLab personal access token or group access token |
| Refresh Token | The refresh token for token renewal |

### GitLab Self-Hosted

GitLab connections support self-hosted GitLab instances. Provide your instance's URL as the GitLab host. This is useful for organizations that run GitLab on their own infrastructure for compliance or data residency requirements.

### Creating GitLab Tokens

In GitLab:

1. Navigate to your group's **Settings > Access Tokens** (for group tokens) or **User Settings > Access Tokens** (for personal tokens).
2. Create a token with the `api` scope for full API access, or scope it down to `read_repository` and `write_repository` for minimal permissions.
3. Copy both the access token and the refresh token.

---

## How Git Connections Are Used

Once a git connection is established, Service Hub uses it automatically when you create a service:

1. **Service creation** — When you configure a service's source repository, Planton uses the git connection to verify the repository exists and is accessible.
2. **Pipeline execution** — When a pipeline runs, Planton clones the repository using the git connection. For monorepos, sparse checkout is used to pull only the relevant directory.
3. **Learning about pushes** — with a GitHub App, GitHub delivers push, pull request, and tag events to Planton through the App's own subscription (nothing is registered on your repository); with the sign-in on a machine, Planton watches the repository for new commits and starts the run itself.
4. **Status reporting** — build and deployment results are reported back to the git provider as commit statuses or check runs (an App identity).

You don't need to specify which git connection to use on each service — Planton resolves it automatically based on the repository's organization.

## Related Documentation

- [Connections Overview](/docs/connections) — Understanding the Connect system
- [CI/CD](/docs/ci-cd) — How services use git connections for builds and deployments
- [Container Registries](/docs/connections/container-registries) — Where build artifacts are stored
