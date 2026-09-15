# DigitalOcean Droplet -- Operational Guide

Judgment calls that matter when you run DigitalOcean droplets.

## Inject SSH keys at create, or live with a password email

`sshKeys` is the standard access path to a droplet, and it is create-only: DigitalOcean injects the keys into the image on first boot, and the provider recreates the droplet on any change to the list. A droplet created without keys gets a root password by email — fine for a throwaway, wrong for anything real. Keys must already exist on the account: reference a `DigitalOceanSshKey` Planton manages (the wiring resolves its numeric id), or pass a literal id or fingerprint (`doctl compute ssh-key list` shows both; either works). Rotating access later without recreation is an in-OS operation (edit `authorized_keys` via cloud-init or configuration management), not an API one.

## Region and VPC are safe to omit

Unset `region` lets DigitalOcean pick a region with available capacity — useful for dev boxes, wrong for anything that must sit next to other resources (volumes, private-network peers): those must pin the region. Unset `vpc` lands the droplet in the chosen region's default VPC. Either way the real landing zone comes back in the `vpc_uuid` output, so downstream wiring never has to guess.

## Resizes: the disk decision is permanent

Changing `size` powers the droplet off and resizes it. What happens to the disk is governed by `resizeDisk`, which DigitalOcean defaults ON: the disk grows to the new size's allocation and the droplet can never move to a smaller-disk size again. `resizeDisk: false` scales CPU/RAM only — fully reversible, at the cost of not using the larger disk. Decide before the first resize, not after.

## Backups: the policy needs the toggle

`enableBackups` turns automated backups on (weekly by default, 4-week retention); `backupPolicy` picks the window — `daily`, or `weekly` with a `weekday` and an `hour` on DigitalOcean's four-hour grid (0, 4, 8, 12, 16, 20). A policy without the toggle is rejected before any provisioner runs, mirroring the provider's own create-time error. Toggling backups later updates in place.

## The two agents are different things

`monitoring` is the metrics agent: enhanced graphs and monitor alert policies. `dropletAgent` is the web-console agent behind the control panel's Console button. Both are create-only. `dropletAgent` is tri-state on purpose: unset lets DigitalOcean install it where the image supports it and skip it silently otherwise; explicit `true` makes an unsupported image a hard error; explicit `false` blocks installation.

## IPv6 and public networking are one-way doors

Enabling `enableIpv6` on a running droplet updates in place; disabling it recreates the droplet. `publicNetworking: false` creates a droplet with no public interface at all — reachable only inside its VPC — and is create-only. The verifier's public-IP checks and most bootstrap flows assume a public interface; private-only droplets belong behind a load balancer or bastion you have already built.

## Tags are the targeting fabric

DigitalOcean firewalls and load balancers target droplets by tag. Both provisioners always add the standard Planton labels as `key:value` tags alongside `spec.tags`, so a fresh droplet is immediately targetable by the tags you declared, and identifiable by the ones Planton added.

## Importing an existing droplet

Import uses the bare integer droplet id (the `droplet_id` output). Expect `sshKeys`, `userData`, `gracefulShutdown`, `dropletAgent`, and `backupPolicy` to stay at their configured values after import: the API never reports them back. The importer seeds `resizeDisk` as `true` regardless of history.

## What is deliberately NOT here

The deprecated `private_networking` flag (superseded by `vpc`). The provider's `ipv6_address` argument — it is inert at the pinned provider (never sent to the API on create or update), so it exists here only as the `ipv6_address` output. Droplet autoscale pools, snapshots, reserved IPs, and firewalls are separate resources.
