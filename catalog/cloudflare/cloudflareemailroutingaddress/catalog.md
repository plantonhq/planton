# Cloudflare Email Routing Address

Registers a verified destination address for Cloudflare Email Routing -- an account-scoped mailbox that routing rules and zone catch-alls forward to. Creating it sends a verification email to the mailbox; the address cannot receive forwarded mail until its owner clicks the link, and Cloudflare rejects any rule that forwards to an unverified address. Because addresses are account-scoped, you register a teammate's inbox once and reference it from any routing rule or zone catch-all in the account.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Destination Address** -- an account-scoped verified mailbox usable as a forwarding target
- **Verification Email** -- Cloudflare emails the address a confirmation link on creation; forwarding stays inert until it is verified

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Email Routing edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Account-level access** -- destination addresses are created at the account level, so the API token must be scoped to the account.
- **Mailbox access** -- the owner of the destination mailbox must be able to click the verification link Cloudflare sends. This is a human step by design; there is no API shortcut around it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Email Routing Address**, and click **Deploy**. The creation wizard captures the owning account and the destination email. Both are fixed at creation -- changing either replaces the address. Start from the **Destination Address** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingAddress
metadata:
  name: ops-mailbox
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  email: ops@example.com
```

```shell
planton apply -f cloudflare-email-routing-address.yaml
```

This registers `ops@example.com` as a destination. An Infra Job tracks the provisioning in real time, and Cloudflare emails the mailbox a verification link.

## Key Configuration

These are the most important decisions when configuring a destination address. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Verification gates everything downstream** -- until the mailbox owner clicks the emailed link, the `verified` output stays empty and no rule or catch-all can forward to the address (Cloudflare rejects the configuration). Sequence rollouts as: create addresses, have owners verify, then wire rules.

**Destination Email (`email`)** -- the real mailbox forwarded mail is delivered to. Immutable -- changing it replaces the address, and the new mailbox must be re-verified. Cloudflare enforces uniqueness per account, so registering the same email as a second resource in one account fails on create: model each real mailbox once and share it.

**Account (`accountId`)** -- the owning Cloudflare account. Immutable, and must match the account of the rules and zones that forward here.

**Verification Override (`status`)** -- normally leave empty. It exists for one real operation: flipping a verified address back to `unverified` (e.g. after a mailbox changes hands) to cut forwarding without deleting the address. Setting `verified` requires account admin privileges -- it is not a verification shortcut. The Pulumi engine cannot send this field yet; use the Terraform engine when the override matters.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- a destination address is a leaf resource identified by the `accountId` string and the literal email.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `email` | The destination email address | Referenced by a `CloudflareEmailRoutingRule` forward action or a `CloudflareEmailRoutingZone` catch-all (`forwardTo`) |
| `verified` | RFC3339 timestamp once verified, empty until then | Gating rule deployment on forwarding readiness |
| `address_id` | The Cloudflare-assigned identifier of the address | Verification tooling and imports |

## Common Patterns

**Shared team inbox** -- register `support@acme.com` once, then reference it from routing rules across every zone in the account; account-level scoping makes the single registration reusable. Start from the **Destination Address** preset.

**Personal forwarding target** -- register a personal mailbox, verify it, then point a custom-domain alias at it through a routing rule.

**Verify-first rollout** -- in a chart that also deploys rules, create the address in an early wave and hold the rules until `verified` is non-empty; wiring both in one shot fails when the human step has not happened yet.

## Works With

- [**Cloudflare Email Routing Rule**](/infra-catalog/cloudflare-email-routing-rule) -- forwards matched mail to this address
- [**Cloudflare Email Routing Zone**](/infra-catalog/cloudflare-email-routing-zone) -- a forwarding catch-all delivers unmatched mail to this address
