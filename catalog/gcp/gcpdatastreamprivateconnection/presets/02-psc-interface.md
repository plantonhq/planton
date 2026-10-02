# PSC Interface

## Use Case

Connect Datastream into your VPC through a Private Service Connect interface instead of peering -- no reserved range, no peering to manage, and it reaches what the attachment's subnet routes to.

## When to Use

- VPCs where every range is spoken for, or peering is not allowed
- Networks where a PSC interface is the standard way services connect

## What This Creates

- A private connection in `us-central1` on the `datastream` network attachment

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pscInterfaceConfig.networkAttachment` | `.../networkAttachments/datastream` | Your attachment in the same region; it must accept connections from Datastream's service project. |
