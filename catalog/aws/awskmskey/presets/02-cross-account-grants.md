# Shared Key with Grants

This preset creates a rotating symmetric key whose access is delegated through KMS grants instead of key-policy edits: one grant wires a workload role (referenced from an `AwsIamRole` in the same chart) to encrypt and decrypt under an encryption-context constraint, and a second grant lets a role in another AWS account decrypt -- the cross-account sharing pattern -- with a named admin role allowed to retire it.

## When to Use

- Wiring "this workload may use this key" as a first-class dependency edge in an Infra Chart, without hand-editing key policies
- Sharing encrypted data with another AWS account (the partner decrypts under a scoped, revocable grant)
- Keeping the key policy small and auditable while access grows and shrinks with deployments

## Key Configuration Choices

- **Grants over policy edits** (`grants`) — each grant is scoped, revocable, and carries no state; adding or removing one never touches the key policy
- **Referenced grantee** (`grantee_principal.valueFrom`) — resolves the workload role's ARN at deploy time; literal ARNs work for principals outside the chart (the partner account here)
- **Encryption-context constraint** (`encryption_context_subset`) — the workload grant only applies to requests carrying `app: <app-name>` context, scoping what the role can touch
- **Graceful teardown** (`retire_on_delete: true`) — the workload grant RETIRES at destroy (AWS's recommended path once a grant's work is done); the partner grant defaults to REVOKE (immediate hard stop)
- **Retiring principal** (`retiring_principal`) — the admin role may retire the partner grant without full key-admin rights

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the key lives | Your deployment region |
| `<key-description>` | Purpose of the key | Your team's naming conventions |
| `alias/shared-data-key` | Rename to your alias (e.g. `alias/myapp/shared-data`) | Must match `alias/[0-9A-Za-z_/-]+` |
| `<workload-role-name>` | The AwsIamRole resource name in your chart | Your chart's role definition |
| `<app-name>` | Encryption-context value the workload sends | Your application's KMS call sites |
| `<partner-account-id>` / `<partner-role-name>` | The external principal allowed to decrypt | The partner's account details |
| `<your-account-id>` / `<admin-role-name>` | The role allowed to retire the partner grant | Your account's admin role |

## Related Presets

- `01-symmetric-encryption` — the same key shape without delegation
- `03-external-key-store` — key material held outside AWS KMS
