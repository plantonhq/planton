# Core EC2 instance and related resources for AwsEc2Instance
#
# This Terraform configuration provisions a single EC2 instance following
# production best practices: private subnet deployment, controlled security
# groups, IAM instance profiles, and modern access patterns (SSM or SSH).
#
# The configuration enforces connection method requirements:
# - SSM: Requires IAM instance profile with SSM permissions (AmazonSSMManagedInstanceCore)
# - BASTION/INSTANCE_CONNECT: Requires SSH key pair (key_name)
#
# Network Placement:
# - Deployed in a private subnet (no public IP by default)
# - Security groups control inbound/outbound traffic
# - Subnet must have route to NAT Gateway or VPC endpoints for AWS API access
#
# Storage Configuration:
# - Root volume uses gp3 (modern default), size configurable
# - Additional EBS volumes can be added post-creation or via user_data scripts
#
# IAM Instance Profile:
# - When using SSM connection method, the instance profile is attached automatically
# - The profile must exist and include the necessary IAM role with SSM permissions
# - Extracted from the ARN provided in var.spec.iam_instance_profile_arn

resource "aws_instance" "this" {
  # Amazon Machine Image - region-specific, must be valid in target AWS region
  # Example: ami-0abcdef1234567890 for Amazon Linux 2023 or Ubuntu LTS
  ami = var.spec.ami_id

  # Instance type determines vCPU count and memory allocation
  # Common choices: t3.micro (2 vCPUs, 1 GiB), t3.small (2 vCPUs, 2 GiB),
  # m5.large (2 vCPUs, 8 GiB), c5.xlarge (4 vCPUs, 8 GiB)
  instance_type = var.spec.instance_type

  # Target subnet (typically private) within existing VPC
  # Private subnets have no direct internet access, requiring NAT Gateway or VPC endpoints
  subnet_id = try(var.spec.subnet_id.value, null)

  # Security groups control network access (inbound/outbound firewall rules)
  # At least one security group is required; multiple can be attached
  vpc_security_group_ids = local.security_group_ids_values

  # SSH key pair name - required for BASTION and INSTANCE_CONNECT methods
  # Not needed for SSM (Session Manager uses IAM authentication)
  # Conditional logic ensures key is only attached when connection method requires it
  key_name = local.needs_key_name ? try(var.spec.key_name, null) : null

  # IAM instance profile (role) - required for SSM, optional otherwise
  # Profile name is extracted from ARN in locals.tf
  # Grants the instance permissions to AWS APIs (SSM, S3, CloudWatch, etc.)
  iam_instance_profile = local.use_ssm ? local.iam_instance_profile_name : null

  # EBS optimization improves storage performance for instance types that support it
  # Recommended for production workloads with high disk I/O
  ebs_optimized = try(var.spec.ebs_optimized, null)

  # Termination protection prevents accidental instance deletion via API/console
  # Useful for long-lived production instances (databases, application servers)
  disable_api_termination = try(var.spec.disable_api_termination, null)

  # Root volume configuration (boot disk)
  # Default volume type is gp3 (AWS default), size is configurable (default: 30 GiB)
  # gp3 offers better price/performance than older gp2 volumes
  root_block_device {
    volume_size = var.spec.root_volume_size_gb
    # Additional root volume settings (volume_type, iops, throughput) can be added here
    # if needed for specific workloads requiring higher disk performance
  }

  # User data script for instance initialization (cloud-init on Linux)
  # Runs once at first boot - use for minimal bootstrap tasks
  # For complex configuration, prefer baking into AMI with Packer
  user_data = try(var.spec.user_data, null)

  # Tags for resource identification and cost tracking
  # Merges user-provided tags with Name tag (instance name)
  tags = merge({
    Name = var.spec.instance_name
  }, local.safe_tags)

  # Lifecycle rules (optional, can be uncommented if needed):
  # lifecycle {
  #   # Prevent accidental replacement due to AMI ID changes
  #   ignore_changes = [ami]
  #
  #   # Require explicit confirmation before destroying long-lived instances
  #   prevent_destroy = false
  # }
}


