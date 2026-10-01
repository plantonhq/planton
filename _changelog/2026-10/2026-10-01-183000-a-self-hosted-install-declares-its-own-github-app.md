# A Self-Hosted Install Declares Its Own GitHub App

**Date**: October 1, 2026
**Type**: Feature
**Components**: KubernetesPlantonPlatform, Planton operator

## Summary

**A self-hosted install can now give every organization a one-click GitHub connection.** The platform team registers one GitHub App for the whole install on each GitHub host the company uses and declares it under `github` on the KubernetesPlantonPlatform resource. Every organization's connection wizard then offers that App. Installation tokens are signed with its key, and every push is checked against its webhook secret before anything runs.

**The key stays exactly as GitHub generated it.** The App's PEM private key and webhook secret live in a Secret in the platform's namespace, named by reference. The operator mounts them as files, so rotating the key is a Secret edit that is live on the next token, with no restart.

**The operator stops handing the control plane placeholder GitHub credentials.** Platform release v0.0.113 reads the install's GitHub declaration from the operator's facts file alone. So the operator no longer renders the dummy App client id, key and webhook secret, or the variables that marked the platform App unavailable and turned host login on. The operator's oldest supported platform release moves to v0.0.113.

## What Changed

- **KubernetesPlantonPlatform:**
  - `spec.github` mirrors the operator's `GithubSpec`: `hosts[]` of `host`, an optional `app` (`client_id`, `private_key_secret_ref`, `webhook_secret_ref`) and a `webhooks` posture (`auto`, `reachable`, `unreachable`), plus `host_login`;
  - validation refuses a host written as a URL, the same host twice, an App without a client id or key reference, and an unknown webhook posture;
  - the Pulumi and OpenTofu modules render the same map into the custom resource, pinned by a module test;
  - a new preset, One GitHub App for the Whole Install, and a GUIDE section on registering the App and its Secret;
  - the README table and catalog page describe the field; the kind's scenarios and guide examples move to v0.0.113.
- **Operator:**
  - drops `GITHUB_APP_CLIENT_ID`, `GITHUB_APP_PRIVATE_KEY_BASE64`, `GITHUB_WEBHOOKS_SECRET_TOKEN`, the platform App's availability and reason, and host login's availability from the control plane's environment;
  - `MinimumSupported` moves to v0.0.113, with its reason; the boot contract fixture follows.
- **Generated:** the kind's Go stub, reference page and the proto-docs index.
