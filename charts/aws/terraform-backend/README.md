# AWS Terraform State Backend InfraChart

The AWS Terraform State Backend InfraChart provides a streamlined way to rapidly provision cloud resources necessary to
securely store and manage Terraform state files. This chart specifically focuses on creating an AWS S3 bucket and
DynamoDB table essential for Terraform state storage and locking, ensuring ease of use and robust infrastructure
management for development and operations teams.

The resources defined by this chart are available in the [templates](templates) folder. Configuration parameters are
managed through [values.yaml](values.yaml).

---

## Included Cloud Resources

This Terraform State Backend Chart creates the following AWS resources:

1. **AWS S3 Bucket**:

    * Secure storage bucket for Terraform state files
    * Provisioned in the AWS `us-east-1` region

2. **AWS DynamoDB Table**:

    * Provides state locking mechanism for concurrent Terraform operations
    * Uses provisioned billing mode
    * Defined with a primary key named `LockID`

---

## Chart Input Values

The following values must be provided or will default as specified in [values.yaml](values.yaml):

| Input Parameter       | Description                                            | Example                        | Required/Default |
|-----------------------|--------------------------------------------------------|--------------------------------|------------------|
| `s3-bucket-name`      | Name of the S3 bucket for Terraform state backend      | my-org-terraform-state-backend | Required         |
| `dynamodb-table-name` | Name of the DynamoDB table for Terraform state locking | my-org-tf-state-lock           | Required         |

---

## Chart Customization

Resources created by this Terraform State Backend Chart can be customized post-deployment to fit specific requirements.
Individual configurations may be modified directly in your AWS account or through Planton.

---

## Important Notes

* The Terraform State Backend Chart is intended for initial provisioning only. Subsequent management or updates of the
  resources must be handled individually.
* Verify access permissions and security configurations for both the S3 bucket and DynamoDB table to ensure secure
  operation.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
