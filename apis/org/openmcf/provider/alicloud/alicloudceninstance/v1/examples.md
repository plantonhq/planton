# AliCloudCenInstance Examples

## Minimal: Single VPC Attachment

The simplest CEN setup: one instance with a single VPC attached. Useful as a placeholder before adding more networks.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: basic-cen
spec:
  region: cn-hangzhou
  cenInstanceName: basic-cen
  attachments:
    - childInstanceId:
        value: vpc-abc123
      childInstanceRegionId: cn-hangzhou
```

## Multi-VPC in Same Region

Connect production and shared-services VPCs in the same region for private communication.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: intra-region-cen
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  cenInstanceName: intra-region-backbone
  description: Connects production and shared-services VPCs in cn-shanghai
  tags:
    team: network
    costCenter: shared-infra
  attachments:
    - childInstanceId:
        value: vpc-prod-shanghai
      childInstanceRegionId: cn-shanghai
    - childInstanceId:
        value: vpc-shared-shanghai
      childInstanceRegionId: cn-shanghai
```

## Cross-Region Connectivity

Connect VPCs across multiple Alibaba Cloud regions for a global backbone. CIDR overlap protection is relaxed with `protectionLevel: REDUCED` to allow overlapping address spaces managed by route maps.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: global-backbone
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  cenInstanceName: global-backbone
  description: Multi-region backbone connecting China and Southeast Asia
  protectionLevel: REDUCED
  resourceGroupId: rg-network-team
  tags:
    team: platform
    purpose: global-connectivity
  attachments:
    - childInstanceId:
        value: vpc-hangzhou
      childInstanceRegionId: cn-hangzhou
    - childInstanceId:
        value: vpc-shanghai
      childInstanceRegionId: cn-shanghai
    - childInstanceId:
        value: vpc-singapore
      childInstanceRegionId: ap-southeast-1
```

## Cross-Reference with valueFrom

CEN instance referencing VPCs managed by other OpenMCF resources using `valueFrom`.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: managed-cen
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  cenInstanceName: managed-backbone
  description: References VPCs from other OpenMCF components
  attachments:
    - childInstanceId:
        valueFrom:
          name: prod-vpc-hangzhou
      childInstanceRegionId: cn-hangzhou
    - childInstanceId:
        valueFrom:
          name: prod-vpc-shanghai
      childInstanceRegionId: cn-shanghai
```
