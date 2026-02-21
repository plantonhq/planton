# AliCloudCenInstance Terraform Examples

Apply any of the manifests below with the OpenMCF CLI:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

---

## Minimal: Single-Region Multi-VPC

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
    - childInstanceId:
        value: vpc-def456
      childInstanceRegionId: cn-hangzhou
```

```shell
openmcf tofu apply --manifest basic-cen.yaml --auto-approve
```

---

## Cross-Region Backbone

```yaml
apiVersion: ali-cloud.openmcf.org/v1
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

```shell
openmcf tofu apply --manifest global-cen.yaml --auto-approve
```

---

## Managed VPC References

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

```shell
openmcf tofu apply --manifest managed-cen.yaml --auto-approve
```

---

## After Deploying

Verify the CEN instance:

```shell
openmcf tofu output cen_id
openmcf tofu output cen_instance_name
```

To tear down:

```shell
openmcf tofu destroy --manifest <manifest>.yaml --auto-approve
```
