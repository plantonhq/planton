# AliCloud Enterprise Network Foundation

Multi-AZ VPC networking foundation with NAT gateway, security groups, and optional VPN/CEN hybrid connectivity for Alibaba Cloud.

## Resources

| Resource | Kind | Purpose |
|----------|------|---------|
| VPC | `AliCloudVpc` | Isolated virtual network |
| VSwitch (×2) | `AliCloudVswitch` | Subnets across two availability zones |
| Security Group | `AliCloudSecurityGroup` | Default firewall with HTTP/HTTPS ingress and all egress |
| EIP | `AliCloudEipAddress` | Static public IP for NAT Gateway |
| NAT Gateway | `AliCloudNatGateway` | Outbound internet access via SNAT for both VSwitches |
| VPN Gateway | `AliCloudVpnGateway` | Site-to-site VPN connectivity (conditional) |
| CEN Instance | `AliCloudCenInstance` | Multi-VPC peering via Cloud Enterprise Network (conditional) |

## Dependency Graph

```
                     ┌─────────┐
                     │   VPC   │
                     └────┬────┘
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
   ┌──────────┐    ┌──────────┐    ┌──────────┐
   │ VSwitch  │    │ VSwitch  │    │ Security │
   │   AZ-1   │    │   AZ-2   │    │  Group   │
   └────┬─────┘    └────┬─────┘    └──────────┘
        │               │
   ┌────┴───┐           │         ┌──────────┐
   │  EIP   │           │         │   VPN    │ (optional)
   └────┬───┘           │         └──────────┘
        │               │
   ┌────┴───────────────┴──┐      ┌──────────┐
   │     NAT Gateway       │      │   CEN    │ (optional)
   │  (SNAT for both AZs)  │      └──────────┘
   └───────────────────────┘
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Alibaba Cloud region | `cn-hangzhou` |
| `availability_zone_1` | First AZ | `cn-hangzhou-h` |
| `availability_zone_2` | Second AZ | `cn-hangzhou-i` |
| `vpc_cidr` | VPC CIDR block | `10.0.0.0/16` |
| `vswitch_cidr_1` | VSwitch 1 CIDR | `10.0.0.0/20` |
| `vswitch_cidr_2` | VSwitch 2 CIDR | `10.0.16.0/20` |
| `allowHttpIngress` | Add HTTP/HTTPS ingress rules | `true` |
| `eip_bandwidth` | NAT EIP bandwidth (Mbps) | `10` |
| `vpnEnabled` | Create VPN Gateway | `false` |
| `vpn_bandwidth` | VPN bandwidth (Mbps) | `10` |
| `cenEnabled` | Create CEN instance | `false` |

## Deployment Time

| Phase | Duration |
|-------|----------|
| VPC | ~15 seconds |
| VSwitches + SG + EIP | ~30 seconds |
| NAT Gateway | ~2 minutes |
| VPN Gateway (if enabled) | ~5 minutes |
| CEN (if enabled) | ~1 minute |
| **Total (core)** | **~3 minutes** |
| **Total (with VPN + CEN)** | **~8 minutes** |
