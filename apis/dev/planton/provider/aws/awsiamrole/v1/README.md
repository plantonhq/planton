# AwsIamRole

AWS IAM roles enable secure delegation of permissions to AWS services, applications, or users through temporary credentials without embedding long-lived access keys. This resource defines an IAM role with its trust policy (who can assume it), managed policy attachments, and inline policies, providing a production-ready abstraction for role-based access control.

## Spec fields (summary)
- description: Optional human-readable description of the role's purpose
- path: IAM path for organizational grouping (defaults to "/")
- trust_policy: JSON document defining who can assume this role (principals, actions, conditions)
- managed_policy_arns: List of ARNs for AWS-managed or customer-managed policies to attach
- inline_policies: Map of policy name to JSON document for role-specific permissions embedded directly in the role

## Stack outputs
- role_arn: Amazon Resource Name (ARN) of the created IAM role
- role_name: Name of the IAM role in AWS

## How it works
This resource is orchestrated by the Planton CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes IaC backends in this repo:
- Pulumi (Go modules under iac/pulumi)
- Terraform (modules under iac/tf)

The trust policy controls **who** can assume the role, while permissions policies (managed or inline) control **what** the role can do once assumed.

## Common use cases
- **Lambda execution roles**: Allow Lambda functions to access AWS services (S3, DynamoDB, etc.)
- **ECS task roles**: Grant containerized applications permissions to AWS resources
- **EC2 instance roles**: Enable EC2 instances to securely access AWS APIs
- **Cross-account access**: Allow principals from another AWS account to access resources
- **Service-to-service delegation**: Let one AWS service act on behalf of another

## Security best practices
- **Least privilege**: Grant only the minimum permissions required
- **Specific trust policies**: Never use wildcard principals; always specify exact service principals or account ARNs
- **Add conditions**: Use `aws:SourceAccount`, `aws:SourceArn`, or `sts:ExternalId` to prevent confused deputy attacks
- **Prefer managed policies**: For reusability and centralized version control
- **Monitor usage**: Use CloudTrail to track AssumeRole calls and Access Advisor to identify unused permissions

## References
- AWS IAM Roles: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html
- Trust policies: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_terms-and-concepts.html
- AssumeRole: https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html
- Policy evaluation logic: https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html
- Confused deputy problem: https://docs.aws.amazon.com/IAM/latest/UserGuide/confused-deputy.html
- Research documentation: [docs/README.md](docs/README.md)
- Examples: [examples.md](examples.md)
