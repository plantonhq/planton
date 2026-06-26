# Preset: Account default virtual network

The single virtual network that routes and WARP clients fall back to when they do not
name one explicitly.

## When to use

- You want one canonical segment for the common case and only create additional virtual
  networks when you genuinely need to isolate overlapping CIDRs.

## Key choices

- `isDefaultNetwork: true`: marks this as the account default. Only one virtual network
  can hold this at a time — applying it elsewhere moves the default.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
