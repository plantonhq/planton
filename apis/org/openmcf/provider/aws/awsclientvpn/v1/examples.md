# AWS Client VPN Examples

Create a YAML manifest using one of the examples below. After the YAML is created, apply it with OpenMCF:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

Or, if using Terraform:

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

---

## Minimal Example (Certificate Auth)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: minimal-vpn
spec:
  region: us-east-1
  vpcId:
    value: vpc-12345678
  subnets:
    - value: subnet-abc123
  clientCidrBlock: 10.0.0.0/22
  authenticationType: certificate
  serverCertificateArn:
    value: arn:aws:acm:us-east-1:123456789012:certificate/abc-123-def
  cidrAuthorizationRules:
    - 10.0.0.0/16
```

This minimal example:
- Creates a Client VPN endpoint in the specified VPC and subnet.
- Allocates client IPs from 10.0.0.0/22.
- Uses certificate-based authentication (mutual TLS).
- Authorizes clients to access the 10.0.0.0/16 network.
- Uses default settings: port 443, TCP protocol, split-tunnel enabled.

---

## Split-Tunnel VPN (Recommended for Remote Access)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: split-tunnel-vpn
spec:
  region: us-east-1
  description: Remote developer access to production VPC
  vpcId:
    value: vpc-prod123
  subnets:
    - value: subnet-private-a
    - value: subnet-private-b
  clientCidrBlock: 10.100.0.0/22
  authenticationType: certificate
  serverCertificateArn:
    value: arn:aws:acm:us-east-1:123456789012:certificate/vpn-cert
  cidrAuthorizationRules:
    - 10.0.0.0/16
    - 172.16.0.0/12
  disableSplitTunnel: false
  vpnPort: 443
  transportProtocol: tcp
  logGroupName: /aws/clientvpn/production
  dnsServers:
    - 10.0.0.2
```

This configuration:
- Enables split-tunnel mode (only internal traffic goes through VPN).
- Associates with two subnets for high availability across AZs.
- Authorizes access to multiple private CIDR ranges.
- Uses port 443 with TCP for maximum firewall compatibility.
- Enables connection logging to CloudWatch.
- Configures custom DNS server for VPC name resolution.

---

## Full-Tunnel VPN (Maximum Security)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: full-tunnel-vpn
spec:
  region: us-east-1
  description: Secure full-tunnel VPN for compliance requirements
  vpcId:
    value: vpc-secure123
  subnets:
    - value: subnet-dmz-a
    - value: subnet-dmz-b
  clientCidrBlock: 192.168.100.0/22
  authenticationType: certificate
  serverCertificateArn:
    value: arn:aws:acm:us-west-2:987654321098:certificate/secure-vpn
  cidrAuthorizationRules:
    - 0.0.0.0/0
  disableSplitTunnel: true
  vpnPort: 443
  transportProtocol: tcp
  securityGroups:
    - value: sg-vpn-clients
  logGroupName: /aws/clientvpn/secure
```

This configuration:
- Enables full-tunnel mode - all client traffic routes through VPN.
- Authorizes access to all networks (0.0.0.0/0).
- Applies custom security group for fine-grained traffic control.
- Enables connection logging for compliance and auditing.
- Associates with multiple subnets for redundancy.

---

## UDP VPN for Low Latency

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: udp-vpn
spec:
  region: us-east-1
  description: Low-latency VPN for real-time applications
  vpcId:
    value: vpc-apps123
  subnets:
    - value: subnet-apps-a
  clientCidrBlock: 10.200.0.0/22
  authenticationType: certificate
  serverCertificateArn:
    value: arn:aws:acm:us-east-1:123456789012:certificate/udp-vpn-cert
  cidrAuthorizationRules:
    - 10.0.0.0/8
  vpnPort: 1194
  transportProtocol: udp
  logGroupName: /aws/clientvpn/apps
```

This configuration:
- Uses UDP on port 1194 for lower latency (better for real-time apps).
- Note: UDP may be blocked by some corporate firewalls.
- Single subnet association for simplicity.
- Authorizes access to the entire 10.0.0.0/8 private range.

---

## Multi-AZ High Availability VPN

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: ha-vpn
spec:
  region: us-east-1
  description: High-availability VPN across three availability zones
  vpcId:
    value: vpc-ha123
  subnets:
    - value: subnet-az1
    - value: subnet-az2
    - value: subnet-az3
  clientCidrBlock: 10.50.0.0/22
  authenticationType: certificate
  serverCertificateArn:
    value: arn:aws:acm:us-east-1:123456789012:certificate/ha-vpn
  cidrAuthorizationRules:
    - 10.50.0.0/16
  vpnPort: 443
  transportProtocol: tcp
  securityGroups:
    - value: sg-vpn-ha
  logGroupName: /aws/clientvpn/ha
  dnsServers:
    - 10.50.0.2
    - 10.50.1.2
```

This configuration:
- Associates with three subnets across different Availability Zones.
- Ensures VPN remains available even if one AZ fails.
- Configures two DNS servers for redundancy.
- Enables connection logging for monitoring.

---

## Using Foreign Key References

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsClientVpn
metadata:
  name: vpn-with-refs
spec:
  region: us-east-1
  description: VPN with dynamic resource references
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-production-vpc
      field: status.outputs.vpc_id
  subnets:
    - valueFrom:
        kind: AwsVpc
        name: my-production-vpc
        field: status.outputs.private_subnet_ids[0]
    - valueFrom:
        kind: AwsVpc
        name: my-production-vpc
        field: status.outputs.private_subnet_ids[1]
  clientCidrBlock: 10.150.0.0/22
  authenticationType: certificate
  serverCertificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: vpn-server-cert
      field: status.outputs.cert_arn
  cidrAuthorizationRules:
    - 10.0.0.0/16
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: vpn-client-sg
        field: status.outputs.security_group_id
  vpnPort: 443
  transportProtocol: tcp
  logGroupName: /aws/clientvpn/production
```

This example demonstrates:
- Using foreign key references to other OpenMCF resources.
- Dynamically referencing VPC ID, subnet IDs, certificate ARN, and security group ID.
- This pattern promotes composable infrastructure definitions.
- All referenced resources must exist before applying this manifest.

---

## After Deploying

Once you've applied your manifest with OpenMCF, you can:

1. **Export client configuration**:
```shell
aws ec2 export-client-vpn-client-configuration \
  --client-vpn-endpoint-id <endpoint-id> \
  --output text > client-config.ovpn
```

2. **Verify endpoint status**:
```shell
aws ec2 describe-client-vpn-endpoints \
  --client-vpn-endpoint-ids <endpoint-id>
```

3. **View connection logs** (if logging enabled):
```shell
aws logs tail /aws/clientvpn/production --follow
```

Distribute the client configuration file to VPN users along with their client certificates and private keys. Users can import the configuration into an OpenVPN client (like AWS VPN Client, Tunnelblick, or OpenVPN Connect) to establish connections.

---

## Validation

Before deploying, validate your manifest:

```bash
openmcf validate --manifest ./aws-client-vpn.yaml
```

This checks the manifest against the protobuf schema and reports any validation errors.
