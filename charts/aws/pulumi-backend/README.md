# AWS Pulumi State Backend InfraChart

The AWS Pulumi State Backend InfraChart provides a streamlined approach to rapidly provision cloud resources essential
for securely storing and managing Pulumi state files. This chart specifically creates an AWS S3 bucket to act as the
backend storage for Pulumi states, facilitating ease of use and efficient infrastructure management for development and
operations teams.

The resources defined by this chart are available in the [templates](templates) folder. Configuration parameters are
managed through [values.yaml](values.yaml).

---

## Included Cloud Resources

This Pulumi State Backend Chart creates the following AWS resource:

1. **AWS S3 Bucket**:

    * Secure storage bucket dedicated to Pulumi state files
    * Provisioned in the AWS `us-east-1` region

---

## Chart Input Values

The following values must be provided or will default as specified in [values.yaml](values.yaml):

| Input Parameter | Description                                            | Example                     | Required/Default |
|-----------------|--------------------------------------------------------|-----------------------------|------------------|
| `bucket_name`   | Name of the S3 bucket for Pulumi state backend storage | my-org-pulumi-state-backend | Required         |

---

## Chart Customization

The AWS S3 Bucket created by this Pulumi State Backend Chart can be customized post-deployment to meet specific
requirements. Individual configurations may be modified directly in your AWS account or through Planton.

---

## Important Notes

* The Pulumi State Backend Chart is intended for initial provisioning only. Subsequent management or updates of the
  bucket must be handled individually.
* Verify access permissions and security settings for the S3 bucket to ensure secure and efficient operation.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
