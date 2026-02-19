# AlicloudEcsInstance Examples

## Minimal Development Instance

A basic ECS instance with SSH key authentication, no public IP, and default system disk.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEcsInstance
metadata:
  name: dev-server
spec:
  region: cn-hangzhou
  instanceType: ecs.g7.large
  imageId: ubuntu_22_04_x64_20G_alibase_20230515.vhd
  vswitchId:
    value: vsw-abc123
  securityGroupIds:
    - value: sg-abc123
  keyName: my-dev-keypair
```

## Production Web Server

A larger instance with ESSD PL1 system disk, encrypted data disk for application data, public IP for serving traffic, deletion protection, and a RAM role for accessing other Alibaba Cloud services.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEcsInstance
metadata:
  name: prod-web-01
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  instanceType: ecs.g7.2xlarge
  imageId: aliyun_3_x64_20G_alibase_20230727.vhd
  vswitchId:
    ref:
      kind: AlicloudVswitch
      name: prod-vswitch-a
      fieldPath: status.outputs.vswitch_id
  securityGroupIds:
    - ref:
        kind: AlicloudSecurityGroup
        name: prod-web-sg
        fieldPath: status.outputs.security_group_id
    - ref:
        kind: AlicloudSecurityGroup
        name: prod-mgmt-sg
        fieldPath: status.outputs.security_group_id
  keyName: prod-keypair
  instanceName: prod-web-01
  hostName: web-01
  description: Production web server
  systemDisk:
    category: cloud_essd
    size: 100
    performanceLevel: PL1
    encrypted: true
    kmsKeyId: kms-abc123
  dataDisks:
    - size: 200
      category: cloud_essd
      name: app-data
      performanceLevel: PL1
      encrypted: true
      kmsKeyId: kms-abc123
  internetMaxBandwidthOut: 20
  internetChargeType: PayByTraffic
  roleName: WebServerRole
  deletionProtection: true
  tags:
    team: platform
    service: web-frontend
```

## Spot Batch Worker

A cost-efficient spot instance for batch processing workloads that can tolerate interruption.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEcsInstance
metadata:
  name: batch-worker-01
spec:
  region: cn-hangzhou
  instanceType: ecs.c7.2xlarge
  imageId: ubuntu_22_04_x64_20G_alibase_20230515.vhd
  vswitchId:
    value: vsw-abc123
  securityGroupIds:
    - value: sg-worker123
  keyName: batch-keypair
  spotStrategy: SpotAsPriceGo
  systemDisk:
    category: cloud_essd
    size: 40
  dataDisks:
    - size: 500
      category: cloud_efficiency
      name: scratch-space
      deleteWithInstance: true
  userData: IyEvYmluL2Jhc2gKZWNobyAiU3RhcnRpbmcgYmF0Y2ggd29ya2VyLi4uIg==
  tags:
    workload: batch
```

## PrePaid Instance with RAM Role

A subscription (PrePaid) instance for predictable long-term workloads with lower per-hour costs.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEcsInstance
metadata:
  name: app-server
  org: acme-corp
  env: staging
spec:
  region: cn-beijing
  instanceType: ecs.g7.xlarge
  imageId: aliyun_3_x64_20G_alibase_20230727.vhd
  vswitchId:
    value: vsw-staging123
  securityGroupIds:
    - value: sg-staging123
  keyName: staging-keypair
  instanceChargeType: PrePaid
  period: 12
  periodUnit: Month
  systemDisk:
    category: cloud_essd
    size: 80
    performanceLevel: PL1
  roleName: AppServerRole
  deletionProtection: true
  securityEnhancementStrategy: Active
  resourceGroupId: rg-staging
```
