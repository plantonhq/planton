package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificatemanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certManagerCert provisions one Certificate Manager certificate. Exactly
// one arm is configured (enforced pre-deploy):
//   - managed: Google provisions and renews automatically. Domain control
//     is proven via referenced DNS authorizations, a private-PKI issuance
//     config, or load-balancer authorization when neither is set.
//   - self_managed: uploaded PEM chain + private key; rotation is an
//     in-place update with new material.
func certManagerCert(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) error {
	spec := locals.GcpCertManagerCert.Spec

	args := &certificatemanager.CertificateArgs{
		Name:   pulumi.String(locals.CertName),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.Location != "" {
		args.Location = pulumi.StringPtr(spec.Location)
	}

	if spec.Scope != "" {
		args.Scope = pulumi.StringPtr(spec.Scope)
	}

	if spec.Managed != nil {
		managedArgs := &certificatemanager.CertificateManagedArgs{
			Domains: pulumi.ToStringArray(spec.Managed.Domains),
		}
		if len(spec.Managed.DnsAuthorizations) > 0 {
			authorizations := pulumi.StringArray{}
			for _, authorization := range spec.Managed.DnsAuthorizations {
				authorizations = append(authorizations, pulumi.String(authorization.GetValue()))
			}
			managedArgs.DnsAuthorizations = authorizations
		}
		if spec.Managed.IssuanceConfig.GetValue() != "" {
			managedArgs.IssuanceConfig = pulumi.StringPtr(spec.Managed.IssuanceConfig.GetValue())
		}
		args.Managed = managedArgs
	}

	if spec.SelfManaged != nil {
		args.SelfManaged = &certificatemanager.CertificateSelfManagedArgs{
			PemCertificate: pulumi.StringPtr(spec.SelfManaged.PemCertificate),
			// Secret material: ToSecret marks it encrypted in Pulumi state —
			// the same handling GcpSslCertificate gives its private key.
			PemPrivateKey: pulumi.ToSecret(pulumi.String(spec.SelfManaged.PemPrivateKey)).(pulumi.StringOutput),
		}
	}

	// Unset defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdCertificate, err := certificatemanager.NewCertificate(ctx,
		locals.GcpCertManagerCert.Metadata.Name,
		args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	ctx.Export(OpCertificateId, createdCertificate.ID())
	ctx.Export(OpCertificateName, createdCertificate.Name)
	ctx.Export(OpSanDnsnames, createdCertificate.SanDnsnames)

	// The location output mirrors the spec with the "global" default —
	// identical derivation to the Terraform module.
	location := spec.Location
	if location == "" {
		location = "global"
	}
	ctx.Export(OpLocation, pulumi.String(location))

	// The managed block is a computed optional struct: match the bridged
	// field's pointer-ness and degrade to "" on the self-managed arm — the
	// optional-output export contract, mirrored by try(..., "") on
	// Terraform.
	ctx.Export(OpManagedState, createdCertificate.Managed.ApplyT(func(managed *certificatemanager.CertificateManaged) string {
		if managed == nil || managed.State == nil {
			return ""
		}
		return *managed.State
	}).(pulumi.StringOutput))

	return nil
}
