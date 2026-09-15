# DigitalOcean Droplet -- Operational Guide

Judgment calls that matter when you run DigitalOcean droplets.

## Inject SSH keys at create, or live with a password email

`sshKeys` is the standard access path to a droplet, and it is create-only: DigitalOcean injects the keys into the image on first boot and never reports them back. A droplet created without keys gets a root password by email — fine for a throwaway, wrong for anything real. Keys must already exist on the account (`doctl compute ssh-key list` shows IDs and fingerprints; either works). Editing the list later does nothing: both provisioners deliberately ignore changes to `sshKeys` (see "Create-only fields are ignored after creation" below). Rotating access is an in-OS operation (edit `authorized_keys` via cloud-init or configuration management), not an API one.

## Create-only fields are ignored after creation, by design

`sshKeys`, `userData`, and `dropletAgent` only have meaning at first boot, and the DigitalOcean API never reports any of them back. The Terraform provider's answer to a change in any of them is to destroy the droplet and create a new one — with a new disk and a new IP. Both Planton provisioners refuse that trade: after creation they ignore edits to these three fields, so a manifest change to `userData` is a no-op rather than a silent server replacement. Two consequences: to run new cloud-init or inject different keys, replace the droplet deliberately (a new manifest, or delete and re-create); and adopting an existing droplet is safe even when the manifest describes those fields, because the first apply after import plans no replacement.

## Region and VPC are safe to omit

Unset `region` lets DigitalOcean pick a region with available capacity — useful for dev boxes, wrong for anything that must sit next to other resources (volumes, private-network peers): those must pin the region. Unset `vpc` lands the droplet in the chosen region's default VPC. Either way the real landing zone comes back in the `vpc_uuid` output, so downstream wiring never has to guess.

## Resizes: the disk decision is permanent

Changing `size` powers the droplet off and resizes it. What happens to the disk is governed by `resizeDisk`, which DigitalOcean defaults ON: the disk grows to the new size's allocation and the droplet can never move to a smaller-disk size again. `resizeDisk: false` scales CPU/RAM only — fully reversible, at the cost of not using the larger disk. Decide before the first resize, not after.

## Backups: the policy needs the toggle

`enableBackups` turns automated backups on (weekly by default, 4-week retention); `backupPolicy` picks the window — `daily`, or `weekly` with a `weekday` and an `hour` on DigitalOcean's four-hour grid (0, 4, 8, 12, 16, 20). A policy without the toggle is rejected before any provisioner runs, mirroring the provider's own create-time error. Toggling backups later updates in place.

Expect a permanent, harmless diff while backups are on. The provider at the pinned version decides whether backups are enabled by looking for `"backups"` in the droplet's `features` list, and DigitalOcean's current backup system no longer reports it there — the truth lives at `GET /v2/droplets/{id}/backups/policy`, which shows `backup_enabled: true` and the applied window. So backups ARE on, but every plan proposes `backups: false → true` and every apply sends an enable action DigitalOcean treats as a no-op. Verify your backups from the policy endpoint or the control panel, not from the plan. Tracked upstream as [digitalocean/terraform-provider-digitalocean#1525](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1525); the noise disappears with a provider release that reads backup state from the policy endpoint.

## The two agents are different things

`monitoring` is the metrics agent: enhanced graphs and monitor alert policies. `dropletAgent` is the web-console agent behind the control panel's Console button. Both are create-only; `monitoring` is reported back by the API (so a change to it plans a recreate), while `dropletAgent` is not (so later edits are ignored — see above). `dropletAgent` is tri-state on purpose: unset lets DigitalOcean install it where the image supports it and skip it silently otherwise; explicit `true` makes an unsupported image a hard error; explicit `false` blocks installation.

## IPv6 and public networking are one-way doors

Enabling `enableIpv6` on a running droplet updates in place; disabling it recreates the droplet. `publicNetworking: false` creates a droplet with no public interface at all — reachable only inside its VPC — and is create-only. The verifier's public-IP checks and most bootstrap flows assume a public interface; private-only droplets belong behind a load balancer or bastion you have already built.

## Tags are the targeting fabric

DigitalOcean firewalls and load balancers target droplets by tag. Both provisioners always add the standard Planton labels as `key:value` tags alongside `spec.tags`, so a fresh droplet is immediately targetable by the tags you declared, and identifiable by the ones Planton added.

## Importing an existing droplet

Import uses the bare integer droplet id (the `droplet_id` output). The API never reports `sshKeys`, `userData`, `dropletAgent`, `gracefulShutdown`, or `backupPolicy` back, so a fresh import leaves them empty in state. The first three are ignored after creation by both provisioners, so the first apply after import plans no replacement — adopting a running droplet is safe by design. `gracefulShutdown` and `backupPolicy` are re-applied in place from the manifest. The importer seeds `resizeDisk` as `true` regardless of history, so a manifest with `resizeDisk: false` shows a one-time in-place update.

## What is deliberately NOT here

The deprecated `private_networking` flag (superseded by `vpc`). The provider's `ipv6_address` argument — it is inert at the pinned provider (never sent to the API on create or update), so it exists here only as the `ipv6_address` output. Droplet autoscale pools, snapshots, reserved IPs, and firewalls are separate resources.
