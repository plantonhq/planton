# Hetzner Cloud SSH Key — Research Documentation

## Introduction

SSH keys are the standard authentication mechanism for Linux servers on Hetzner Cloud. Every server created through the Hetzner Cloud API accepts a list of SSH key IDs; the corresponding public keys are injected into the server's `authorized_keys` file at boot time. Without a registered SSH key, the only alternative is root password authentication — a practice universally discouraged in production.

The `HetznerCloudSshKey` component registers a single SSH public key in the Hetzner Cloud account. The key is a **foundation resource**: it has no dependencies, but it is referenced by `HetznerCloudServer` (and future infra charts) via `ssh_key_ids`. Getting SSH keys right is a prerequisite for every compute workflow on Hetzner Cloud.

OpenMCF exposes exactly one spec field — `publicKey` — because the SSH key resource has exactly one user-controlled attribute. The name comes from `metadata.name`, labels are computed from metadata, and everything else (fingerprint, numeric ID) is a computed output. This is one of the simplest possible OpenMCF components, and that simplicity is intentional.

## Historical Context

SSH key management on cloud platforms has followed a consistent pattern across providers:

**Manual era:** Operators copy-paste public keys into a web console. Keys accumulate without audit trails. When an engineer leaves the team, nobody remembers which keys to revoke. Servers launched months later still inject stale keys because nobody cleaned up the account-level key list.

**Script era:** Teams write shell scripts that call `hcloud ssh-key create` or equivalent CLI commands. The scripts live in a wiki or a shared repo. They run once and are forgotten. There is no state tracking — if someone deletes a key through the console, the script doesn't know.

**IaC era:** Terraform and Pulumi bring state tracking and drift detection to SSH key management. Keys are declared in code, reviewed in pull requests, and applied through CI pipelines. This is a significant improvement, but each team still writes their own module with their own naming conventions, label schemas, and output structures.

**OpenMCF approach:** A standardized manifest format that works across both Pulumi and Terraform backends. The key resource is declared once, outputs are referenced by downstream components through `StringValueOrRef`, and the entire lifecycle — creation, update (name/labels), replacement (key material change), deletion — is handled through the same `openmcf apply` workflow.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Security > SSH Keys**
3. Click **Add SSH Key**
4. Paste the public key content
5. Enter a name
6. Click **Add SSH Key**

**Pros:**
- Zero tooling required
- Immediate visual confirmation

**Cons:**
- No audit trail beyond Hetzner's internal logs
- No version control — impossible to review changes
- Cannot be automated or reproduced
- Keys accumulate; no systematic cleanup process
- No labeling or organizational metadata

**Verdict:** Acceptable for personal projects. Not suitable for any team or production environment.

### Level 1: CLI (`hcloud`)

The Hetzner Cloud CLI provides direct SSH key management:

```bash
# Create from a key string
hcloud ssh-key create --name deploy-key \
  --public-key "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... deploy@ci"

# Create from a file
hcloud ssh-key create --name deploy-key \
  --public-key-from-file ~/.ssh/id_ed25519.pub

# List keys
hcloud ssh-key list

# Update name
hcloud ssh-key update deploy-key --name new-name

# Add labels
hcloud ssh-key add-label deploy-key env=production

# Delete
hcloud ssh-key delete deploy-key
```

**Pros:**
- Scriptable
- Full access to all attributes (name, labels)
- Fast for ad-hoc operations

**Cons:**
- No state tracking — cannot detect drift
- No dependency awareness
- Shell scripts are fragile across environments
- No structured output referencing for downstream resources

**Verdict:** Good for quick operations and debugging. Not a management solution for production infrastructure.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides the `hcloud_ssh_key` resource:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_ssh_key" "deploy" {
  name       = "deploy-key"
  public_key = file("~/.ssh/id_ed25519.pub")
  labels = {
    environment = "production"
    team        = "platform"
  }
}

output "ssh_key_id" {
  value = hcloud_ssh_key.deploy.id
}

output "fingerprint" {
  value = hcloud_ssh_key.deploy.fingerprint
}
```

**Attributes:**
- `name` (required) — Display name in Hetzner Cloud
- `public_key` (required, forces replacement) — Key material in OpenSSH format
- `labels` (optional) — Key-value metadata map

**Computed:**
- `id` — Hetzner Cloud numeric ID
- `fingerprint` — MD5 hash of the public key

**Behavior:**
- Changing `public_key` triggers resource replacement (destroy + create) because the Hetzner Cloud API does not support in-place key material updates
- Changing `name` or `labels` triggers an in-place update

**Pros:**
- State tracking and drift detection
- Plan/apply workflow for safe changes
- Integrates with remote backends for team collaboration
- Output referencing for server configurations

**Cons:**
- Requires HCL knowledge
- State management overhead for simple resources
- No built-in organizational conventions

**Verdict:** Production-grade for teams already using Terraform. The standard choice before OpenMCF.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `hcloud.SshKey`:

```go
package main

import (
    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        key, err := hcloud.NewSshKey(ctx, "deploy-key", &hcloud.SshKeyArgs{
            Name:      pulumi.String("deploy-key"),
            PublicKey: pulumi.String("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... deploy@ci"),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("sshKeyId", key.ID())
        ctx.Export("fingerprint", key.Fingerprint)
        return nil
    })
}
```

**Pros:**
- Full programming language (Go, TypeScript, Python)
- Type safety catches errors at compile time
- Built-in secret management
- Reusable components via packages

**Cons:**
- More verbose than HCL for a single resource
- Requires programming skills
- Smaller community for Hetzner Cloud specifically

**Verdict:** Excellent for Go/TypeScript teams. OpenMCF uses Pulumi (Go) internally for its IaC modules.

## Comparative Analysis

| Method | State Tracking | Drift Detection | Team Collaboration | Audit Trail | Automation |
|--------|---------------|-----------------|-------------------|-------------|------------|
| Console | No | No | No | Minimal | No |
| CLI | No | No | No | No | Partial |
| Terraform | Yes | Yes | Via remote state | Via VCS | Yes |
| Pulumi | Yes | Yes | Via backend | Via VCS | Yes |
| **OpenMCF** | **Yes** | **Yes** | **Via backend** | **Via VCS** | **Yes** |

The key differentiator for OpenMCF is not the state management (Terraform and Pulumi handle that) — it is the **standardized manifest format** and **output referencing** that makes SSH keys composable with servers, infra charts, and other components without writing custom glue code.

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: deploy-key
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... deploy@ci"
```

### What OpenMCF Automates

1. **Naming:** The SSH key name in Hetzner Cloud is derived from `metadata.name` — no separate `name` field in the spec
2. **Labeling:** Standard labels (`resource`, `resource_name`, `resource_kind`, `org`, `env`, `resource_id`) are computed from metadata and merged with user-specified labels
3. **Provider configuration:** Hetzner Cloud API token is resolved from provider config or environment variables, not hardcoded
4. **Dual IaC:** The same manifest drives both Pulumi and Terraform backends
5. **Output referencing:** The `ssh_key_id` output feeds into `HetznerCloudServer.spec.sshKeyIds` via `StringValueOrRef`, enabling declarative composition

### The 80/20 Principle

The Hetzner Cloud SSH key API has 3 user-controllable attributes: `name`, `public_key`, and `labels`. OpenMCF's `HetznerCloudSshKeySpec` exposes 1 field: `publicKey`.

**Included:**
- `publicKey` — The SSH public key content. This is the only attribute that varies per key.

**Handled by the platform:**
- `name` — Derived from `metadata.name`. Consistent naming across all components.
- `labels` — Computed from metadata (org, env, kind, id) with user labels merged in. Consistent labeling across all Hetzner Cloud resources.

This is not a reduction in capability — it is a shift of responsibility. The user still controls all three attributes, but `name` and `labels` are managed through the metadata system that is consistent across every OpenMCF component.

### API Design Decisions

**Single field spec:** `HetznerCloudSshKeySpec` contains only `publicKey`. Adding a separate `name` field would create ambiguity with `metadata.name`. Adding a `labels` field would conflict with `metadata.labels`. The spec contains only what is unique to this resource type.

**String validation (`min_len = 1`):** The proto enforces a non-empty string. Deeper validation (key format, algorithm, bit length) is delegated to the Hetzner Cloud API, which returns clear error messages for malformed keys. Duplicating format validation in proto would create a maintenance burden and risk rejecting valid keys that newer OpenSSH versions support.

**Force replacement on key change:** Both the Terraform and Pulumi providers enforce resource replacement when `public_key` changes because the Hetzner Cloud API does not support in-place key updates. This behavior is documented in the spec proto comments rather than enforced in proto validation — it is a provider-level constraint, not a schema-level one.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Description |
|------------|----------|-------|-------------|
| Pulumi | `hcloud.SshKey` | 1 | SSH public key registered in Hetzner Cloud |
| Terraform | `hcloud_ssh_key` | 1 | SSH public key registered in Hetzner Cloud |

This is a single-resource component. The SSH key has no sub-resources, attachments, or optional companions.

### Dependency Role

`HetznerCloudSshKey` is a **root resource** — it has no foreign key dependencies. It is referenced by:

- `HetznerCloudServer.spec.sshKeyIds` — Servers inject these keys at boot time

In infra charts, the pattern is:

```
HetznerCloudSshKey (foundation)
  └── ssh_key_id output
        └── HetznerCloudServer.spec.sshKeyIds (via StringValueOrRef)
```

### Label Management

Both IaC modules apply a standard label set to the Hetzner Cloud SSH key resource:

| Label Key | Source | Example |
|-----------|--------|---------|
| `planton-ai_resource` | Constant | `"true"` |
| `planton-ai_name` | `metadata.name` | `"deploy-key"` |
| `planton-ai_kind` | Constant | `"HetznerCloudSshKey"` |
| `planton-ai_org` | `metadata.org` | `"my-org"` |
| `planton-ai_env` | `metadata.env` | `"production"` |
| `planton-ai_id` | `metadata.id` | `"hcssh-abc123"` |

User-specified `metadata.labels` are merged in, with standard labels taking precedence in case of key conflicts.

## Production Best Practices

### Key Algorithm Selection

- **ED25519** (recommended): Shorter keys, faster operations, strong security. Supported by all modern OpenSSH versions (6.5+) and Hetzner Cloud.
- **RSA (>= 2048 bits)**: Broader compatibility with legacy systems. Use 4096 bits for long-lived keys. Required if connecting from systems that don't support ED25519.
- **ECDSA**: Supported but offers no practical advantage over ED25519. Some organizations avoid it due to concerns about NIST curve provenance.

**Recommendation:** Use ED25519 unless you have a specific compatibility requirement.

### Key Rotation

Hetzner Cloud SSH keys are injected at server creation time, not continuously synced. This has implications for rotation:

1. Registering a new key does not affect running servers — they already have the old key in `authorized_keys`
2. Replacing an SSH key resource (changing `publicKey`) creates a new key with a new ID. Servers referencing the old ID are unaffected until recreated.
3. To rotate keys on running servers, use configuration management (Ansible, cloud-init on reboot) to update `authorized_keys` directly.

**Practical rotation strategy:**
- Create the new SSH key resource (new manifest or updated `publicKey`)
- Update server manifests to reference the new key ID
- Redeploy servers (or update `authorized_keys` via configuration management)
- Delete the old SSH key resource

### Team Key Management

For teams, register one SSH key per person or per role rather than sharing a single key:

```yaml
# ops-team-alice.yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: ops-alice
  org: my-org
  env: production
spec:
  publicKey: "ssh-ed25519 AAAAC3... alice@company.com"
```

```yaml
# ops-team-bob.yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: ops-bob
  org: my-org
  env: production
spec:
  publicKey: "ssh-ed25519 AAAAC3... bob@company.com"
```

Then reference both keys in the server manifest. When Alice leaves, delete her SSH key manifest and redeploy.

### Security Considerations

- **Never commit private keys** to version control. Only public keys belong in manifests.
- **Use dedicated deploy keys** for CI/CD pipelines — do not reuse personal keys.
- **Label keys by purpose** (via `metadata.labels`) so it is clear which keys are for humans vs automation.
- **Audit registered keys regularly.** Hetzner Cloud does not expire SSH keys; stale keys remain until explicitly deleted.

## References

- [Hetzner Cloud SSH Keys Documentation](https://docs.hetzner.cloud/#ssh-keys)
- [Hetzner Cloud API — SSH Keys](https://docs.hetzner.cloud/#ssh-keys-get-all-ssh-keys)
- [Terraform hcloud_ssh_key Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/ssh_key)
- [Pulumi hcloud.SshKey Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/sshkey/)
- [OpenSSH Key Types](https://www.openssh.com/manual.html)
