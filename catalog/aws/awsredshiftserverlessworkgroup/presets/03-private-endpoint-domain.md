# Governed Workgroup with Private Endpoints and a Custom Domain

This preset creates a production serverless workgroup whose spend, reachability, and identity are governed declaratively: usage limits deactivate queries when daily RPU-hours exceed the cap and log cross-region datasharing transfer, a VPC endpoint exposes the workgroup inside a separate BI-tooling VPC without peering, and a custom DNS name fronts the endpoint with TLS from an ACM certificate.

## When to Use

- Serverless warehouses that need a hard daily compute budget (deactivate stops queries until the period resets -- data is untouched)
- Hub-and-spoke networks where consumers live in their own VPC and peering is off the table
- Client tooling that must connect to a stable branded hostname (`warehouse.example.com`) instead of the AWS-generated endpoint

## Key Configuration Choices

- **Fixed baseline with a cap** (`baseCapacity: 32`, `maxCapacity: 128`) -- Predictable query performance with bounded worst-case spend
- **Daily compute deactivation** (`usageLimits: serverless-compute / 200 RPU-hours / deactivate`) -- The serverless breach vocabulary is `deactivate` (provisioned clusters use `disable`); queries stop when the cap is hit and resume when the period resets
- **Datasharing transfer log** (`usageLimits: cross-region-datasharing / 5 TB / log`) -- Visibility without blocking
- **Cross-VPC endpoint** (`endpointAccesses` with its own `subnetIds`) -- The endpoint lands in the CONSUMING VPC's subnets; omit the entry's subnetIds to reuse the workgroup's own. The private address surfaces in the `endpoint_access_addresses` output keyed by endpoint name
- **Custom domain** (`customDomain`) -- One per workgroup (AWS's model); the certificate must cover the domain and live in the workgroup's region. You own the CNAME pointing the domain at the workgroup endpoint; ACM renewals update the association's expiry in place (the `custom_domain_certificate_expiry_time` output)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<namespace-name>` | The namespace (data plane) this workgroup serves | `AwsRedshiftServerlessNamespace` status outputs |
| `<private-subnet-id-az1>` / `<private-subnet-id-az2>` / `<private-subnet-id-az3>` | Three subnets in three distinct AZs for the workgroup | AWS VPC console or `AwsSubnet` status outputs |
| `<consumer-subnet-id-az1>` / `<consumer-subnet-id-az2>` / `<consumer-subnet-id-az3>` | Subnets in the CONSUMING VPC for the endpoint | AWS VPC console or `AwsSubnet` status outputs |
| `<acm-certificate-arn>` | ACM certificate covering warehouse.example.com in the workgroup's region | AWS ACM console or `AwsCertManagerCert` status outputs |

## Related Presets

- **01-capped-dev** -- Minimal dev workgroup with a low RPU cap
- **02-price-performance-production** -- AWS-owned capacity baseline against a price-performance dial
