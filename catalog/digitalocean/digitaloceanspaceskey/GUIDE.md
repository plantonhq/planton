# DigitalOcean Spaces Key -- Operational Guide

What experience with this component teaches that the field reference cannot.

## The secret is shown once -- design your handoff around that

`secret_key` exists in exactly one place: the create response. DigitalOcean never returns it again, from any API, ever. Both provisioners store it as a sensitive output, and consumers should read it from there (or through a reference) rather than expecting to look it up later. If the secret is lost, there is no recovery -- destroy the key and create a new one. Both provisioners are proven to capture it: the live proof signs a real S3 request with the pair each engine exported, and a freshly minted pair is accepted by the Spaces data plane within two seconds.

## Adopting an existing key cannot work -- create a new one instead

There is no import for a Spaces key: the provider ships no importer, and even one that existed could never recover the secret. Do not try to bring an existing key under management by any workaround that re-creates it -- that mints a SECOND access key while the first keeps working unmanaged, and every consumer of the old pair silently keeps old credentials. Create the key through this component, hand its pair to consumers, and retire the hand-made key on your own schedule.

## Rotation means a new key, and consumers must follow

The key material is immutable: rotating credentials is destroy-and-recreate, which changes BOTH the access key and the secret. Plan rotation as a two-step move -- create the replacement key first, re-point every consumer, then destroy the old one. Deleting a key invalidates it immediately; anything still signing with it starts failing on the spot.

## Prefer per-bucket grants; treat fullaccess as an admin credential

A `fullaccess` grant covers every bucket in the account -- including buckets created after the key. That is occasionally right (a backup operator, an admin tool) and usually wrong for workloads. The per-bucket `read` / `readwrite` grants exist precisely so each workload's key unlocks only its own data. A key with NO grants is valid and authorizes nothing -- a safe placeholder state.

## The permission wall here is the only one anywhere

The provider performs no validation on `permission`: a typo like `write` upstream becomes an EMPTY permission, producing a grant that silently authorizes nothing -- the failure appears as mysterious 403s at the workload, far from the cause. This spec rejects anything outside `read` / `readwrite` / `fullaccess` at validation time, so the class never reaches DigitalOcean.

## Grant updates replace the whole list

Updating grants sends the complete list in one call -- there is no add-one-grant API. That is invisible in normal declarative use (the spec IS the whole list), but worth knowing when diagnosing: a grant you removed from the spec is genuinely revoked on the next apply, not left behind.

## What is deliberately NOT here

Import (the resource has no upstream importer, and the write-once secret could never round-trip -- recorded exclusion in the import map); a `created_at` output (a timestamp with no operational consumer); and any Spaces credential requirement for MANAGING keys -- key management rides the account API token; the key this component creates is itself the Spaces credential.
