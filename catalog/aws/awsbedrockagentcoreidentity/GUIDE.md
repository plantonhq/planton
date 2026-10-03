# AwsBedrockAgentCoreIdentity — Kind Guide

Authored operational judgment for the AgentCore identity component: the
design decisions behind the spec's shape, and what to know before
running agent credentials in production.

## Design decisions

- **One bundle, four arms.** Workload identities, credential providers,
  and the policy engine are AWS-standalone resources, but they form one
  identity-and-access control plane a team manages together — the kind
  bundles them as name-keyed collections, each arm optional (at least
  one, CEL-guarded).
- **The vendor field IS the discriminator.** The provider spells six
  structurally-identical OAuth vendor blocks; the spec carries one
  `vendor` enum plus the shared `client_id`/`client_secret` pair, and
  the modules select the block (the fan-in is recorded in the parity
  manifest). CUSTOM pairs exactly with `oauth_discovery`.
- **Write-only credential variants are excluded by design.** The spec's
  sensitive fields arrive just-in-time resolved, and the plain provider
  arguments let rotation be detected; a write-only value the provider
  cannot read back would make rotation invisible to state.
- **The token-vault CMK is deliberately absent.** It sets the KMS key on
  the account/region's ONE default vault (delete is a no-op) — an
  account-level settings singleton; folding it here would make multiple
  Identity instances fight over one object.
- **A Cedar policy is a structural child of its engine** — created
  after, destroyed before, imported as `{policy_engine_id},{policy_id}`.

## Running agent credentials in production

- **Reference providers, never secrets.** Consumers (gateway targets,
  harness tools) take the provider ARN; the secret lives in the token
  vault. Rotate by re-applying with a new JIT-resolved value — consumers
  never change.
- **Names are identity and ForceNew.** Renaming a provider replaces it —
  in-flight token grants against the old name die; rename during quiet
  windows.
- **Scope one bundle per trust domain.** One Identity per team/agent
  fleet keeps the blast radius of a leaked provider ARN grant small and
  the Cedar policy set readable.
- **Prefer FAIL_ON_ANY_FINDINGS.** Cedar's static analysis catches
  policies that can never match; IGNORE_ALL_FINDINGS is for deliberate
  forward-references, not a default.
- **Always constrain the Cedar `resource`.** AWS rejects a fully-wildcard resource at CreatePolicy ("a wildcard resource was detected" — live-verified 2026-08-14) regardless of validation mode: scope every statement to a specific `AgentCore::Gateway` entity or at least the type (`resource is AgentCore::Gateway`).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
