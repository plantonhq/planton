# HSM-Protected Symmetric Key

The compliance-grade CMEK key: identical to the symmetric workhorse but
with every key version generated and held inside Cloud HSM
(FIPS 140-2 Level 3 validated hardware).

## What this preset creates

A symmetric `ENCRYPT_DECRYPT` key whose version template pins
`protectionLevel: HSM`. Protection level is immutable — an HSM key can
never be quietly downgraded to software protection, which is exactly the
guarantee auditors ask for.

## Prerequisites

- A `GcpKmsKeyRing` named `compliance-keys` in the location where your
  data lives (see the `GcpKmsKeyRing` presets).

## Cost note

HSM key versions carry a higher per-version monthly price than software
versions, and rotation accumulates versions until old ones are destroyed
— factor the rotation cadence into cost planning.

## Remix ideas

- Drop `rotationPeriod` for a manually rotated HSM key.
- Use this template for `MAC` or asymmetric purposes by switching
  `purpose` and the algorithm together and dropping `rotationPeriod`:
  Cloud KMS rotates only symmetric encryption keys automatically, and the
  spec refuses a rotation period on any other purpose.
