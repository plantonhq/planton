# Examples

## A Record -- Point Subdomain to IP Address

The most common DNS record type. Maps a subdomain to an IPv4 address.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: www-a-record
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: www
  type: A
  value: "203.0.113.10"
  ttl: 600
```

## CNAME Record -- Alias to Another Domain

Points a subdomain to another domain name. Commonly used for CDN integration or service aliasing.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: cdn-cname
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: cdn
  type: CNAME
  value: example.com.cdn-provider.com
```

## MX Record -- Mail Exchange with Priority

Routes email to mail servers. The `priority` field is required for MX records (1 = highest priority, 10 = lowest).

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: mail-mx
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: "@"
  type: MX
  value: mx1.example.com
  priority: 5
  ttl: 3600
```

## TXT Record -- Domain Verification / SPF

TXT records are used for domain ownership verification, SPF, DKIM, and DMARC.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: spf-txt
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: "@"
  type: TXT
  value: "v=spf1 include:example.com ~all"
```

## Wildcard Record

Catches all subdomains that do not have an explicit record.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: wildcard
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: "*"
  type: CNAME
  value: fallback.example.com
```

## Disabled Record

Create a record in DISABLE status. The record exists in Alidns but is not served in DNS responses. Useful for pre-staging records or temporarily taking a record offline.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudDnsRecord
metadata:
  name: staging-record
spec:
  region: cn-hangzhou
  domainName: example.com
  rr: staging-api
  type: A
  value: "10.0.1.100"
  status: DISABLE
  remark: Pre-staged for next release
```
