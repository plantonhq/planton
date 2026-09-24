package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificateTemplate creates the template. Everything but the location and
// ID updates in place.
func certificateTemplate(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPrivateCaCertificateTemplate.Spec
	resourceName := locals.GcpPrivateCaCertificateTemplate.Metadata.Name

	// Enable the Certificate Authority Service API first so a fresh project
	// works on the first deploy. DisableOnDestroy stays false: tearing down
	// one template must never disable the API for every pool, template, and
	// certificate in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("privateca.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcppcat-privateca.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable privateca.googleapis.com api")
	}

	args := &certificateauthority.CertificateTemplateArgs{
		Location:        pulumi.String(spec.Location),
		Name:            pulumi.String(locals.TemplateId),
		Description:     stringPtr(spec.Description),
		MaximumLifetime: stringPtr(spec.MaximumLifetime),
		Labels:          pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.PredefinedValues != nil {
		args.PredefinedValues = x509Parameters(spec.PredefinedValues)
	}
	if identity := spec.IdentityConstraints; identity != nil {
		identityArgs := &certificateauthority.CertificateTemplateIdentityConstraintsArgs{
			AllowSubjectPassthrough:         pulumi.Bool(identity.AllowSubjectPassthrough),
			AllowSubjectAltNamesPassthrough: pulumi.Bool(identity.AllowSubjectAltNamesPassthrough),
		}
		if expression := identity.CelExpression; expression != nil {
			identityArgs.CelExpression = &certificateauthority.CertificateTemplateIdentityConstraintsCelExpressionArgs{
				Expression:  stringPtr(expression.Expression),
				Title:       stringPtr(expression.Title),
				Description: stringPtr(expression.Description),
				Location:    stringPtr(expression.Location),
			}
		}
		args.IdentityConstraints = identityArgs
	}
	if passthrough := spec.PassthroughExtensions; passthrough != nil {
		passthroughArgs := &certificateauthority.CertificateTemplatePassthroughExtensionsArgs{
			KnownExtensions: stringArray(passthrough.KnownExtensions),
		}
		if len(passthrough.AdditionalExtensions) > 0 {
			additional := certificateauthority.CertificateTemplatePassthroughExtensionsAdditionalExtensionArray{}
			for _, oid := range passthrough.AdditionalExtensions {
				additional = append(additional, &certificateauthority.CertificateTemplatePassthroughExtensionsAdditionalExtensionArgs{
					ObjectIdPaths: objectIdPath(oid),
				})
			}
			passthroughArgs.AdditionalExtensions = additional
		}
		args.PassthroughExtensions = passthroughArgs
	}

	// Engine-side destroy stance: DELETE (the provider's default), PREVENT,
	// or ABANDON. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := certificateauthority.NewCertificateTemplate(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate template")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpTemplateId, created.Name)
	return nil
}
