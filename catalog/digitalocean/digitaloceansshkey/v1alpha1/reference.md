# DigitalOceanSshKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanSshKeySpec models the full digitalocean_ssh_key resource
surface: an SSH public key registered on the DigitalOcean account, ready
to be injected into droplets (and droplet autoscale pools) at create time.

Only the name can change in place. The key material is create-only:
DigitalOcean compares it after trimming only LEADING and TRAILING
whitespace, so any change inside the line -- a different key comment, a
re-encoded body, even casing -- REPLACES the key, which rotates the
numeric id and the fingerprint that droplets reference. Deleting a key
never touches droplets that were created with it; they keep the key in
their authorized_keys.

The material is also the key's identity on the account: DigitalOcean
deduplicates on the public key, not the name, so registering the same
material twice -- under any name -- fails at create ("SSH Key is already
in use on your account"). Register shared material once and reference it.

## Example

```yaml
# Reference manifests for DigitalOceanSshKey -- protovalidate-valid,
# embedded as the reference page's Example block, and the document the
# offline tofu plan renders. The key below is a throwaway ed25519 public
# key generated for documentation; public keys are not secrets.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSshKey
metadata:
  name: ci-deploy-key
spec:
  keyName: ci-deploy-key
  publicKey: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB5K1XlWQxr9nMytXvvFyzYZaFVNSTAmTUYbSGXPqIQd ci@example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.keyName` | `string` | yes |  |  |
| `spec.publicKey` | `string` | yes |  |  |

## Field Details

### spec.keyName

`string` · required

Human-friendly name of the key, shown in the DigitalOcean console.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.publicKey

`string` · required

The SSH public key material in OpenSSH single-line format (for example
"ssh-ed25519 AAAA... comment"). DigitalOcean validates the material
server-side at create time; a malformed key fails the apply, never the
validation here. Create-only: any change replaces the key.

- rule: {"required":true,"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanSshKey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ssh_key_id` | `string` | Numeric id of the SSH key, as a string (for example "263654"). This is the key's API identity and the ONLY id imports accept -- the fingerprint does not work as an import id even though droplets accept it as a key reference. |
| `status.outputs.fingerprint` | `string` | MD5 fingerprint of the key in colon-separated hex form, computed by DigitalOcean from the key material (never derived locally). Droplets accept it interchangeably with the numeric id in their ssh_keys list. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDroplet | `spec.sshKeys` | `status.outputs.ssh_key_id` |
| DigitalOceanDropletAutoscalePool | `spec.dropletTemplate.sshKeys` | `status.outputs.ssh_key_id` |

## See Also

- [Overview](../README.md)
