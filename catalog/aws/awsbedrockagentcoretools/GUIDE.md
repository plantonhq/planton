# AwsBedrockAgentCoreTools — Kind Guide

Authored operational judgment for the AgentCore tools component: the
design decisions behind the spec's shape, and what to know before
handing agents browsers and code sandboxes in production.

## Design decisions

- **One bundle, three arms.** Browsers, profiles, and code interpreters
  are AWS-standalone resources, but they form one tool belt a team
  provisions together — name-keyed collections, each arm optional (at
  least one, CEL-guarded).
- **Everything is replace-on-change and the spec says so.** AWS exposes
  no update for any of the three resources; the kind's comments carry
  that truth so nobody hunts for a phantom in-place path.
- **The network modes differ per tool deliberately.** Browsers take
  PUBLIC|VPC; code interpreters add SANDBOX (no network) — modeled as
  separate messages rather than one watered-down union, because SANDBOX
  on a browser is not a thing AWS accepts.
- **Certificates and enterprise policies are S3/Secrets Manager
  locations** — single-member location wrappers flattened to their ARN /
  object leaves, recorded in the parity manifest.

## Handing agents tools in production

- **Default code to SANDBOX.** Model-written code gets network access
  only when the task genuinely needs it — and then prefer VPC posture so
  egress rides your controls.
- **Recording is your audit trail.** For browsers touching authenticated
  sites, enable recording to S3 and lifecycle the prefix — replay is how
  you answer "what did the agent actually do?"
- **Profiles hold credentials — treat them like credentials.** A saved
  login in a profile is a live session for whoever starts a browser from
  it; scope profile use narrowly.
- **Enterprise policies are your kill switch** for browser capabilities
  (downloads, extensions, URL allow-lists) — the MANAGED type enforces;
  RECOMMENDED merely suggests.
- **Recreates are cheap; sessions are not.** A changed tool replaces the
  shell — plan rollouts so long-running sessions drain before you drop
  the old tool's grants.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
