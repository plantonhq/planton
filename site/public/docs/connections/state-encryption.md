---
title: "State Encryption"
description: "Every state backend encrypts your infrastructure state with the engine's own encryption, under Planton's key or a key you hold: passphrase, AWS KMS, Google Cloud KMS, Azure Key Vault, or Vault/OpenBao"
icon: shield
order: 55
tags:
  - Connect
  - State Backends
  - Encryption
  - OpenTofu
  - Pulumi
---

# State Encryption

Every state backend in Planton names the key that encrypts your infrastructure state. Planton uses each engine's own encryption: OpenTofu (and the Terraform provisioner, which runs on it) encrypts the whole state file and every saved plan, and Pulumi encrypts every value your program marks secret. By default the key is Planton's, with nothing for you to configure; you can point any backend at a key you hold instead.

## Why It Matters

A state file holds the values your resources were given and produced: database passwords, API keys, private keys. Anyone who can read the file can read them. With encryption, a person or system that can read your bucket -- a leaked storage credential, a misconfigured policy, Planton's own staff -- sees only ciphertext.

Because Planton uses each engine's native encryption and never wraps a file in a format of its own, you can always open your own state with the engine's own command-line tool and your key.

Encrypted state is never read as empty. A runner without the key, or with the wrong one, stops with an error naming the problem, because state read as empty would plan to recreate your infrastructure.

## Choosing a Key

You choose the key when you create a state backend (the **Key** step of the connection wizard, or `--key-source` on `planton infra state-backend create`):

| Key source | OpenTofu | Pulumi | Who holds the key |
|---|---|---|---|
| Planton's key (default) | Yes | Yes | Planton. On Planton-managed storage, a key derived for your organization alone; on your own storage, a passphrase Planton creates in your organization's secrets |
| A passphrase you hold | Yes | Yes | You, as a secret in your organization |
| AWS KMS | Yes | Yes | Your AWS account, through an AWS connection |
| Google Cloud KMS | Yes | Yes, with the limits below | Your Google Cloud project, through a Google Cloud connection |
| Azure Key Vault | Yes, Managed HSM included | Yes, key vault keys only | Your Azure subscription, through an Azure connection |
| Vault or OpenBao transit | Yes | Yes | Your Vault or OpenBao, through a connection |
| Pulumi Cloud's keys | No | Yes, for state in Pulumi Cloud | Pulumi Cloud |
| No encryption | Yes, in your own bucket only | No | Nobody -- the state is readable |

If you choose nothing, a backend gets its type's default: Planton's key for every storage type, Pulumi Cloud's keys for state in Pulumi Cloud, and no encryption for Terraform Cloud, which must read your state to show it and run against it. A cloud key's credentials come from its own connection and stay apart from the bucket's credentials and from the cloud your resources deploy to.

## What Your Key's Identity Needs

The connection a key source names needs permission to use that one key, and nothing more. It never needs to manage the key, and it needs no access to your bucket or to the resources you deploy. The two engines call different operations, so grant the row for every engine that will use the backend:

| Key source | OpenTofu calls | Pulumi calls | What to grant |
|---|---|---|---|
| AWS KMS | `kms:GenerateDataKey`, `kms:Decrypt` | `kms:Encrypt`, `kms:Decrypt` | Those actions on the key: in its key policy, in a grant, or in an IAM policy the key policy allows |
| Google Cloud KMS | encrypt and decrypt | encrypt and decrypt | The Cloud KMS CryptoKey Encrypter/Decrypter role (`roles/cloudkms.cryptoKeyEncrypterDecrypter`) on the key |
| Azure Key Vault | encrypt and decrypt, with RSA-OAEP-256 | encrypt and decrypt, with RSA-OAEP-256 | The data actions `Microsoft.KeyVault/vaults/keys/encrypt/action` and `Microsoft.KeyVault/vaults/keys/decrypt/action` on the key (an RSA key), or an access policy with the key permissions Encrypt and Decrypt |
| Vault or OpenBao transit | `<mount>/datakey/plaintext/<key>`, `<mount>/decrypt/<key>` | `<mount>/encrypt/<key>`, `<mount>/decrypt/<key>` | A policy with `update` on those paths for the connection's token |

For a transit key mounted at `transit` and named `state`, a token both engines can use carries this policy:

```hcl
path "transit/datakey/plaintext/state" { capabilities = ["update"] }
path "transit/encrypt/state"           { capabilities = ["update"] }
path "transit/decrypt/state"           { capabilities = ["update"] }
```

A key's own versions rotate in AWS, Google Cloud, Azure, Vault or OpenBao without any change in Planton: state sealed under an older version still opens, and new state is sealed under the newest one.

## Changing the Key

A backend's key is part of the backend, like any other field. Change it with **Change Key** on the backend's page, with the CLI, or by editing the backend's manifest and applying it:

```bash
# Move a backend onto your own AWS KMS key
planton infra state-backend rekey team-s3 --key-source aws-kms \
  --aws-kms-connection acme-prod --aws-kms-key-id alias/state --aws-kms-region us-east-1

# Rotate Planton's passphrase on your own storage, or resume a change that stopped
planton infra state-backend rekey team-s3
```

Planton then re-encrypts the state of every resource on the backend under the new key. The backend's page and `planton infra state-backend get` show the progress. Deploys keep running throughout, because every job reads with either key and writes with the new one. If a change cannot finish -- a resource's state the runner cannot reach, a key your connection cannot use -- it stops and names each reason; nothing is lost, and **Resume** (or `rekey` again) picks it up once you fix the cause.

The same door turns encryption on for OpenTofu state in your own bucket, or off again. If your bucket keeps object versions, its earlier versions written before encryption stay as they were; Planton tells you and does not delete them.

Rotating a key you hold is yours: store a new passphrase in a new secret and change the backend's key to it, or let your cloud key rotate its own versions, which needs no change in Planton.

## Keys That Cannot Be Lost by Accident

- A secret that encrypts state -- a passphrase you hold, or one Planton created -- cannot be deleted, its latest version cannot be deleted, and it cannot be overwritten in place. Planton names the backends it encrypts and points you at Change Key.
- A connection whose key encrypts state cannot be deleted.
- A backend in the middle of a key change cannot be the source or the destination of a state move.

## Opening Your Own State

- **With a key you hold**, give the engine the same key: an OpenTofu encryption configuration (`TF_ENCRYPTION`) naming the key provider under the label your state file shows in its `meta` (it starts with `planton_`), or Pulumi the same secrets provider.
- **With Planton's passphrase on your own storage**, the passphrase is a secret in your own organization, shown on the backend's page.
- **With Planton's key on Planton-managed storage**, download the state from the resource's IaC tab or the CLI. Planton decrypts it for you into a plain-text copy in its transfer storage: the link works for an hour, and Planton deletes the copy about two days after it is written.

## Not Supported, and Why

Planton supports every combination each engine can do safely, and refuses the rest with a sentence, rather than offering a path that half works. The console never offers these:

- **Pulumi with a Google Cloud KMS key on a Cloud Storage backend**, unless the bucket and the key both use the runner's own Google Cloud identity (runner mode). Pulumi uses one Google Cloud identity for both, so their credentials cannot be kept apart.
- **Pulumi with a Google Cloud KMS key, deploying Google Cloud resources through a different connection.** Pulumi uses one Google Cloud identity for the key and for the resources it deploys. Point the key at the resource's own Google Cloud connection.
- **Pulumi with an Azure Key Vault key, deploying through a different Azure connection that uses the runner's identity.** Both read the same identity variables. Point the key at the resource's own Azure connection.
- **Pulumi with an Azure Managed HSM key.** Pulumi's Azure Key Vault provider accepts only keys in a key vault.
- **OpenTofu with an Azure Key Vault key that does not sign in with a client secret of its own, beside Azure credentials that would fill in for it.** OpenTofu fills anything an Azure key leaves blank from the Azure variables of the resource it deploys, and tries a client secret before a token or a managed identity. So a keyless key cannot sit beside a resource that deploys through a different Azure connection signing in with a client secret, and a key using your runner's own Azure identity cannot sit beside an Azure identity of another kind. A key whose connection signs in with a client secret works beside any deploy. The runner names the variable that would cross.
- **OpenTofu with an Azure Key Vault key on a runner of your own running OpenTofu older than 1.11.9, or 1.12.0.** Those releases sign the key in with the tenant of the resources being deployed instead of the key's own. Planton's runners ship a release that reads the key's tenant; on your own runner, use OpenTofu 1.11.9 or later (1.12.1 or later in the 1.12 line).
- **Pulumi Cloud's keys on any backend but Pulumi Cloud**, and **Pulumi with no encryption** -- Pulumi always encrypts secret values with some key.
- **Any key on Terraform Cloud or Terraform Enterprise**, which reads your state itself.
- **OpenTofu's experimental external key provider and method.** They are experimental, and on a Planton-hosted runner they would run your program on Planton's machines.

## What It Does Not Cover

- **Values Pulumi was not told are secret.** Pulumi encrypts what your program marks secret; OpenTofu encrypts the whole file.
- **Earlier versions in a versioned bucket**, written before encryption was turned on.
- **The open-source CLI used on its own**, against state you point it at: there, encryption is the engine's own setting, as described in [Deploy from GitHub Actions](/docs/ci-cd/deploy-from-github-actions).
- **Downloaded and uploaded copies.** A downloaded state file is plain text, and so is any copy of it you keep. Between you and the backend, Planton holds each copy in its own transfer storage, a download as plain text and an upload exactly as you sent it, behind links that last at most two hours, and deletes it about two days after it is written. Only Planton's storage access protects it there.
- **A deploy whose state could not be saved, on a backend you keep unencrypted.** When OpenTofu changes your infrastructure but cannot save the new state (a credential that expired mid-deploy, a network failure), Planton keeps that state until the resource's next run stores it, before anything else runs. It is sealed under the backend's key, so on a backend you keep unencrypted, it waits unencrypted too.
- **A saved state that can no longer be stored.** If that kept state no key on the backend opens, or the backend was written outside Planton since, every run on the resource stops with a sentence saying so, until you upload the state the resource should keep.
- **A lock a failed run left in your own bucket.** On Google Cloud Storage, Azure Blob Storage, or S3 with a DynamoDB lock table, a run that failed mid-deploy can leave its lock behind. Later runs stop with OpenTofu's lock message until you release it with `tofu force-unlock <lock ID>`; Planton cannot release it for you.
- **A Pulumi create that was in flight when a save failed.** Pulumi keeps every step it saved. A resource it had started creating is recorded as a pending operation: the next run warns about it, and you adopt it with an [import](/docs/infrastructure/importing-resources). Planton does not adopt it for you.

## Related Documentation

- [State Backends](/docs/connections/state-backends) -- where state is stored and how a backend is configured
- [Runner Security Model](/docs/runner/security-model) -- what a runner holds and what it never sees
- [Secrets](/docs/secrets) -- the secrets a passphrase lives in
