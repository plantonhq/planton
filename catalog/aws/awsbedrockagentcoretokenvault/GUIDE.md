# AwsBedrockAgentCoreTokenVault — Kind Guide

Authored operational judgment for the AgentCore token-vault
singleton: the design decisions behind the spec's shape, and what to
know before putting agent credentials under your own key.

## Design decisions

- **A settings singleton, not an Identity field.** The vault is ONE
  account/region object every AgentCore Identity resource shares —
  folding its key setting into AwsBedrockAgentCoreIdentity would make
  multiple Identity instances fight over it (that fold was re-judged
  out on schema evidence).
- **The key pairing is CEL-enforced.** `CustomerManagedKey` requires
  `kms_key_arn`; `ServiceManagedKey` forbids it. The provider would
  accept the stray ARN silently; the spec refuses it.
- **`token_vault_id` exists for AWS's forward surface.** Today every
  account has exactly one vault, "default" — the field is modeled
  because the API keys on it, with the module defaulting it so
  manifests never need to care.

## Operating vault encryption in production

- **THE trap: destroy does not revert.** The provider's delete is a
  no-op — the vault keeps the last-applied key forever. The revert is
  an APPLY with `ServiceManagedKey`. Concretely: never schedule the
  KMS key's deletion while the vault still points at it, or every
  stored agent credential becomes unreadable.
- **The revert itself needs the OLD key alive (live-caught).** `SetTokenVaultCMK` validates the previous key's state on every write — a vault pointing at a disabled or pending-deletion key refuses even the revert to `ServiceManagedKey` with `Old KMS Key validation failed ... expected KeyState:ENABLED`. Always revert FIRST, retire the key SECOND.
- **Wedged anyway? The deletion window is your way out.** While the key is still in its 7–30 day deletion window: `aws kms cancel-key-deletion --key-id <id>`, then `aws kms enable-key --key-id <id>`, apply `ServiceManagedKey` (or run `aws bedrock-agentcore-control set-token-vault-cmk --token-vault-id default --kms-configuration keyType=ServiceManagedKey`), and only then re-schedule the key's deletion. Once the key is truly gone, only AWS Support can help.
- **Key revocation is an agent outage.** AgentCore reads the vault on
  every credential fetch; revoking the key's grants or disabling the
  key locks every agent out of its OAuth tokens and API keys at once.
- **Same-region, symmetric keys only.** AWS creates its own grants on
  the key when the setting is applied; cross-region or asymmetric
  keys are rejected.
- **Order of operations for adopting a CMK:** create the key, apply
  this component, THEN start storing credentials — existing
  credentials are re-encrypted by AWS on the switch, but doing it
  before first use keeps the audit story clean.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
