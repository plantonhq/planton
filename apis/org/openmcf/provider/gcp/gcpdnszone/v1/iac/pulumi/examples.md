Here are several examples for the `GcpDnsZone` API resource, similar to how you've formatted the examples for the `MicroserviceKubernetes` API. These examples showcase different configurations for creating and managing DNS zones and records in Google Cloud using Planton.

### Example 1: Basic Google Cloud DNS Zone

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsZone
metadata:
  name: example-com
spec:
  projectId: my-gcp-project
  records:
    - record_type: A
      name: www.example.com.
      values:
        - 192.0.2.1
      ttl_seconds: 300
```

### Example 2: Google Cloud DNS Zone with Multiple Records

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsZone
metadata:
  name: example-com
spec:
  projectId: my-gcp-project
  records:
    - record_type: A
      name: www.example.com.
      values:
        - 192.0.2.1
      ttl_seconds: 300
    - record_type: CNAME
      name: api.example.com.
      values:
        - www.example.com.
      ttl_seconds: 3600
    - record_type: MX
      name: example.com.
      values:
        - "10 mail.example.com."
      ttl_seconds: 3600
```

### Example 3: Google Cloud DNS Zone with IAM Service Accounts

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsZone
metadata:
  name: prod-example-com
spec:
  projectId: prod-gcp-project
  iam_service_accounts:
    - service-account1@example-project.iam.gserviceaccount.com
    - service-account2@example-project.iam.gserviceaccount.com
  records:
    - record_type: A
      name: prod.example.com.
      values:
        - 203.0.113.1
      ttl_seconds: 300
    - record_type: TXT
      name: prod.example.com.
      values:
        - "v=spf1 include:_spf.google.com ~all"
      ttl_seconds: 600
```

### Example 4: Minimal Google Cloud DNS Zone

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpDnsZone
metadata:
  name: dev-example-com
spec:
  projectId: dev-gcp-project
```