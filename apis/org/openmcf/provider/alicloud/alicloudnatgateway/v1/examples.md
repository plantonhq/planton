# AlicloudNatGateway Examples

## Minimal: Single VSwitch SNAT

The simplest NAT Gateway configuration: one EIP, one SNAT entry for a single VSwitch.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: dev-nat
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  vswitchId:
    value: vsw-nat-zone-a
  natGatewayName: dev-nat
  eipId:
    value: eip-abc123
  snatEntries:
    - sourceVswitchId:
        value: vsw-app-zone-a
      snatEntryName: app-internet
```

## Production: Multi-AZ with Deletion Protection

NAT Gateway serving multiple VSwitches across availability zones with deletion protection enabled.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: prod-nat
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  vswitchId:
    valueFrom:
      name: nat-vswitch-a
  natGatewayName: prod-nat-gateway
  description: Production NAT Gateway for all workload VSwitches
  deletionProtection: true
  tags:
    team: platform
    cost-center: shared-infra
  eipId:
    valueFrom:
      name: prod-nat-eip
  snatEntries:
    - sourceVswitchId:
        valueFrom:
          name: app-vswitch-a
      snatEntryName: app-zone-a
    - sourceVswitchId:
        valueFrom:
          name: app-vswitch-b
      snatEntryName: app-zone-b
    - sourceVswitchId:
        valueFrom:
          name: db-vswitch-a
      snatEntryName: db-zone-a
```

## CIDR-based SNAT

Use source CIDR instead of VSwitch ID for fine-grained control over which IP ranges get NATed.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: cidr-nat
spec:
  region: us-west-1
  vpcId:
    value: vpc-usw1
  vswitchId:
    value: vsw-nat-usw1a
  natGatewayName: cidr-nat
  eipId:
    value: eip-usw1
  snatEntries:
    - sourceCidr: "10.0.1.0/24"
      snatEntryName: app-subnet
    - sourceCidr: "10.0.2.0/24"
      snatEntryName: worker-subnet
```
