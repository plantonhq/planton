# DigitalOcean SSH Key -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Key changes replace, and droplets do not follow

`public_key` is create-only: any material change (even a new comment at the end of the line) replaces the key, minting a NEW numeric id and fingerprint. Droplets created with the old key keep it in their `authorized_keys` -- replacement does not rotate access on running droplets, it only changes what NEW droplets get. Real rotation means replacing the key here AND rebuilding (or manually re-keying) the droplets that trusted the old one.

## Keys inject at droplet create time only

DigitalOcean copies the key into a droplet exactly once, at creation. Adding a key to the account never touches existing droplets; referencing it from a droplet manifest only matters for droplets created after that. Plan fleets accordingly: register keys FIRST, then create droplets that reference them.

## One key per purpose, not per person-forever

A CI deploy key, an operations break-glass key, a per-team key -- each with its own lifecycle -- beats one long-lived key everyone shares. Deleting a key is instant and safe for running droplets (see above), so pruning unused keys costs nothing.

## Prefer ed25519

DigitalOcean accepts any OpenSSH key type and enforces no algorithm floor. Use `ssh-ed25519` for new keys -- small, fast, and modern; reserve RSA for legacy clients that cannot speak it.

## Trailing whitespace is forgiven; nothing else is

The provider trims leading/trailing whitespace before comparing (so `file("~/.ssh/id_ed25519.pub")` with its trailing newline converges), but internal differences are real changes. Paste the key as one exact line.

## One key body per account -- the material is the identity, not the name

DigitalOcean deduplicates SSH keys on the public key itself: registering the same material a second time, under any name, fails with `SSH Key is already in use on your account`. Two manifests that embed one key body cannot coexist on one account, and no naming scheme rescues that, because key material is not a name. Generate a distinct pair per key resource; if two teams genuinely share a key, register it once and reference that one resource from every consumer.

## What is deliberately NOT here

Private keys (never accepted, never stored); per-droplet key management (droplets declare their own `sshKeys` lists); and fingerprint-based import (the provider's read requires the numeric id -- the import map says exactly where to find it).
