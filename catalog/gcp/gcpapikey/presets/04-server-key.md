# Server Key

A key for a backend that calls a Google API from known addresses:
presentable only from the listed IPs, only against one service and one
method family, and protected from accidental destroy.

## What it configures

- `serverKeyRestrictions.allowedIps` — the caller's egress addresses or
  CIDR blocks.
- `apiTargets` — one service (`translate.googleapis.com`) narrowed to one
  method family. Method-level narrowing is what makes a server key safe
  to hold.
- `deletionPolicy: PREVENT` — a destroy fails rather than breaking a
  running backend.

## Adjust before deploying

- **`allowedIps`** — your NAT or load-balancer egress; a `0.0.0.0/0` here is
  no restriction at all.
- **`apiTargets`** — the service and methods your backend actually calls.
- Consider `serviceAccountEmail` (a service-account-bound key) if the API
  accepts it — requests then authenticate as that account, so its IAM
  governs what the key may do.

## When to choose something else

A backend running on Google Cloud should use Workload Identity and no key
at all — a key is a static bearer credential. Client platforms take the
Firebase Android / iOS or Browser presets.
