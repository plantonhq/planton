# AwsIamUser

AWS IAM users provide long-lived programmatic credentials for CI/CD pipelines, third-party integrations, and service accounts that require AWS API access. This resource creates an IAM user with managed or inline policies and optional access keys, following modern security best practices by focusing on service account use cases rather than human console access.

## Spec fields (summary)
- user_name: IAM username (1-64 characters, alphanumeric plus `+=,.@_-`; renamable in place)
- path: IAM path for organizational grouping (defaults to "/")
- managed_policy_arns: Managed policies to attach — references to AwsIamPolicy resources or literal ARNs (how AWS-managed policies attach)
- inline_policies: Map of policy name to JSON document for user-specific permissions
- permissions_boundary: Managed policy (reference or literal ARN) capping the user's maximum permissions — valuable on long-lived credentials
- disable_access_keys: Set to true to create user without access keys (default: false, creates access keys)
- force_destroy: Delete out-of-band credentials (login profile, extra keys, MFA devices) on teardown instead of failing

## Stack outputs
- user_arn: Amazon Resource Name (ARN) of the created IAM user
- user_name: Name of the IAM user in AWS
- user_id: Stable unique ID of the IAM user
- access_key_id: Access key ID for programmatic access (if keys were created)
- secret_access_key: Base64-encoded secret access key (if keys were created, marked as sensitive)
- console_url: AWS console sign-in URL

## How it works
This resource is orchestrated by the Planton CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes IaC backends in this repo:
- Pulumi (Go modules under iac/pulumi)
- Terraform (modules under iac/tf)

Access keys are created by default (since programmatic access is the primary use case). The secret key is automatically encrypted by Pulumi and should be immediately stored in a secret manager (AWS Secrets Manager, HashiCorp Vault, etc.) and not committed to version control.

## Common use cases
- **CI/CD service accounts**: Automated pipelines that deploy infrastructure or push artifacts
- **Third-party integrations**: External services requiring AWS API access (monitoring tools, backup services, etc.)
- **Legacy application service accounts**: Apps running outside AWS that haven't migrated to roles or federation
- **Cross-account automation**: Service accounts that assume roles in other AWS accounts

## Security best practices
- **Use IAM roles instead when possible**: For workloads running in AWS (EC2, Lambda, ECS), use IAM roles instead of users
- **Rotate access keys regularly**: Set up automated key rotation (90 days is a common standard)
- **Principle of least privilege**: Grant only the minimum permissions required
- **Monitor usage**: Use CloudTrail to track API calls and Access Advisor to identify unused permissions
- **Store secrets securely**: Never commit access keys to Git; use secret managers immediately after creation
- **Consider IAM Identity Center**: For human users, use AWS IAM Identity Center (SSO) instead of IAM users

## When NOT to use IAM users
- **Human users**: Use AWS IAM Identity Center or federated identities (SAML, OIDC) instead
- **AWS workloads**: Use IAM roles for EC2, Lambda, ECS, and other AWS services
- **Modern CI/CD**: Prefer OIDC federation (GitHub Actions, GitLab CI, etc.) over long-lived credentials

## References
- AWS IAM Users: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users.html
- Access keys best practices: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
- IAM best practices: https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html
- Access key rotation: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#Using_RotateAccessKey
- Research documentation: [docs/README.md](docs/README.md)
- Examples: [examples.md](examples.md)

