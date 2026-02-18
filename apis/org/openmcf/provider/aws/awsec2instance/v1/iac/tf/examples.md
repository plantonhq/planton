# AWS EC2 Instance Examples

Below are several examples demonstrating how to define an AWS EC2 Instance component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic EC2 Instance with SSM Access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: basic-ec2
spec:
  region: us-west-2
  instanceName: web-server-1
  amiId: ami-0123456789abcdef0
  instanceType: t3.small
  subnetId:
    value: subnet-aaa111
  securityGroupIds:
    - value: sg-000111222
  connectionMethod: SSM
  iamInstanceProfileArn:
    value: arn:aws:iam::123456789012:instance-profile/ssm
  rootVolumeSizeGb: 30
  tags:
    env: production
    app: web-server
```

This example creates a basic EC2 instance:
• Uses SSM for secure access without SSH keys.
• Requires IAM instance profile with SSM permissions.
• Deployed in a private subnet with security groups.
• Includes resource tagging for organization.

---

## EC2 Instance with SSH Access

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: ssh-ec2
spec:
  region: us-west-2
  instanceName: bastion-host
  amiId: ami-0123456789abcdef0
  instanceType: t3.micro
  subnetId:
    value: subnet-public-1a
  securityGroupIds:
    - value: sg-bastion-access
  connectionMethod: BASTION
  keyName: my-keypair
  rootVolumeSizeGb: 20
  disableApiTermination: true
  tags:
    env: production
    role: bastion
```

This example uses traditional SSH access:
• Requires an existing EC2 key pair.
• Suitable for bastion hosts or direct SSH access.
• Includes termination protection for safety.
• Deployed in a public subnet for external access.

---

## EC2 Instance with EC2 Instance Connect

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: instance-connect-ec2
spec:
  region: us-west-2
  instanceName: app-server
  amiId: ami-0123456789abcdef0
  instanceType: t3.medium
  subnetId:
    value: subnet-private-1a
  securityGroupIds:
    - value: sg-app-access
  connectionMethod: INSTANCE_CONNECT
  keyName: my-keypair
  rootVolumeSizeGb: 50
  ebsOptimized: true
  tags:
    env: staging
    app: application-server
```

This example uses EC2 Instance Connect:
• Temporary SSH key injection for secure access.
• Requires a key pair name (keys are injected on-demand).
• EBS optimized for better storage performance.
• Larger root volume for application data.

---

## EC2 Instance with User Data

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: userdata-ec2
spec:
  region: us-west-2
  instanceName: web-app
  amiId: ami-0123456789abcdef0
  instanceType: t3.large
  subnetId:
    value: subnet-private-1b
  securityGroupIds:
    - value: sg-web-access
  connectionMethod: SSM
  iamInstanceProfileArn:
    value: arn:aws:iam::123456789012:instance-profile/web-app
  rootVolumeSizeGb: 100
  userData: |
    #!/bin/bash
    yum update -y
    yum install -y httpd
    systemctl start httpd
    systemctl enable httpd
    echo "<h1>Hello from OpenMCF!</h1>" > /var/www/html/index.html
  tags:
    env: development
    app: web-application
```

This example includes user data for initialization:
• Automatically installs and configures Apache web server.
• Uses cloud-init script for post-launch configuration.
• Larger root volume for web application data.
• Includes comprehensive tagging.

---

## Production EC2 Instance

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: production-ec2
spec:
  region: us-west-2
  instanceName: production-app-server
  amiId: ami-0123456789abcdef0
  instanceType: m5.large
  subnetId:
    value: subnet-private-1a
  securityGroupIds:
    - value: sg-production-app
    - value: sg-monitoring
  connectionMethod: SSM
  iamInstanceProfileArn:
    value: arn:aws:iam::123456789012:instance-profile/production-app
  rootVolumeSizeGb: 200
  ebsOptimized: true
  disableApiTermination: true
  userData: |
    #!/bin/bash
    # Production server initialization
    yum update -y
    yum install -y docker
    systemctl start docker
    systemctl enable docker
  tags:
    env: production
    app: production-application
    cost-center: engineering
    owner: devops-team
```

This example is production-ready:
• Uses larger instance type for performance.
• Multiple security groups for different access patterns.
• EBS optimized and termination protection enabled.
• Comprehensive tagging for cost tracking and ownership.
• Docker installation via user data.

---

## Minimal EC2 Instance

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: minimal-ec2
spec:
  region: us-west-2
  instanceName: test-instance
  amiId: ami-0123456789abcdef0
  instanceType: t3.nano
  subnetId:
    value: subnet-aaa111
  securityGroupIds:
    - value: sg-000111222
  connectionMethod: SSM
  iamInstanceProfileArn:
    value: arn:aws:iam::123456789012:instance-profile/ssm
```

A minimal configuration with:
• Only required fields specified.
• Uses default root volume size (30GB).
• SSM access method for simplicity.
• No additional features or optimizations.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the EC2 instance is active via the AWS console or by
using the AWS CLI:

```shell
aws ec2 describe-instances --instance-ids <your-instance-id>
```

You should see your new EC2 instance with its configuration details, including private IP, availability zone, and status.
For SSM access, use the AWS Systems Manager Session Manager to connect. For SSH access, use your key pair to connect directly or via a bastion host.


