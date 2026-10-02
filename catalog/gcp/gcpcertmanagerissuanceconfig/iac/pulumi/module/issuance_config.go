package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificatemanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// issuanceConfig provisions one certificate issuance config: how
// Google-managed certificates that name it are issued from a Certificate
// Authority Service pool. Every argument except labels and deletion_policy
// forces replacement.
func issuanceConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpCertManagerIssuanceConfig.Spec

	args := &certificatemanager.CertificateIssuanceConfigArgs{
		Name:                     pulumi.String(locals.IssuanceConfigName),
		Labels:                   pulumi.ToStringMap(locals.GcpLabels),
		KeyAlgorithm:             pulumi.String(spec.KeyAlgorithm),
		Lifetime:                 pulumi.String(spec.Lifetime),
		RotationWindowPercentage: pulumi.Int(int(spec.RotationWindowPercentage)),
		// The spec lifts the provider's two single-field wrappers into ca_pool.
		CertificateAuthorityConfig: &certificatemanager.CertificateIssuanceConfigCertificateAuthorityConfigArgs{
			CertificateAuthorityServiceConfig: &certificatemanager.CertificateIssuanceConfigCertificateAuthorityConfigCertificateAuthorityServiceConfigArgs{
				CaPool: pulumi.String(spec.CaPool.GetValue()),
			},
		},
	}
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}
	// Empty defers to the provider default ("global").
	if locals.Location != "" {
		args.Location = pulumi.StringPtr(locals.Location)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	// Unset defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdIssuanceConfig, err := certificatemanager.NewCertificateIssuanceConfig(ctx,
		locals.GcpCertManagerIssuanceConfig.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate issuance config")
	}

	ctx.Export(OpIssuanceConfigId, createdIssuanceConfig.ID())
	ctx.Export(OpIssuanceConfigName, createdIssuanceConfig.Name)

	// The location output reports the effective location, "global" when
	// unset -- identical derivation to the Terraform module.
	location := locals.Location
	if location == "" {
		location = "global"
	}
	ctx.Export(OpLocation, pulumi.String(location))

	return nil
}
