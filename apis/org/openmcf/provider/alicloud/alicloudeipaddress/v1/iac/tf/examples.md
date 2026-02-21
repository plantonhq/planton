# Examples

## Minimal EIP

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: my-eip
spec:
  region: cn-hangzhou
```

## EIP for NAT Gateway

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: nat-eip
spec:
  region: cn-shanghai
  addressName: prod-nat-eip
  description: EIP for production NAT gateway
  bandwidth: 10
  tags:
    purpose: nat
```

## High-Bandwidth Production EIP

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudEipAddress
metadata:
  name: prod-lb-eip
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  addressName: prod-alb-eip
  bandwidth: 100
  internetChargeType: PayByBandwidth
  isp: BGP_PRO
  resourceGroupId: rg-prod-123
  tags:
    team: platform
```
