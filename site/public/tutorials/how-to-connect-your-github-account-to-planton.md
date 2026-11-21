---
title: "How to Connect Your GitHub Account to Planton"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "github"
  - "connect"
  - "scm-connection"
  - "getting-started"
category: "connect"
excerpt: "Connect your GitHub repositories to Planton -- from the sign-in your machine already holds, or with the Planton GitHub App -- so the platform can clone code, learn about pushes, and deploy your services."
---

# How to Connect Your GitHub Account to Planton

Before Planton can build and deploy your code, it needs access to your GitHub repositories. A GitHub connection gives it that access as one of two identities, and this tutorial walks through both: **the sign-in your machine already holds** (the path for Planton Desktop on your laptop, or a self-hosted server), and **the Planton GitHub App** (the path for hosted Planton). Neither path ever asks you to paste a personal access token.

## What You Will Learn

- The two identities a GitHub connection can sign in as, and which one fits where Planton runs
- How to turn the `gh` sign-in on your machine into a connection that stores no token -- from the desktop or the CLI
- How to install the Planton GitHub App on your GitHub organization or personal account
- How to prove a connection works the moment it exists, and list the repositories it can see

## Prerequisites

- [ ] A Planton organization -- on Planton Desktop, a hosted account, or a self-hosted instance
- [ ] The `planton` CLI installed and authenticated
- [ ] For Path A: the GitHub CLI signed in on the machine running Planton (`gh auth login`), or a `GH_TOKEN` exported in its environment
- [ ] For Path B: a GitHub account with **owner** or **admin** access to the organization you want to connect (or a personal account)

## How GitHub Connections Work

A GitHub connection describes how Planton signs in; it never holds a pasted token.

- **The sign-in on this machine.** When Planton runs on your own machine, the connection names the GitHub account that machine is signed in as. Planton reads the machine's sign-in (`gh`'s credential store, or the token in its environment) each time it reaches GitHub, and stores nothing. Sign out of `gh` and the connection stops the same second. This identity clones and reads everything your account can see; because GitHub cannot call a machine it does not know, Planton watches your repositories for pushes instead of receiving webhooks, and check runs on pull requests are not available.
- **A GitHub App.** Planton's own App (or one you register) is installed on your organization. GitHub mints short-lived installation tokens, delivers push events to Planton, and accepts check runs. This is the identity for hosted Planton.

For more, see the [Git provider connections documentation](/docs/connections/git-providers).

## Path A: The Sign-In on This Machine

### Step 1: Confirm the Machine Is Signed In

```bash
gh auth status
```

You should see the account and host you expect (`Logged in to github.com account priya-dev`). If not, run `gh auth login`, or export `GH_TOKEN` in your shell profile and restart Planton -- a shell token is read when Planton starts; `gh auth login` takes effect at once.

### Step 2: Use the Detected Card, or One Command

In Planton Desktop, open **Connections** and click **GitHub**. The card shows the account this machine is signed in as -- *GitHub Account priya-dev · github.com · Signed in with gh* -- and offers **Use This Account**. Click it.

Or, from the terminal:

```bash
planton connect github detect
```

Either way Planton writes a connection named after the account (`github.account.priya-dev`), then proves it: it reads the sign-in once and asks GitHub who it belongs to. You read **GitHub Connection Ready** -- *Confirmed -- GitHub attributes this sign-in to priya-dev* -- or the exact sentence explaining what to fix.

### Step 3: Verify Any Time

The connection's page has **Verify Sign-In**. Press it whenever you like; it answers **Confirmed** with the account GitHub attributes the sign-in to, or **Sign-In Not Working Yet** with the `gh auth login` to run. From the CLI:

```bash
planton connect github list-repos github-account-priya-dev
```

lists the repositories the sign-in can see.

## Path B: The Planton GitHub App

### Step 1: Start the Installation from the Console

Open the console and navigate to **Connections**. Click **GitHub**, then **Set Up Manually** if a detected card is offered. On the **Identity** step choose **Planton GitHub App** (or **Your Own GitHub App** on GitHub Enterprise Server), and give the connection a **name** (for example `acme-github`).

The console then directs you to install the App. This opens GitHub's App installation page in your browser.

### Step 2: Install the App on GitHub

On GitHub's installation page, you choose:

1. **Which account** to install the App on -- your personal account or one of your GitHub organizations
2. **Which repositories** to grant access to -- all repositories, or a selection of specific repositories

For most teams, installing on the organization account and granting access to **all repositories** is the simplest approach. You can change this scope later from your GitHub organization settings without affecting the Planton connection.

After you approve the installation, GitHub redirects you back to the console with the installation, and the connection is created.

### Step 3: Verify the Connection

The completed screen shows **Verify Sign-In**: press it and read **Confirmed** with the installation GitHub minted the token for. From the CLI:

```bash
planton get github-connection acme-github
planton connect github list-repos acme-github
```

If you do not see a repository you expect, check the App's repository access settings in your GitHub organization: **Settings > Integrations > GitHub Apps > Planton > Configure**.

## Connecting GitHub Enterprise Server

Both identities work against GitHub Enterprise Server. For the sign-in on this machine, `gh auth login --hostname github.enterprise.corp` (or `GH_ENTERPRISE_TOKEN`) is what the connection names, and the connection records the server URL:

```yaml
apiVersion: connect.planton.ai/v1alpha1
kind: GithubConnection
metadata:
  name: github.account.priya@github.enterprise.corp
  org: your-org
spec:
  authMode: host_login
  accountLogin: priya
  host:
    value: "https://github.enterprise.corp"
```

For an App, choose **Your Own GitHub App** in the wizard and register it on your server; the connection carries your App's client ID, private key reference, and the installation.

## What to Do Next

- **Connect a container registry** so Planton can push the images it builds: [How to Connect a Container Registry to Planton](/tutorials/how-to-connect-a-container-registry-to-planton)
- **Deploy your first service** with zero-config CI/CD: [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd)
