# Preset: Internal TCP with Connection Draining

## When to Use

- Internal (VPC-private) L4 load balancer for microservice traffic.
- Graceful deployments where in-flight connections must complete before backends are removed.
- Source-IP consistent hashing for session affinity without cookies.

## What It Creates

- Intranet NLB (no public access)
- Server group with connection draining (60s timeout) and Sch (source-IP hash) scheduling
- TCP listener on port 8080 with Proxy Protocol for real client IP forwarding
- 600s idle timeout for long-lived connections

## Customization Points

- Replace `<placeholders>` with actual resource references
- Adjust `connectionDrainTimeout` (10-900s) based on your longest request lifecycle
- Change scheduler to `Tch` (four-tuple hash) for finer-grained distribution
- Disable `proxyProtocolEnabled` if backends don't support Proxy Protocol
