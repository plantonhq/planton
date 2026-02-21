# Examples

## CLI

All examples use the `openmcf pulumi` CLI. Save the YAML manifest to a file
(e.g., `manifest.yaml`) and run:

```shell
# Preview changes
openmcf pulumi preview \
  --manifest manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Deploy
openmcf pulumi update \
  --manifest manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

---

## Example 1: Minimal Web CDN

Accelerate a website with a single IP-based origin. This is the simplest
configuration — only the required fields are specified.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCdnDomain
metadata:
  name: my-cdn
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: cdn-project
    pulumi.openmcf.org/stack.name: dev.AliCloudCdnDomain.my-cdn
spec:
  region: cn-hangzhou
  domainName: cdn.example.com
  cdnType: web
  sources:
    - type: ipaddr
      content: "203.0.113.10"
```

**Key Points:**

- `cdnType: web` optimizes caching for images, small files, and web pages.
- A single `ipaddr` source uses default port 80, priority 20, and weight 10.
- Geographic scope defaults to `domestic` (mainland China).
- After deployment, create a DNS CNAME record pointing `cdn.example.com` to
  the `cname` stack output.

---

## Example 2: Multi-Origin with Failover and OSS

Serve static assets from an OSS bucket as the primary origin with a domain-based
standby origin for failover. Includes tags and a resource group for
production tracking.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCdnDomain
metadata:
  name: assets-cdn
  org: platform-team
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: platform-team
    pulumi.openmcf.org/project: cdn-project
    pulumi.openmcf.org/stack.name: prod.AliCloudCdnDomain.assets-cdn
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
      content: origin-standby.example.com
      port: 443
      priority: 30
      weight: 10
  checkUrl: http://my-assets.oss-cn-shanghai.aliyuncs.com/health.txt
  resourceGroupId: rg-prod-cdn-456
  tags:
    team: platform
    costCenter: engineering
```

**Key Points:**

- `scope: global` enables edge acceleration worldwide (requires ICP filing).
- The OSS source at priority 20 is the primary origin; the domain source at
  priority 30 is standby.
- `checkUrl` validates OSS bucket reachability during domain creation.
- `resourceGroupId` assigns the domain to a specific resource group for
  access control and billing.
- Tags `team` and `costCenter` are merged with system-generated tags.

---

## Example 3: HTTPS with CAS Certificate

Production CDN domain with HTTPS enabled via a certificate from Alibaba Cloud
Certificate Management Service. Uses multiple weighted origins for load
distribution.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudCdnDomain
metadata:
  name: secure-cdn
  org: my-org
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: cdn-project
    pulumi.openmcf.org/stack.name: prod.AliCloudCdnDomain.secure-cdn
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

**Key Points:**

- Two domain origins share priority 20, so CDN distributes traffic by weight
  (60/40 split).
- `certificateConfig.certType: cas` references a certificate stored in
  Alibaba Cloud Certificate Management Service.
- `certRegion: cn-hangzhou` is used for domestic-scoped domains; use
  `ap-southeast-1` for overseas or global scope.
- `serverCertificateStatus: "on"` enables HTTPS on edge nodes.
- Certificate rotation requires only updating `certId` and redeploying.

---

## Next Steps

- [Pulumi README](./README.md) — CLI reference and debugging guide.
- [Module Overview](./overview.md) — architecture and design decisions.
- [AliCloudCdnDomain Overview](../../../README.md) — full field reference.
