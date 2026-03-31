# Examples

## Minimal Single-Zone VSwitch

Creates a VSwitch in one availability zone with a /24 CIDR block. Suitable for development or small workloads.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: dev-vswitch
spec:
  region: cn-hangzhou
  vpcId: vpc-abc123def456
  zoneId: cn-hangzhou-a
  cidrBlock: "192.168.0.0/24"
  vswitchName: dev-vswitch
```

## Production VSwitch with Tags

A production VSwitch using a larger CIDR block for Kubernetes node pools. Tags are applied for cost tracking and organizational grouping.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: prod-app-vswitch
  org: my-org
  env: production
spec:
  region: cn-shanghai
  vpcId: vpc-prod-001
  zoneId: cn-shanghai-b
  cidrBlock: "10.1.0.0/20"
  vswitchName: prod-app-tier-b
  description: Application tier VSwitch in zone B for Kubernetes workers
  tags:
    team: platform
    costCenter: engineering
    tier: application
```

## VSwitch with Cross-Resource Reference

Uses a `valueFrom` reference to resolve the VPC ID from an existing AliCloudVpc component, establishing a declarative dependency.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: db-vswitch
  env: staging
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  zoneId: cn-hangzhou-c
  cidrBlock: "10.2.0.0/24"
  vswitchName: staging-db-vswitch
  description: Database tier VSwitch for RDS and Redis instances
```

## IPv6-Enabled VSwitch

A dual-stack VSwitch with IPv6 support. The parent VPC must also have IPv6 enabled for this to take effect.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudVswitch
metadata:
  name: ipv6-vswitch
spec:
  region: us-west-1
  vpcId: vpc-ipv6-enabled
  zoneId: us-west-1a
  cidrBlock: "172.16.0.0/24"
  vswitchName: ipv6-app-vswitch
  description: Dual-stack VSwitch for IPv6 workloads
  enableIpv6: true
  ipv6CidrBlockMask: 42
  tags:
    networkType: dual-stack
```
