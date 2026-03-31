# AliCloudCenInstance Pulumi Examples

Apply any of the manifests below with the OpenMCF CLI:

```bash
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

---

## Minimal: Single-Region Multi-VPC

Connects two VPCs in the same region for private inter-VPC communication.

```yaml
apiVersion: alicloud.openmcf.org/v1
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
    - childInstanceId:
        value: vpc-def456
      childInstanceRegionId: cn-hangzhou
```

---

## Cross-Region Backbone with Tags

A global backbone connecting VPCs across China and international regions with
REDUCED protection level for CIDR overlap tolerance.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: global-cen
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  cenInstanceName: global-backbone
  description: Multi-region backbone
  protectionLevel: REDUCED
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

---

## Managed VPC References

Connects VPCs managed as OpenMCF resources using `valueFrom` references that
resolve VPC IDs at deployment time.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudCenInstance
metadata:
  name: managed-cen
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  cenInstanceName: managed-backbone
  description: CEN connecting OpenMCF-managed VPCs
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
