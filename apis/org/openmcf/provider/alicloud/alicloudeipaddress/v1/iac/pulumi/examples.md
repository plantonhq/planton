# Examples

Create a YAML file using one of the examples below. Deploy with the OpenMCF Pulumi CLI:

```shell
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Minimal EIP

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEipAddress
metadata:
  name: my-eip
spec:
  region: cn-hangzhou
```

## EIP for NAT Gateway

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEipAddress
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
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEipAddress
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
