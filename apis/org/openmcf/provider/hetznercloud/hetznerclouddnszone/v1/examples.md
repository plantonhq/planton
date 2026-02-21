# HetznerCloudDnsZone Examples

## Minimal Primary Zone

The simplest configuration: a primary DNS zone for a domain with no records. Hetzner Cloud creates the zone and assigns authoritative nameservers. Records can be added later by updating the manifest.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: example-zone
spec:
  domainName: example.com
  mode: primary
```

After deploying, check the `nameservers` output and configure these at your domain registrar to activate the zone.

---

## Primary Zone with Common Records

A production-ready zone with the most common record types: A records for the apex, a CNAME alias for `www`, MX records for email, and TXT records for SPF and DMARC.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: acme-corp-zone
  org: acme-corp
  env: production
spec:
  domainName: acme-corp.com
  mode: primary
  ttl: 3600
  recordSets:
    - name: "@"
      type: A
      ttl: 300
      records:
        - value: "93.184.216.34"
          comment: "web server 1"
        - value: "93.184.216.35"
          comment: "web server 2"
    - name: "@"
      type: AAAA
      records:
        - value: "2606:2800:220:1:248:1893:25c8:1946"
    - name: www
      type: CNAME
      records:
        - value: "acme-corp.com."
    - name: "@"
      type: MX
      records:
        - value: "10 mail.acme-corp.com."
        - value: "20 backup-mail.acme-corp.com."
    - name: "@"
      type: TXT
      records:
        - value: "\"v=spf1 include:_spf.google.com ~all\""
    - name: _dmarc
      type: TXT
      records:
        - value: "\"v=DMARC1; p=reject; rua=mailto:dmarc@acme-corp.com\""
```

Note the formatting conventions:
- MX values include the priority as part of the string: `"10 mail.acme-corp.com."`
- CNAME and MX targets end with a trailing dot (FQDN)
- TXT values are enclosed in escaped double quotes within the YAML string
- The apex A records use a low TTL (300s) for faster failover, while other records inherit the zone default (3600s)

---

## Primary Zone with Cross-Component References

A DNS zone where A records reference the IP addresses of other Hetzner Cloud resources using `valueFrom`. When the server or load balancer IP changes, the DNS record updates automatically on the next deployment.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: webapp-zone
  org: acme-corp
  env: production
spec:
  domainName: webapp.example.com
  mode: primary
  ttl: 3600
  recordSets:
    - name: "@"
      type: A
      ttl: 60
      records:
        - value:
            valueFrom:
              kind: HetznerCloudLoadBalancer
              name: web-lb
              fieldPath: status.outputs.ipv4_address
    - name: api
      type: A
      ttl: 60
      records:
        - value:
            valueFrom:
              kind: HetznerCloudServer
              name: api-server
              fieldPath: status.outputs.ipv4_address
    - name: failover
      type: A
      ttl: 60
      records:
        - value:
            valueFrom:
              kind: HetznerCloudFloatingIp
              name: failover-ip
              fieldPath: status.outputs.ip_address
    - name: www
      type: CNAME
      records:
        - value: "webapp.example.com."
    - name: "@"
      type: TXT
      records:
        - value: "\"v=spf1 -all\""
    - name: "@"
      type: CAA
      records:
        - value: "0 issue \"letsencrypt.org\""
```

The `valueFrom` references establish dependency edges — the DNS zone waits for the referenced resources to be provisioned before creating the records with their resolved IP addresses. Literal values (CNAME, TXT, CAA) coexist naturally with referenced values in the same manifest.

---

## Primary Zone with Advanced Records

A zone demonstrating less common record types, wildcard records, per-record-set TTL overrides, record comments, and delete protection.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: advanced-zone
  org: acme-corp
  env: production
spec:
  domainName: advanced.example.com
  mode: primary
  ttl: 3600
  deleteProtection: true
  recordSets:
    - name: "@"
      type: A
      ttl: 300
      records:
        - value: "93.184.216.34"
          comment: "primary datacenter"
        - value: "93.184.216.35"
          comment: "secondary datacenter"
    - name: "*"
      type: A
      ttl: 300
      records:
        - value: "93.184.216.34"
          comment: "wildcard catch-all"
    - name: _sip._tcp
      type: SRV
      records:
        - value: "10 60 5060 sip.advanced.example.com."
          comment: "SIP over TCP"
    - name: "@"
      type: CAA
      records:
        - value: "0 issue \"letsencrypt.org\""
        - value: "0 iodef \"mailto:security@advanced.example.com\""
    - name: "@"
      type: NS
      ttl: 86400
      records:
        - value: "ns1.advanced.example.com."
        - value: "ns2.advanced.example.com."
    - name: mail._domainkey
      type: TXT
      records:
        - value: "\"v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQ...\""
          comment: "DKIM signing key"
    - name: "@"
      type: MX
      records:
        - value: "10 mail.advanced.example.com."
        - value: "20 backup.advanced.example.com."
```

Key details:
- `deleteProtection: true` prevents accidental deletion of this production zone
- The wildcard `*` A record catches all subdomains not explicitly defined
- SRV records use the `_service._protocol` naming convention
- CAA records specify both the allowed issuer and an incident reporting address (`iodef`)
- The DKIM TXT record sits under `mail._domainkey` — the standard DKIM selector path

---

## Secondary Zone with TSIG Authentication

A secondary zone that synchronizes records from an external primary nameserver. TSIG authentication secures the zone transfer. No `recordSets` are specified — records come from the primary.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: secondary-zone
  org: acme-corp
  env: production
spec:
  domainName: internal.example.com
  mode: secondary
  primaryNameservers:
    - address: "10.0.0.1"
      port: 53
      tsigAlgorithm: hmac-sha256
      tsigKey: "dGhpcyBpcyBhIHNlY3JldCBrZXk="
    - address: "10.0.0.2"
      port: 53
      tsigAlgorithm: hmac-sha256
      tsigKey: "dGhpcyBpcyBhIHNlY3JldCBrZXk="
```

After deploying, Hetzner Cloud initiates zone transfers from the primary nameservers. Records appear once the first transfer completes. The `tsigKey` is a base64-encoded shared secret that must match the key configured on the primary nameservers.

Multiple primary nameservers provide redundancy: if the first is unreachable, Hetzner Cloud attempts the second. Both use the same TSIG credentials in this example, but each can have independent credentials if needed.

For a secondary zone without TSIG (IP-based ACL only), omit the `tsigAlgorithm` and `tsigKey` fields:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudDnsZone
metadata:
  name: secondary-no-tsig
spec:
  domainName: partner.example.com
  mode: secondary
  primaryNameservers:
    - address: "203.0.113.53"
```
