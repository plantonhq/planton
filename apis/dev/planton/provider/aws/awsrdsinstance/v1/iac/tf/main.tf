# AWS RDS Instance Terraform Module
#
# This module creates an AWS RDS DB Instance with optional subnet group.
# The actual resource definitions are organized in the resources/ subdirectory
# for better code organization:
#
#   - resources/instance.tf: RDS instance resource
#   - resources/subnet_group.tf: DB subnet group resource
#
# This structure separates concerns while maintaining a clean module interface.

# Note: Terraform automatically loads all .tf files in the current directory
# and subdirectories, so the resources defined in resources/*.tf are part of
# this module's configuration.

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

# The module creates the following resources:
# 1. aws_db_subnet_group (conditional, if subnet_ids provided)
# 2. aws_db_instance (always created)
#
# See provider.tf for AWS provider configuration
# See variables.tf for input variable definitions
# See outputs.tf for output value definitions
# See locals.tf for local value computations
# See resources/ for actual resource definitions

