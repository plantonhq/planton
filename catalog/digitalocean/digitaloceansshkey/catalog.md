# DigitalOcean SSH Key

Registers an SSH public key on the DigitalOcean account, ready to be injected into droplets and droplet autoscale pools at create time. Only the public half is stored -- the private key never leaves your machine. The key material is create-only: any change inside the line replaces the key, minting a new numeric id and fingerprint, while droplets created with the old key keep it in their `authorized_keys` and never follow the rotation.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **SSH Key** -- the named public key on the account, with a DigitalOcean-computed fingerprint (never derived locally)

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **The public key material** -- `publicKey` takes the key in OpenSSH single-line format. DigitalOcean validates it server-side at create time, so a malformed key fails the apply, never the manifest validation. Nothing else: SSH keys are free account metadata.

## Deploy

### Console

Open the deployment store, find **DigitalOcean SSH Key**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CI Deploy Key** preset in the [Presets](#presets) tab to give a CI pipeline its own credential instead of borrowing a person's.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSshKey
metadata:
  name: ops-key
  org: acme-corp
  env: prod
spec:
  keyName: ops-key
  publicKey: ssh-ed25519 your-public-key-material ops@acme-corp.com
```

```shell
planton apply -f do-ssh-key.yaml
```

This registers the named public key on the account (paste your real key as one exact line), and DigitalOcean computes its numeric id and fingerprint. A Stack Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an SSH key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`publicKey` is create-only -- changes replace, and droplets do not follow** -- Any material change to the line, even a new comment at the end, replaces the key and rotates both the numeric id and the fingerprint. Replacement never touches running droplets: they keep the old key in `authorized_keys`. Real rotation means replacing the key here AND rebuilding (or manually re-keying) every droplet that trusted the old one.

**Keys inject at droplet create time only** -- DigitalOcean copies the key into a droplet exactly once, at creation. Registering a key never touches existing droplets. Order matters: register keys first, then create the droplets and pools that reference them.

**One key per purpose** -- A CI deploy key, an operations break-glass key, a per-team key -- each with its own lifecycle -- beats one long-lived shared key. Deleting a key is instant and safe for running droplets (see above), so pruning unused keys costs nothing.

**Prefer ed25519** -- DigitalOcean accepts any OpenSSH key type and enforces no algorithm floor. Use `ssh-ed25519` for new keys; reserve RSA for legacy clients that cannot speak it.

**Paste one exact line** -- DigitalOcean trims only leading and trailing whitespace before comparing, so a trailing newline from `file("~/.ssh/id_ed25519.pub")` converges cleanly -- but any difference inside the line (a re-encoded body, a changed comment, casing) is a real change that replaces the key.

**`keyName` renames in place** -- The display name is the only field that updates without replacement.

**One key body per account** -- DigitalOcean deduplicates on the public key material, not the name: a second resource embedding the same key fails at create with `SSH Key is already in use on your account`, whatever it is called. Register shared material once and reference that resource from every droplet or pool that needs it.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies -- both spec fields are literal values.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ssh_key_id` | Numeric id of the key, as a string -- the key's API identity and the only id imports accept | Droplet and autoscale pool `sshKeys` lists |
| `fingerprint` | MD5 fingerprint in colon-separated hex, computed by DigitalOcean | Accepted interchangeably with the id in droplet `sshKeys` lists; key verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CI deploy key** -- a dedicated key for a pipeline, so automation reaches droplets with its own credential and a person's departure never breaks deploys. The private half stays in the CI system's secret store. Start from the **CI Deploy Key** preset.

**Break-glass operations key** -- an emergency-access key kept separate from day-to-day automation keys, so incident access survives a CI credential rotation (and vice versa). Ideally the private half lives in a hardware token or a vault. Start from the **Operator Break-Glass Key** preset.

## Works With

- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- droplets reference this key's `ssh_key_id` in their `sshKeys` lists at create time
- [**DigitalOcean Droplet Autoscale Pool**](/cloud-catalog/digital-ocean-droplet-autoscale-pool) -- pool templates require at least one key; every member is born with it installed
