# One Zone Dev EFS

One Zone storage (us-east-1a), encrypted, bursting throughput. Lower cost for dev/test. Single subnet.

## When to Use

- Development and test environments where cost matters more than availability
- Ephemeral or easily recreated data
- Single-AZ deployments (e.g., dev cluster in one subnet)
- CI runners or build caches with local backup

## What It Configures

- **One Zone** (`availabilityZoneName: us-east-1a`) — Data stored in a single AZ; change to your region's AZ (e.g., us-west-2a)
- **Single subnet** — Exactly one subnet in the specified AZ; only one mount target is created
- **Encrypted** — Encryption at rest for compliance even in dev
- **Bursting throughput** — Cost-effective for typical dev workloads
- **No backup** — Backups disabled to reduce cost; enable if dev data is valuable

## What to Customize

- Change `availabilityZoneName` to match your region (e.g., us-west-2a, eu-west-1a)
- Enable `backupEnabled: true` if dev data should be backed up
- Add access points if multiple apps share the file system
