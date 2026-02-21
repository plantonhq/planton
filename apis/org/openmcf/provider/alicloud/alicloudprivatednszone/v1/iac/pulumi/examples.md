# Pulumi Examples

## Minimal Zone

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: minimal-zone
spec:
  region: cn-hangzhou
  zoneName: internal.example.com
  vpcAttachments:
    - vpcId: vpc-abc123
```

## Zone with Service Discovery Records

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: svc-zone
  org: my-org
  env: production
spec:
  region: cn-shanghai
  zoneName: svc.internal
  vpcAttachments:
    - vpcId: vpc-prod-app
  records:
    - rr: api
      type: A
      value: "10.0.1.50"
    - rr: db
      type: A
      value: "10.0.2.100"
  tags:
    team: platform
```

## Multi-VPC Cross-Region Zone

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudPrivateDnsZone
metadata:
  name: shared-zone
spec:
  region: cn-hangzhou
  zoneName: shared.corp
  vpcAttachments:
    - vpcId: vpc-hz
    - vpcId: vpc-sh
      regionId: cn-shanghai
```
