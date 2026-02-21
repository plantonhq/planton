# Examples

## Web Tier -- Allow HTTP and HTTPS

A security group for web-facing instances that allows HTTP (80) and HTTPS (443) from the internet, with unrestricted outbound access.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudSecurityGroup
metadata:
  name: web-tier-sg
  env: production
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  securityGroupName: web-tier
  description: Security group for web-facing instances
  tags:
    tier: web
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "80/80"
      cidrIp: "0.0.0.0/0"
      description: Allow HTTP from anywhere
    - type: ingress
      ipProtocol: tcp
      portRange: "443/443"
      cidrIp: "0.0.0.0/0"
      description: Allow HTTPS from anywhere
    - type: egress
      ipProtocol: all
      portRange: "-1/-1"
      cidrIp: "0.0.0.0/0"
      description: Allow all outbound traffic
```

## Database Tier -- Restrict to Internal VPC Traffic

A locked-down security group for database instances that only allows connections from the internal VPC CIDR range on specific database ports.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudSecurityGroup
metadata:
  name: db-tier-sg
  env: production
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  securityGroupName: db-tier
  description: Security group for database instances
  innerAccessPolicy: Drop
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "3306/3306"
      cidrIp: "10.0.0.0/8"
      priority: 1
      description: Allow MySQL from VPC
    - type: ingress
      ipProtocol: tcp
      portRange: "5432/5432"
      cidrIp: "10.0.0.0/8"
      priority: 2
      description: Allow PostgreSQL from VPC
    - type: ingress
      ipProtocol: tcp
      portRange: "6379/6379"
      cidrIp: "10.0.0.0/8"
      priority: 3
      description: Allow Redis from VPC
```

## SG-to-SG Reference -- Application Tier

A security group that allows traffic from another security group (the web tier) on a specific application port, demonstrating cross-SG references.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudSecurityGroup
metadata:
  name: app-tier-sg
  env: production
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123def456
  securityGroupName: app-tier
  description: Security group for application instances
  rules:
    - type: ingress
      ipProtocol: tcp
      portRange: "8080/8080"
      sourceSecurityGroupId: sg-web-tier-id
      description: Allow app traffic from web tier SG
    - type: ingress
      ipProtocol: tcp
      portRange: "22/22"
      cidrIp: "10.0.0.0/8"
      priority: 10
      description: Allow SSH from VPC
    - type: egress
      ipProtocol: all
      portRange: "-1/-1"
      cidrIp: "0.0.0.0/0"
      description: Allow all outbound traffic
```
