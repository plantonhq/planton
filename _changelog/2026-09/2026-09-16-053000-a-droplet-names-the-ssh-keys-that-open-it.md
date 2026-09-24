# A droplet names the SSH keys that open it

## What changed

- **`DigitalOceanDropletSpec.ssh_keys` is a list of typed references to `DigitalOceanSshKey`** (default wiring: the key's numeric `ssh_key_id` output; the `fingerprint` output works too, since droplets accept either), returning at a new field number with the old one reserved. The autoscale pool's template has carried this shape since the SSH-key kind was forged; a single droplet still took raw IDs and fingerprints.
- `repeated.unique` cannot apply to a list of messages, so uniqueness is a rule over the literal values (`ssh_keys_unique`), with a message that says what DigitalOcean does with a duplicate. An empty literal is refused by the reference message's own rule, as before.
- The Pulumi module reads each literal through the reference; the preset, the catalog page's manifest, and the e2e manifest carry their keys under `value:`; the preset explainer, the catalog page, and the guide teach the reference; the e2e profile's note on `ssh_keys` says a key fixture can now be composed.

## Why

An SSH key is the standard access path to a droplet, and the catalog has a kind for it. With the droplet taking raw strings, a key drew a line to the fleet that injects it and none to a droplet that does, and the droplet's wizard offered no picker.

## How to check

```bash
go test ./catalog/digitalocean/digitaloceandroplet/v1alpha1/      # duplicates refused, a literal beside a reference accepted
go build ./catalog/digitalocean/digitaloceandroplet/iac/pulumi/module/
grep -n 'reserved 15;' catalog/digitalocean/digitaloceandroplet/v1alpha1/spec.proto
```
