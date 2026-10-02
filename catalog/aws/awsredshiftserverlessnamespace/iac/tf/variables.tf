variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsRedshiftServerlessNamespace specification"
  type = object({
    # The AWS region the namespace is created in. Workgroups that attach
    # to this namespace must live in the same region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The name of the first database created in the namespace. Empty
    # keeps the AWS default ("dev"). Create-time only -- changing it
    # replaces the namespace (and the data in it); additional databases
    # are created with SQL, not here.
    db_name = optional(string, "")

    # The admin username for the first database. Empty keeps the AWS
    # default ("admin"). Unlike the provisioned Redshift cluster, a
    # serverless namespace does not hard-require an admin user at create
    # time -- IAM identities can use temporary credentials
    # (GetCredentials) without one.
    admin_username = optional(string, "")

    # Let AWS manage the admin password in Secrets Manager: AWS generates
    # it, stores it, rotates it on schedule, and no secret ever touches
    # this manifest or the IaC state. The managed secret's ARN is
    # exported as the admin_password_secret_arn output. Mutually
    # exclusive with admin_user_password -- and the recommended posture.
    manage_admin_password = optional(bool, false)

    # The admin password, supplied directly (8-64 chars with at least one
    # uppercase letter, one lowercase letter, and one digit). Stored in
    # IaC state -- prefer manage_admin_password, which keeps the secret
    # in Secrets Manager entirely. Mutually exclusive with
    # manage_admin_password.
    admin_user_password = optional(string, "")

    # The KMS key that encrypts the Secrets Manager secret holding the
    # managed admin password. Empty uses the AWS-managed
    # aws/secretsmanager key. Only meaningful with
    # manage_admin_password. Reference an AwsKmsKey key_arn output or
    # pass a literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    admin_password_secret_kms_key_id = optional(string, "")

    # The KMS key that encrypts the namespace's stored data. Empty uses
    # the AWS-owned Redshift service key. Reference an AwsKmsKey key_arn
    # output or pass a literal key ARN. Switching keys on a live
    # namespace is an in-place but long-running re-encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # IAM roles the serverless engine assumes to access other AWS
    # services during COPY, UNLOAD, CREATE EXTERNAL FUNCTION, and
    # Redshift Spectrum queries (S3, DynamoDB, Glue, Lambda, ...).
    # Reference AwsIamRole role_arn outputs or pass literal role ARNs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    iam_roles = optional(list(string), [])

    # The IAM role assumed when a SQL command does not name one
    # explicitly (e.g. COPY ... IAM_ROLE default). Must also be present
    # in iam_roles -- AWS rejects a default role it has not been given.
    # Reference an AwsIamRole role_arn output or pass a literal role ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    default_iam_role_arn = optional(string, "")

    # Which audit log types the namespace exports to CloudWatch Logs:
    # "connectionlog" (connection attempts), "useractivitylog" (every
    # executed query), "userlog" (user create/alter/drop events). Empty
    # exports nothing.
    log_exports = optional(list(string), [])
  })
}
