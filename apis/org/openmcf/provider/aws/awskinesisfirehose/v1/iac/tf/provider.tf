terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "= 5.82.0"
    }
  }
}

variable "aws_region" {
  description = "AWS region for the provider."
  type        = string
  default     = "us-east-1"
}

variable "aws_access_key" {
  description = "AWS access key ID. Leave empty to use environment credentials."
  type        = string
  default     = ""
}

variable "aws_secret_key" {
  description = "AWS secret access key. Leave empty to use environment credentials."
  type        = string
  default     = ""
  sensitive   = true
}

variable "aws_session_token" {
  description = "AWS session token for temporary credentials. Leave empty when not using STS."
  type        = string
  default     = ""
  sensitive   = true
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key != "" ? var.aws_access_key : null
  secret_key = var.aws_secret_key != "" ? var.aws_secret_key : null
  token      = var.aws_session_token != "" ? var.aws_session_token : null
}
