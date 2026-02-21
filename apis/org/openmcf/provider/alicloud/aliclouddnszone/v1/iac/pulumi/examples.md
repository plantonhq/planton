# AlicloudDnsZone Pulumi Examples

Create a YAML manifest using one of the examples below, then deploy with the OpenMCF CLI:

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

## Minimal Domain Registration

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

This adds the domain to the Alidns hosted zone. After deployment, point your domain registrar's NS records to the `dns_servers` output values.

---

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

- Resource group enables access control and cost attribution
- Tags supplement the system tags that OpenMCF adds automatically

---

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

- Domain groups help organize large numbers of domains in the Alidns console
- The `groupId` must reference an existing Alidns domain group
