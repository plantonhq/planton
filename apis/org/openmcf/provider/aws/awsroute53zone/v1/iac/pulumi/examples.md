```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: example
spec: {}
```

CLI:

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

### Example 1: Basic Public Route53 Hosted Zone

This example sets up a basic public Route53 hosted zone for a domain. It provisions the DNS zone and outputs the nameservers for the domain, which can be used to update the domain registrar.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: public-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
```

---

### Example 2: Private Route53 Hosted Zone for VPC

This example provisions a private Route53 hosted zone associated with an AWS VPC. It restricts DNS resolution to the specified VPC, which is useful for internal services.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: private-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: internal.example.com
    private_zone: true
    vpc_associations:
      - vpc_id: vpc-123abc456def
        vpc_region: us-west-2
```

---

### Example 3: Adding A Record for Website

This example creates a public DNS zone and an A record for the domain `www.example.com`, which points to the IP address of a web server.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: website-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: A
      name: www.example.com
      ttl: 300
      values:
        - 192.0.2.44
```

---

### Example 4: CNAME Record for Subdomain

This example sets up a CNAME record for a subdomain (`app.example.com`) that points to another domain (`app-load-balancer.example.net`). This is commonly used for routing traffic to load balancers.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: subdomain-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: CNAME
      name: app.example.com
      ttl: 300
      values:
        - app-load-balancer.example.net
```

---

### Example 5: Alias Record for AWS Load Balancer

This example sets up an alias record for an AWS Elastic Load Balancer (ELB). An alias record is used to point to AWS resources without needing an IP address.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: elb-alias-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: A
      name: www.example.com
      alias:
        dns_name: my-load-balancer-1234567890.us-east-1.elb.amazonaws.com
        evaluate_target_health: true
        hosted_zone_id: Z35SXDOTRQ7X7K
```

---

### Example 6: Multi-Region DNS Setup with Failover Routing

This example provisions DNS records using a failover routing policy, directing traffic to different regions based on availability. The primary region is `us-east-1`, and the secondary region is `us-west-2`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: multi-region-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: A
      name: failover.example.com
      ttl: 60
      routing_policy:
        type: Failover
        primary:
          - value: 192.0.2.44
          health_check: true
        secondary:
          - value: 198.51.100.44
```

---

### Example 7: MX Record for Email Hosting

This example sets up an MX (Mail Exchange) record for `example.com`, directing email traffic to an external email server. MX records are essential for email configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: email-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: MX
      name: example.com
      ttl: 300
      values:
        - "10 mail.example.com"
```

---

### Example 8: TXT Record for Domain Verification

This example provisions a TXT record to verify domain ownership, which is commonly required by services like Google, AWS, and Microsoft for SSL or domain registration verification.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: domain-verification-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: TXT
      name: _acme-challenge.example.com
      ttl: 300
      values:
        - "verification-token-1234abcd"
```

---

### Example 9: SRV Record for Service Discovery

This example creates an SRV (Service) record for enabling service discovery within the domain, typically used for services like SIP, XMPP, or LDAP.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: service-discovery-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: SRV
      name: _sip._tcp.example.com
      ttl: 300
      values:
        - "10 5 5060 sipserver.example.com"
```

---

### Example 10: Delegating DNS to Subdomains

This example demonstrates how to delegate a subdomain (e.g., `sub.example.com`) to another set of nameservers.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRoute53Zone
metadata:
  name: subdomain-delegation-dns-zone
spec:
  region: us-east-1
  zone_config:
    domain_name: example.com
    private_zone: false
  records:
    - type: NS
      name: sub.example.com
      ttl: 300
      values:
        - ns-123.awsdns-45.org
        - ns-678.awsdns-89.co.uk
```

---

### Applying the Configuration

#### Prereq:
* Set [Pulumi Backend](https://www.pulumi.com/docs/iac/concepts/state-and-backends/#local-filesystem)
  * Local `pulumi login --local`
* Set Region `pulumi config set aws:region <region>
* Once the desired YAML file with the configuration is created, apply it using the following command:
  ```shell
  openmcf pulumi up --manifest <yaml-path> --stack <stack-path>
  ```

Refer to the example section for detailed usage instructions.
