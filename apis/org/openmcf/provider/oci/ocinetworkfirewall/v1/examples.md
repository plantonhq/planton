# OciNetworkFirewall Examples

## Basic Allow/Deny Firewall

A firewall allowing internal RFC 1918 traffic and dropping everything else:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: basic-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciNetworkFirewall.basic-fw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  policy:
    addressLists:
      - name: "internal"
        type: ip
        addresses:
          - "10.0.0.0/8"
          - "172.16.0.0/12"
          - "192.168.0.0/16"
    securityRules:
      - name: "allow-internal"
        action: allow
        condition:
          sourceAddresses:
            - "internal"
      - name: "deny-all"
        action: drop
        condition: {}
```

## Web Application Firewall with Services and URLs

A firewall allowing HTTPS traffic to web servers and blocking malicious URLs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: web-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkFirewall.web-fw
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: firewall-subnet
      fieldPath: status.outputs.subnetId
  policy:
    addressLists:
      - name: "web-servers"
        type: ip
        addresses:
          - "10.0.1.0/24"
      - name: "blocked-domains"
        type: fqdn
        addresses:
          - "malware.example.com"
          - "phishing.example.com"
    services:
      - name: "https"
        type: tcp_service
        portRanges:
          - minimumPort: 443
            maximumPort: 443
      - name: "http"
        type: tcp_service
        portRanges:
          - minimumPort: 80
            maximumPort: 80
    serviceLists:
      - name: "web-traffic"
        services:
          - "http"
          - "https"
    urlLists:
      - name: "blocked-urls"
        urls:
          - pattern: "*.malware.example.com"
          - pattern: "phishing.example.com/*"
    securityRules:
      - name: "block-malicious-urls"
        action: drop
        condition:
          urls:
            - "blocked-urls"
      - name: "allow-web-to-servers"
        action: allow
        condition:
          destinationAddresses:
            - "web-servers"
          services:
            - "web-traffic"
      - name: "deny-all"
        action: drop
        condition: {}
```

## IDS/IPS Inspection Firewall

A firewall with intrusion prevention on inbound traffic and private NAT:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: ids-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkFirewall.ids-fw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  natConfiguration:
    mustEnablePrivateNat: true
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: fw-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  policy:
    addressLists:
      - name: "external"
        type: ip
        addresses:
          - "0.0.0.0/0"
      - name: "internal"
        type: ip
        addresses:
          - "10.0.0.0/8"
    services:
      - name: "all-tcp"
        type: tcp_service
        portRanges:
          - minimumPort: 1
            maximumPort: 65535
    securityRules:
      - name: "inspect-inbound"
        action: inspect
        inspection: intrusion_prevention
        condition:
          sourceAddresses:
            - "external"
          destinationAddresses:
            - "internal"
          services:
            - "all-tcp"
      - name: "allow-internal"
        action: allow
        condition:
          sourceAddresses:
            - "internal"
      - name: "deny-all"
        action: drop
        condition: {}
```

## Microsegmentation Firewall

A firewall enforcing strict tier-to-tier communication (web -> app -> database):

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: microseg-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkFirewall.microseg-fw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  policy:
    addressLists:
      - name: "web-tier"
        type: ip
        addresses:
          - "10.0.1.0/24"
      - name: "app-tier"
        type: ip
        addresses:
          - "10.0.2.0/24"
      - name: "db-tier"
        type: ip
        addresses:
          - "10.0.3.0/24"
    services:
      - name: "app-port"
        type: tcp_service
        portRanges:
          - minimumPort: 8080
            maximumPort: 8080
      - name: "db-port"
        type: tcp_service
        portRanges:
          - minimumPort: 1521
            maximumPort: 1521
    securityRules:
      - name: "web-to-app"
        action: allow
        condition:
          sourceAddresses:
            - "web-tier"
          destinationAddresses:
            - "app-tier"
          services:
            - "app-port"
      - name: "app-to-db"
        action: allow
        condition:
          sourceAddresses:
            - "app-tier"
          destinationAddresses:
            - "db-tier"
          services:
            - "db-port"
      - name: "deny-all"
        action: drop
        condition: {}
```

## Common Operations

### Add a new address list

Add a new entry to `policy.addressLists` and reference it in security rules. Re-apply.

### Add a security rule

Append a new rule to `policy.securityRules`. Rules are evaluated top-to-bottom, so position matters — insert the new rule at the appropriate position for the desired priority.

### Update addresses in a list

Modify the `addresses` array in an existing address list and re-apply. Address values are updatable without recreation; the list name is ForceNew.

### Route traffic through the firewall

After deployment, use the `ipv4Address` stack output to create route table entries in your VCN subnets that direct traffic through the firewall for inspection.

## Best Practices

1. **Always end with a deny-all rule** — ensures no traffic bypasses the firewall rules. Place it last in the `securityRules` list.
2. **Order rules from most specific to least specific** — rules are evaluated top-to-bottom; the first match wins.
3. **Use service lists for reuse** — group related services (e.g., `http` + `https` into `web-traffic`) to keep rules readable.
4. **Use descriptive names** — address list, service, and rule names appear in OCI Console logs and audit trails.
5. **Use `valueFrom` references** for `compartmentId` and `subnetId` — maintains dependency ordering in infra charts.
6. **Configure VCN route tables** — the firewall only inspects traffic that is routed through it. Update subnet route tables to point to the firewall's IPv4 address.
