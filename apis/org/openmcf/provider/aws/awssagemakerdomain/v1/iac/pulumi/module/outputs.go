package module

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sagemaker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	OpDomainId                         = "domain_id"
	OpDomainArn                        = "domain_arn"
	OpDomainUrl                        = "domain_url"
	OpHomeEfsFileSystemId              = "home_efs_file_system_id"
	OpSecurityGroupIdForDomainBoundary = "security_group_id_for_domain_boundary"
	OpSingleSignOnApplicationArn       = "single_sign_on_application_arn"
)

func outputs(ctx *pulumi.Context, createdDomain *sagemaker.Domain) {
	ctx.Export(OpDomainId, createdDomain.ID())
	ctx.Export(OpDomainArn, createdDomain.Arn)
	ctx.Export(OpDomainUrl, createdDomain.Url)
	ctx.Export(OpHomeEfsFileSystemId, createdDomain.HomeEfsFileSystemId)
	ctx.Export(OpSecurityGroupIdForDomainBoundary, createdDomain.SecurityGroupIdForDomainBoundary)
	// Only populated when auth_mode is SSO; empty string for IAM mode
	ctx.Export(OpSingleSignOnApplicationArn, createdDomain.SingleSignOnApplicationArn)
}
