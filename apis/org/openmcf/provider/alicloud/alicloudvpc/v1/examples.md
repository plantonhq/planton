# Examples

## Minimal Configuration

Creates a VPC with the smallest standard private CIDR range, suitable for development or testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: my-vpc
spec:
  region: cn-hangzhou
  vpcName: dev-vpc
  cidrBlock: "192.168.0.0/16"
```

## Production VPC with Tags

A production VPC using a large CIDR block to accommodate many VSwitches across multiple availability zones. Tags are applied for organizational tracking.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: prod-vpc
  org: my-org
  env: production
spec:
  region: cn-shanghai
  vpcName: prod-platform-vpc
  cidrBlock: "10.0.0.0/8"
  description: Production VPC for the platform workloads
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## IPv6-Enabled VPC

A VPC with dual-stack networking. Alibaba Cloud allocates a /56 IPv6 CIDR block when IPv6 is enabled. VSwitches within this VPC can then be assigned IPv6 subnets.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudVpc
metadata:
  name: ipv6-vpc
  env: staging
spec:
  region: us-west-1
  vpcName: ipv6-enabled-vpc
  cidrBlock: "172.16.0.0/12"
  description: Dual-stack VPC with IPv6 support
  enableIpv6: true
  tags:
    networkType: dual-stack
```
