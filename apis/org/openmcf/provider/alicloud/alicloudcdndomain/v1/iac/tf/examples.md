# Examples

## CLI

All examples use the `openmcf tofu` CLI. Save the YAML manifest to a file
(e.g., `manifest.yaml`) and run:

```shell
# Plan changes
openmcf tofu plan \
  --manifest manifest.yaml \
  --module-dir .

# Apply
openmcf tofu apply \
  --manifest manifest.yaml \
  --module-dir .
```

---

## Example 1: Minimal Web CDN

Accelerate a website with a single IP-based origin. Only required fields are
specified.

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

- `cdnType: web` optimizes caching for images, small files, and web pages.
- A single `ipaddr` source uses default port 80, priority 20, and weight 10.
- Geographic scope defaults to `domestic` (mainland China).
- After deployment, create a DNS CNAME record pointing `cdn.example.com` to
  the `cname` output.

---

## Example 2: OSS Origin with Global Scope

Serve static assets from an OSS bucket with a standby domain origin for
failover. Includes resource group and tags.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCdnDomain
metadata:
  name: assets-cdn
  org: platform-team
  env: production
spec:
  region: cn-shanghai
  domainName: assets.example.com
  cdnType: web
  scope: global
  sources:
    - type: oss
      content: my-assets.oss-cn-shanghai.aliyuncs.com
      priority: 20
    - type: domain
      content: origin-backup.example.com
      port: 443
      priority: 30
      weight: 10
  checkUrl: http://my-assets.oss-cn-shanghai.aliyuncs.com/health.txt
  resourceGroupId: rg-prod-456
  tags:
    team: platform
    costCenter: engineering
```

- `scope: global` enables worldwide edge acceleration (requires ICP filing).
- The OSS source at priority 20 is primary; the domain source at priority 30
  is standby.
- `checkUrl` verifies OSS bucket reachability during domain creation.
- `resourceGroupId` scopes the domain for access control and billing.

---

## Example 3: HTTPS with CAS Certificate and Weighted Origins

Production configuration with HTTPS enabled via Alibaba Cloud Certificate
Management Service and weighted load distribution across two origins.

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
      content: origin-a.example.com
      port: 443
      priority: 20
      weight: 60
    - type: domain
      content: origin-b.example.com
      port: 443
      priority: 20
      weight: 40
  certificateConfig:
    certType: cas
    certId: cas-cn-abc123
    certRegion: cn-hangzhou
    serverCertificateStatus: "on"
  resourceGroupId: rg-prod-789
  tags:
    team: security
    compliance: pci
```

- Two origins at priority 20 receive traffic weighted 60/40.
- `certType: cas` references a certificate in Certificate Management Service.
- `certRegion: cn-hangzhou` is for domestic scope; use `ap-southeast-1` for
  overseas or global.
- Certificate rotation requires only updating `certId` and reapplying.

---

## After Deploying

Verify the CDN domain status and CNAME assignment using the Alibaba Cloud CLI:

```shell
# Check domain status and CNAME
aliyun cdn DescribeCdnDomainDetail --DomainName cdn.example.com

# Verify DNS CNAME record
dig cdn.example.com CNAME

# Test origin connectivity through CDN
curl -I https://cdn.example.com/index.html

# Verify HTTPS certificate details
openssl s_client -connect cdn.example.com:443 -servername cdn.example.com </dev/null 2>/dev/null \
  | openssl x509 -noout -subject -dates
```
