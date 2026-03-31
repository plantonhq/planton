# Examples

## Minimal Configuration

Create a private zone attached to a single VPC with no records. Records can be added later or managed separately.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: internal-zone
spec:
  region: cn-hangzhou
  zoneName: internal.example.com
  vpcAttachments:
    - vpcId: vpc-abc123def
```

## Service Discovery Zone with Records

A private zone for internal service discovery with A records pointing to private IP addresses. This is the common pattern for microservices that need to find each other by hostname within a VPC.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: svc-discovery
  org: my-org
  env: production
spec:
  region: cn-shanghai
  zoneName: svc.internal
  remark: Internal service discovery for production microservices
  vpcAttachments:
    - vpcId: vpc-app-prod
  records:
    - rr: api
      type: A
      value: "10.0.1.50"
      ttl: 120
    - rr: cache
      type: A
      value: "10.0.2.30"
    - rr: db-master
      type: A
      value: "10.0.3.100"
      remark: Primary database endpoint
    - rr: db-replica
      type: A
      value: "10.0.3.101"
      remark: Read replica endpoint
  tags:
    team: platform
    service: dns
```

## Multi-VPC Database Zone

A private zone shared across multiple VPCs (including cross-region) for database endpoint discovery. This pattern is used when databases in one VPC need to be reachable from application VPCs in different regions.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: db-zone
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  zoneName: db.corp
  remark: Database endpoints shared across all application VPCs
  resourceGroupId: rg-prod-123
  vpcAttachments:
    - vpcId: vpc-app-hangzhou
    - vpcId: vpc-app-shanghai
      regionId: cn-shanghai
    - vpcId: vpc-mgmt
  records:
    - rr: mysql-primary
      type: A
      value: "10.0.10.100"
      ttl: 60
    - rr: mysql-replica
      type: A
      value: "10.0.10.101"
    - rr: redis
      type: A
      value: "10.0.11.50"
    - rr: mongo
      type: CNAME
      value: dds-abc123.mongodb.rds.aliyuncs.com
      remark: MongoDB managed instance internal endpoint
  tags:
    team: dba
    costCenter: infrastructure
```
