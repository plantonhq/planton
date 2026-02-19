# AlicloudDnsRecord Pulumi Examples

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

## A Record

Map a subdomain to an IPv4 address.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: web-server
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: www
  type: A
  value: "203.0.113.10"
  ttl: 600
```

---

## CNAME Record

Alias a subdomain to another domain name. Common for CDN integration.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: cdn-alias
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: cdn
  type: CNAME
  value: example.com.cdn-provider.com
```

- The CNAME target should not have a trailing dot

---

## MX Record with Priority

Route email to a mail server. Priority ranges from 1 (highest) to 10 (lowest).

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudDnsRecord
metadata:
  name: mail-primary
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: "@"
  type: MX
  value: mx1.example.com
  priority: 5
  ttl: 3600
```

- `priority` is required for MX records and ignored for all other types
- Use `@` as the `rr` value for apex-level mail routing
