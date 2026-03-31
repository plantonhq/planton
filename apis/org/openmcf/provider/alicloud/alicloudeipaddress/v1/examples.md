# Examples

## Minimal Configuration

Allocates an EIP with defaults: 5 Mbps, PayByTraffic, BGP. Suitable for development or testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: my-eip
spec:
  region: cn-hangzhou
```

## Named EIP for NAT Gateway

A named EIP intended for association with a NAT gateway, using the default 5 Mbps bandwidth.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: nat-eip
spec:
  region: cn-shanghai
  addressName: prod-nat-eip
  description: EIP for production NAT gateway outbound access
  tags:
    purpose: nat
    team: platform
```

## High-Bandwidth Production EIP

A production EIP with 100 Mbps guaranteed bandwidth, billed per bandwidth allocation. Uses BGP_PRO for optimized routing in mainland China.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: prod-lb-eip
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  addressName: prod-alb-eip
  description: High-bandwidth EIP for production ALB
  bandwidth: 100
  internetChargeType: PayByBandwidth
  isp: BGP_PRO
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```
