# Standard Reserved IP

This preset provisions a static reserved IPv4 address in the London region. Reserved IPs persist independently of instances and can be reassigned, making them suitable for high-availability setups where a public IP must survive instance replacement.

## When to Use

- Production workloads that need a stable public IP address
- High-availability setups where the IP is moved between instances during failover
- Load balancers or bastion hosts that external clients address by IP

## Key Configuration Choices

- **Region** (`region: lon1`) -- the IP can only be attached to resources in the same region; change to match your target region
- **Description** (`description`) -- human-readable label for identification in the Civo dashboard

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `lon1` | Target Civo region | Civo dashboard or `civo region ls` |
| `Reserved IP for production workload` | Descriptive label | Your naming convention |
