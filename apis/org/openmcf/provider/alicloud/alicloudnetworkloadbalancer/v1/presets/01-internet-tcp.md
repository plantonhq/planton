# Preset: Internet-Facing TCP

## When to Use

- You need a public-facing L4 load balancer for TCP traffic.
- Simple setup: one server group, one listener, two availability zones.
- Suitable for development, staging, or simple TCP services.

## What It Creates

- Internet-facing NLB with 2-zone HA
- Single TCP server group with basic health check (TCP probe)
- Single TCP listener on port 80

## Customization Points

- Replace `<placeholders>` with actual VPC, VSwitch, and zone IDs
- Change listener port (e.g., 3306 for MySQL, 6379 for Redis)
- Add `tags` for cost tracking
- Adjust health check parameters for your backend
