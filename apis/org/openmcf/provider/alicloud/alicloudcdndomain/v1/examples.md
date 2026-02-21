# Examples

## Minimal Web CDN

Accelerate a website with a single IP origin.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCdnDomain
metadata:
  name: my-cdn
spec:
  region: cn-hangzhou
  domainName: cdn.example.com
  cdnType: web
  sources:
    - type: ipaddr
      content: "203.0.113.10"
```

## Multiple Origins with Failover

A CDN domain with a primary and standby origin. If the primary origin is unreachable, CDN falls back to the standby.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCdnDomain
metadata:
  name: ha-cdn
  org: my-org
  env: production
spec:
  region: cn-shanghai
  domainName: static.example.com
  cdnType: download
  scope: global
  sources:
    - type: domain
      content: origin-primary.example.com
      port: 443
      priority: 20
      weight: 10
    - type: domain
      content: origin-standby.example.com
      port: 443
      priority: 30
      weight: 10
  resourceGroupId: rg-prod-456
  tags:
    team: platform
    costCenter: engineering
```

## CDN with HTTPS (CAS Certificate)

Enable HTTPS using a certificate from Alibaba Cloud Certificate Management Service.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCdnDomain
metadata:
  name: secure-cdn
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  domainName: secure.example.com
  cdnType: web
  scope: domestic
  sources:
    - type: domain
      content: origin.example.com
      port: 443
  certificateConfig:
    certType: cas
    certId: cas-cn-abc123
    certRegion: cn-hangzhou
    serverCertificateStatus: "on"
```

## OSS Bucket as Origin

Serve static assets from an OSS bucket via CDN.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCdnDomain
metadata:
  name: assets-cdn
spec:
  region: cn-hangzhou
  domainName: assets.example.com
  cdnType: web
  sources:
    - type: oss
      content: my-static-assets.oss-cn-hangzhou.aliyuncs.com
  checkUrl: http://my-static-assets.oss-cn-hangzhou.aliyuncs.com/health.txt
```
