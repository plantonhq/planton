# Immutable Backup Posture

This preset creates the ransomware-readiness framework: are recovery
points encrypted, protected from manual deletion, and held in locked
vaults.

## When to Use

- Proving the immutable-backup story to auditors and insurers
- Watching that vault-lock coverage keeps up as new vaults appear

## What You Get

- Continuous encryption, deletion-protection, and vault-lock checks
  across the account's recovery points
- Evidence that pairs directly with the air-gapped vault and
  compliance-mode lock levers on
  [AWS Backup Vault](/infra-catalog/aws-backup-vault)

## Customize

- Add the coverage pair from the `backup_coverage_audit` preset for
  one consolidated framework
- Wire the framework's ARN into an
  [AWS Backup Report Plan](/infra-catalog/aws-backup-report-plan)
  compliance template for scheduled evidence in S3
