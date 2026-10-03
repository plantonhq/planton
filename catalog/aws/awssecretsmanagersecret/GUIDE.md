# AwsSecretsManagerSecret — Kind Guide

The authored wisdom layer for this component: internal conventions, judgment
calls, and operational judgment earned while building it. The reference for
fields is `v1alpha1/reference.md`; this file explains the decisions the
schema alone cannot.

## Design decisions

- **The policy rides the standalone `aws_secretsmanager_secret_policy`
  resource, never the secret's inline `policy` argument.** Only the
  standalone resource carries `block_public_policy` — the PutResourcePolicy
  validation gate that rejects policies granting anonymous access. The
  inline argument is the same API surface minus the guard, so the module
  never sends it (recorded as a one-arg-two-homes exclusion in the parity
  manifest). `block_public_policy` defaults ON: a public secret policy is
  almost always a mistake, and the opt-out is explicit.
- **Custom staging labels always ride WITH `AWSCURRENT`.** PutSecretValue's
  `VersionStages` REPLACES the automatic AWSCURRENT assignment — sending
  only the custom labels would leave the secret with no current version.
  The modules concat `["AWSCURRENT"] + version_stages` on both engines.
- **`type` is the managed-external-secret partner identifier**, not a
  general classification field. It pairs with the external rotation arm
  (`rotation.externalRotationRoleArn`). The module omits the argument
  entirely when empty — AWS treats absent and empty differently in error
  paths.
- **The rotation block orders after the version** on both engines: with
  `rotateImmediately` (the default) AWS invokes the rotation mechanism as
  soon as it is configured, and the rotation function reads the current
  value — configuring rotation on a valueless secret fails the first
  rotation.
- **`secret_string_wo` (the provider's write-only value variant) is
  deliberately not modeled.** It exists to keep the value out of Terraform
  state; on this platform the value is already a managed-secret reference
  resolved just-in-time, and Planton encrypts every state file under its
  backend's key. One value arm
  per encoding keeps the spec honest (recorded in the parity manifest).

## Operational judgment

- **Recreating a same-named secret during the recovery window** hits AWS's
  "scheduled for deletion" error; the provider retries through it, but the
  create can stall for minutes. For ephemeral secrets (tests, previews)
  always set `recoveryWindowInDays: 0` so destroy frees the name
  immediately.
- **Cross-account consumers need BOTH sides**: a resource policy statement
  granting the reader, AND a customer-managed KMS key whose key policy
  grants the reader `kms:Decrypt` — the AWS-managed key cannot be shared,
  and a policy-only grant fails at GetSecretValue with a KMS error, which
  reads like a permissions bug on the wrong service.
- **Replica regions and KMS**: a replica encrypts under a key IN ITS OWN
  region. Referencing the primary region's key ARN in a replica fails at
  replication time, not at plan.
- **Destroying a replicated secret: keep a recovery window** (live-verified
  2026-08-13). AWS deletes replica secrets ASYNCHRONOUSLY after replication
  is removed (~40-90s), and neither engine waits for it — with
  `recoveryWindowInDays: 0` the primary's immediate force-delete can outrun
  that deletion and strand the replica as a live, billing, standalone
  secret in its region. A non-zero window keeps the primary's tombstone
  alive long enough for the async deletion to complete.
- **Recovering a stranded replica**: it rejects a direct delete
  ("Operation not permitted on a replica secret") even when its primary no
  longer exists. Run `aws secretsmanager stop-replication-to-replica` in
  the REPLICA's region (promotes it to standalone), then `delete-secret`.
- **A swept replica can COME BACK** (live-observed 2026-08-13): the
  replication machinery of a force-deleted primary can re-materialize the
  replica-region secret tens of minutes later (a fresh CreatedDate, same
  inherited ARN suffix, PrimaryRegion pointing at the deleted primary).
  After cleaning up a stranded replica, re-check the region ~an hour
  later; repeat the promote-then-delete if it resurfaced. A recovery
  window on the primary avoids the whole class.
- **A stranded ex-replica blocks the name for future replication** — new
  replication into that region fails with "currently replicated to
  <region> with a different arn", `forceOverwriteReplicaSecret: true` does
  NOT clear it, and the failure is SILENT at apply (neither engine waits
  on replication status). Check `DescribeSecret.ReplicationStatus` reaches
  `InSync` after enabling replication into a region with name history.
- **Rotation Lambda permission**: RotateSecret fails with
  AccessDeniedException until the function grants
  `secretsmanager.amazonaws.com` invoke permission. The provider retries
  through IAM propagation delay, but a missing permission (as opposed to a
  propagating one) fails the deploy after the retry budget.

## Coverage decisions

- `binary_value` covers `secret_binary` (base64 in, decoded by the
  provider). The write-only variants (`secret_string_wo`,
  `secret_string_wo_version`) are excluded by design (see above).
- Live E2E defers the rotation arms (both need real rotation
  infrastructure); the render is offline-proven arm-for-arm. Unblock: a
  minimal rotation-handler Lambda fixture.
