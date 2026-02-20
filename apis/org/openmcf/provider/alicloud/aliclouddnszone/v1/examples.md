# Examples

## Minimal Configuration

Register a domain in Alidns with only the required fields.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsZone
metadata:
  name: my-domain
spec:
  region: cn-hangzhou
  domainName: example.com
```

## Domain with Tags and Resource Group

A domain with organizational tags and resource group placement for access control and cost attribution.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsZone
metadata:
  name: platform-domain
  org: my-org
  env: production
spec:
  region: cn-shanghai
  domainName: platform.example.com
  remark: Primary platform domain for production services
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## Domain with Group Assignment

A domain placed in a specific Alidns domain group for organizational grouping in the console.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsZone
metadata:
  name: grouped-domain
spec:
  region: ap-southeast-1
  domainName: services.example.com
  groupId: group-abc123
  remark: Microservices DNS zone
```
