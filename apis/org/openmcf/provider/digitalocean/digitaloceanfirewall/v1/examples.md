# DigitalOcean Firewall Examples

Complete, copy-paste ready YAML manifests for common firewall configurations.

---

## Example 1: Development Web Server (Permissive)

**Use Case**: Simple firewall for development with open SSH, HTTP, HTTPS, and unrestricted outbound.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: dev-web-firewall
spec:
  name: dev-web-firewall
  tags:
    - dev-web
  inbound_rules:
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      port_range: "80"
      source_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      port_range: "443"
      source_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: icmp
      source_addresses:
        - "0.0.0.0/0"
        - "::/0"
  outbound_rules:
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- **NEVER use in production** - SSH and web ports open to the world
- Suitable only for throwaway dev environments
- Allow all outbound traffic for developer flexibility
- Includes ICMP for ping/traceroute debugging

**Production Equivalent**: See Example 3 (Production Web Tier)

---

## Example 2: Management Firewall (Applied to All Instances)

**Use Case**: Centralized SSH access and monitoring for all servers in an organization.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: management-firewall
spec:
  name: management-firewall
  tags:
    - all-instances
  inbound_rules:
    # SSH from office bastion host only
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
    # Prometheus node exporter
    - protocol: tcp
      port_range: "9100"
      source_tags:
        - prometheus-server
  outbound_rules:
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    # HTTPS for OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    # NTP
    - protocol: udp
      port_range: "123"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- Apply to **all** Droplets via shared tag (`all-instances`)
- SSH locked to specific bastion IP (never `0.0.0.0/0`)
- Monitoring access via tag-based targeting (scales with fleet)
- Allows essential OS services: DNS, HTTPS, NTP

**Pattern**: Compose with other firewalls for multi-layer security

---

## Example 3: Production Web Tier Firewall

**Use Case**: Web servers behind a Load Balancer serving HTTPS traffic with restricted SSH access.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-web-firewall
spec:
  name: prod-web-firewall
  tags:
    - prod-web-tier
  inbound_rules:
    # HTTPS from Load Balancer only (not directly from internet)
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-abc123"
    # HTTP for redirect to HTTPS
    - protocol: tcp
      port_range: "80"
      source_load_balancer_uids:
        - "lb-abc123"
    # SSH from office bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # Database access (PostgreSQL)
    - protocol: tcp
      port_range: "5432"
      destination_tags:
        - prod-db-tier
    # External APIs and OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- **Load Balancer UID enforces single entry point** - prevents bypassing LB
- SSH restricted to office bastion (production best practice)
- Outbound DB access via tags (scalable, auto-scales with DB tier)
- Explicit DNS and HTTPS outbound (avoid "allow all" outbound in production)

**Security**: Never allow HTTPS directly from `0.0.0.0/0` in production web tier

---

## Example 4: Production Database Tier Firewall

**Use Case**: PostgreSQL database accessible only by web tier and administrators.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-db-firewall
spec:
  name: prod-db-firewall
  tags:
    - prod-db-tier
  inbound_rules:
    # PostgreSQL from web tier only
    - protocol: tcp
      port_range: "5432"
      source_tags:
        - prod-web-tier
    # SSH from office bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # OS updates (locked to Ubuntu repos)
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "91.189.88.0/21"  # Ubuntu archive CDN
    # DNS (specific resolver)
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "1.1.1.1/32"  # Cloudflare DNS
        - "1.0.0.1/32"  # Cloudflare DNS backup
```

**Notes:**
- **No internet access** - database is completely isolated
- PostgreSQL port (5432) only accessible from web tier via tags
- Outbound locked to specific Ubuntu repos (high security)
- DNS restricted to Cloudflare resolvers (explicit allow, not `0.0.0.0/0`)

**Security**: This is **maximum security** - DB cannot initiate connections to the internet

---

## Example 5: Redis/Memcached Cache Tier Firewall

**Use Case**: In-memory cache accessible by web and API tiers only.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-cache-firewall
spec:
  name: prod-cache-firewall
  tags:
    - prod-cache-tier
  inbound_rules:
    # Redis from web and API tiers
    - protocol: tcp
      port_range: "6379"
      source_tags:
        - prod-web-tier
        - prod-api-tier
    # SSH from bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
    # HTTPS for OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
```

**Notes:**
- Redis port (6379) accessible from multiple tiers via tags
- Tag-based targeting scales automatically with tier growth
- Minimal outbound (DNS + HTTPS for updates)

**Pattern**: Easily extend to support Memcached (port 11211) or other caches

---

## Example 6: Kubernetes Cluster Firewall

**Use Case**: Protect DigitalOcean Kubernetes nodes with firewall rules.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-k8s-firewall
spec:
  name: prod-k8s-firewall
  tags:
    - k8s-prod-cluster
  inbound_rules:
    # Kubernetes API server (from Load Balancer)
    - protocol: tcp
      port_range: "6443"
      source_load_balancer_uids:
        - "lb-k8s-api"
    # NodePort services
    - protocol: tcp
      port_range: "30000-32767"
      source_tags:
        - prod-web-tier
    # SSH from bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # All outbound (for pod egress)
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- Kubernetes API (6443) only accessible via Load Balancer
- NodePort range (30000-32767) for services
- Allow all outbound for pod egress (can be restricted based on workload)

**Alternative**: Use DigitalOcean Kubernetes-native firewalls (automatically applied)

---

## Example 7: Multi-Tier Architecture (3-Tier Web App)

**Use Case**: Complete multi-tier architecture with Load Balancer, web tier, and database tier.

### Load Balancer → Web Tier

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-web-tier
spec:
  name: prod-web-tier
  tags:
    - prod-web-tier
  inbound_rules:
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-prod-web"
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    - protocol: tcp
      port_range: "5432"
      destination_tags:
        - prod-db-tier
    - protocol: tcp
      port_range: "6379"
      destination_tags:
        - prod-cache-tier
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
```

### Web Tier → Cache Tier

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-cache-tier
spec:
  name: prod-cache-tier
  tags:
    - prod-cache-tier
  inbound_rules:
    - protocol: tcp
      port_range: "6379"
      source_tags:
        - prod-web-tier
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
```

### Web Tier → Database Tier

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-db-tier
spec:
  name: prod-db-tier
  tags:
    - prod-db-tier
  inbound_rules:
    - protocol: tcp
      port_range: "5432"
      source_tags:
        - prod-web-tier
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "91.189.88.0/21"
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "1.1.1.1/32"
```

**Notes:**
- Traffic flow: Internet → LB → Web → Cache/DB
- Each tier is isolated via tag-based targeting
- DB tier has most restrictive outbound (locked to specific IPs)
- Compose all three firewalls + management firewall for complete protection

---

## Example 8: Staging Environment Firewall (Testing)

**Use Case**: Firewall for staging environment with limited access from development team.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: staging-firewall
spec:
  name: staging-firewall
  tags:
    - staging
  inbound_rules:
    # HTTPS from Load Balancer
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-staging"
    # SSH from office and VPN
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.0/24"  # Office network
        - "10.8.0.0/24"     # VPN network
  outbound_rules:
    # All outbound (for testing external integrations)
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- More permissive than production (allows office network range)
- SSH from VPN for remote developers
- Unrestricted outbound for testing external APIs
- Still uses Load Balancer UID for HTTPS (production pattern)

---

## Example 9: Bastion Host Firewall

**Use Case**: Hardened bastion host for SSH jump server access to private instances.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: bastion-firewall
spec:
  name: bastion-firewall
  tags:
    - bastion
  inbound_rules:
    # SSH from office and VPN only
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"  # Office static IP
        - "10.8.0.0/24"      # VPN network
  outbound_rules:
    # SSH to all internal instances
    - protocol: tcp
      port_range: "22"
      destination_addresses:
        - "10.0.0.0/8"       # Private network
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
    # HTTPS for OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
```

**Notes:**
- SSH inbound only from trusted sources (office + VPN)
- Outbound SSH to internal network only (prevents lateral movement)
- Minimal outbound (DNS + HTTPS for updates)

**Security**: Bastion should have strong SSH hardening (`PasswordAuthentication no`, fail2ban, etc.)

---

## Example 10: VPN Server Firewall

**Use Case**: WireGuard VPN server for remote team access.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: vpn-server-firewall
spec:
  name: vpn-server-firewall
  tags:
    - vpn-server
  inbound_rules:
    # WireGuard VPN
    - protocol: udp
      port_range: "51820"
      source_addresses:
        - "0.0.0.0/0"
        - "::/0"
    # SSH from office only
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # All outbound (for VPN client egress)
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

**Notes:**
- WireGuard (UDP 51820) open to internet (VPN clients connect from anywhere)
- SSH locked to office (emergency admin access)
- Unrestricted outbound (VPN clients tunnel all traffic)

**Alternative**: OpenVPN (TCP 1194 or UDP 1194)

---

## Common Patterns Summary

| Use Case | Inbound | Outbound | Target Method |
|----------|---------|----------|---------------|
| Dev/Test | SSH, HTTP, HTTPS from `0.0.0.0/0` | All ports | Tags |
| Management | SSH from bastion, Monitoring from tags | DNS, HTTPS, NTP | Tags (shared) |
| Web Tier | HTTPS from LB UID, SSH from bastion | DB via tags, HTTPS | Tags |
| Database Tier | DB port from web tier tags | Specific repos only | Tags |
| Bastion | SSH from office/VPN | SSH to internal only | Static IDs or Tags |
| VPN | VPN port from `0.0.0.0/0` | All (for client egress) | Static IDs or Tags |

---

## Validation Checklist

Before deploying, ensure:

- ✅ `name` is unique and descriptive (1-255 chars)
- ✅ For production, use tags (not static droplet_ids)
- ✅ SSH never allows `0.0.0.0/0` in production
- ✅ Web tier uses Load Balancer UIDs (not direct `0.0.0.0/0` HTTPS)
- ✅ Database tier is isolated (no internet access)
- ✅ Outbound rules are explicit (avoid "allow all" outbound in high-security tiers)
- ✅ ICMP is allowed if you need ping/traceroute for debugging
- ✅ Both Cloud Firewall AND host firewall (`ufw`) are configured consistently

---

## Troubleshooting

### Connection Times Out Despite Firewall Rule

**Cause**: Host firewall (`ufw` or `iptables`) inside Droplet is blocking traffic.

**Solution**: Check host firewall via Web Console:
```bash
sudo ufw status verbose
```

Either allow the port in `ufw` or disable host firewall:
```bash
sudo ufw allow 22/tcp
# OR
sudo ufw disable
```

### Auto-Scaling Creates Unprotected Droplets

**Cause**: Using static `droplet_ids` instead of tags.

**Solution**: Switch to tag-based targeting:
```yaml
tags:
  - web-tier
```

Apply tag to auto-scaling template.

### Firewall Applied But No Effect

**Cause**: Tag mismatch. Firewall has `tags: ["web-tier"]`, Droplets have `web-server`.

**Solution**: Verify tags match exactly. Check Droplet's "Networking" tab in DigitalOcean dashboard.

---

## Next Steps

1. **Deploy**: Use `openmcf pulumi up` or `terraform apply` to create firewalls
2. **Verify**: Check DigitalOcean dashboard → Droplet → Networking tab to see applied firewalls
3. **Test**: Use `telnet` or `nc` to verify connectivity:
   ```bash
   nc -zv <droplet-ip> <port>
   ```
4. **Monitor**: Set up logging for failed connection attempts (host firewall logs)

For more details, see:
- [README.md](./README.md) - Component overview and best practices
- [docs/README.md](./docs/README.md) - Comprehensive production guide
- [iac/pulumi/README.md](./iac/pulumi/README.md) - Pulumi module usage
- [iac/tf/README.md](./iac/tf/README.md) - Terraform module usage

