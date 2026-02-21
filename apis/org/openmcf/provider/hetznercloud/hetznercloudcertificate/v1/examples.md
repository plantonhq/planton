# HetznerCloudCertificate Examples

## Minimal Managed Certificate

The simplest configuration: a managed (Let's Encrypt) certificate for a single domain. Hetzner Cloud obtains and renews the certificate automatically. The domain must resolve to a Hetzner Cloud load balancer with an HTTPS service referencing this certificate.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: api-cert
spec:
  managed:
    domainNames:
      - api.example.com
```

---

## Managed Certificate with Multiple Domains

A SAN (Subject Alternative Name) certificate covering multiple domains. Hetzner Cloud issues a single Let's Encrypt certificate that covers all listed domains. Every domain must have DNS records pointing to a load balancer with an HTTPS service referencing this certificate.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: web-platform-cert
  org: acme-corp
  env: production
spec:
  managed:
    domainNames:
      - example.com
      - www.example.com
      - app.example.com
```

---

## Uploaded Certificate

A user-provided TLS certificate and private key. Use this for wildcard certificates, EV/OV certificates, or certificates from a CA other than Let's Encrypt. Both fields are required and immutable — changing either forces replacement.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: wildcard-cert
  org: acme-corp
  env: production
spec:
  uploaded:
    certificate: |
      -----BEGIN CERTIFICATE-----
      MIIFYDCCBEigAwIBAgISA1...
      -----END CERTIFICATE-----
      -----BEGIN CERTIFICATE-----
      MIIEdTCCA12gAwIBAgIJAN...
      -----END CERTIFICATE-----
    privateKey: |
      -----BEGIN PRIVATE KEY-----
      MIIEvgIBADANBgkqhkiG9w...
      -----END PRIVATE KEY-----
```

The `certificate` field should contain the full chain: server certificate first, followed by intermediate CA certificates in order (root last). The `privateKey` field is treated as sensitive — it is never exposed in plan output or state.

---

## Certificate for Load Balancer HTTPS

The primary use case: a certificate referenced by a `HetznerCloudLoadBalancer` HTTPS service. This example shows both manifests together.

**Step 1** — Create the managed certificate:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudCertificate
metadata:
  name: web-cert
  org: acme-corp
  env: production
spec:
  managed:
    domainNames:
      - example.com
      - www.example.com
```

**Step 2** — Create the load balancer that references the certificate. The load balancer's HTTPS service uses `valueFrom` to reference the certificate's `certificate_id` output:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudLoadBalancer
metadata:
  name: web-lb
  org: acme-corp
  env: production
spec:
  loadBalancerType: lb11
  location: fsn1
  services:
    - protocol: https
      listenPort: 443
      destinationPort: 80
      http:
        certificateIds:
          - valueFrom:
              kind: HetznerCloudCertificate
              name: web-cert
              fieldPath: status.outputs.certificate_id
  targets:
    - type: server
      serverId:
        valueFrom:
          kind: HetznerCloudServer
          name: web-server
          fieldPath: status.outputs.server_id
```

The `valueFrom` reference establishes a dependency edge — the load balancer waits for the certificate to be created before configuring the HTTPS service. The managed certificate, in turn, requires the load balancer's HTTPS service to be reachable for the ACME HTTP-01 challenge. In practice, the certificate enters a pending state until the load balancer is fully configured, then completes automatically.
