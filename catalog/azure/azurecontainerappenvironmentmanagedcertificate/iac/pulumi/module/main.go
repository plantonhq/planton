package module

import (
	"github.com/pkg/errors"
	azurecontainerappenvironmentmanagedcertificatev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironmentmanagedcertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// domainControlValidationWireValues maps the spec's validation enum value
// names to Azure's wire vocabulary.
var domainControlValidationWireValues = map[string]string{
	"HTTP":  "HTTP",
	"CNAME": "CNAME",
}

// Resources provisions an Azure-managed TLS certificate on a Container
// App Environment -- free, domain-validated, renewed by Azure.
//
// Lifecycle notes worth knowing before operating this resource:
//   - CREATE BLOCKS ON DOMAIN VALIDATION: Azure only completes the
//     operation once it has proven you control subject_name, so the
//     deployment polls until the required DNS records resolve publicly
//     (the asuid TXT record with the app's custom_domain_verification_id,
//     plus the CNAME/HTTP routing record) -- or fails around the
//     30-minute mark. Publish the records BEFORE deploying this resource.
//   - Only tags update in place; every other change re-issues the
//     certificate.
//   - Azure attaches the issued certificate to the matching
//     AzureContainerAppCustomDomain binding asynchronously -- the binding
//     deploys certificate-less first, then Azure fills it in.
func Resources(ctx *pulumi.Context, iacInput *azurecontainerappenvironmentmanagedcertificatev1alpha1.AzureContainerAppEnvironmentManagedCertificateIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppEnvironmentManagedCertificate.Spec

	certificateArgs := &containerapp.EnvironmentManagedCertificateArgs{
		Name:                      pulumi.String(spec.CertificateName),
		ContainerAppEnvironmentId: pulumi.String(locals.EnvironmentId),
		SubjectName:               pulumi.String(spec.SubjectName),
		Tags:                      pulumi.ToStringMap(locals.AzureTags),
	}

	// Unspecified deploys HTTP -- Azure's own default; sending it
	// explicitly keeps both engines identical on iac-input paths.
	if wire, ok := domainControlValidationWireValues[spec.DomainControlValidation.String()]; ok {
		certificateArgs.DomainControlValidation = pulumi.String(wire)
	} else {
		certificateArgs.DomainControlValidation = pulumi.String("HTTP")
	}

	createdCertificate, err := containerapp.NewEnvironmentManagedCertificate(ctx,
		spec.CertificateName,
		certificateArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create container app environment managed certificate %s", spec.CertificateName)
	}

	ctx.Export(OpCertificateId, createdCertificate.ID())
	ctx.Export(OpValidationToken, createdCertificate.ValidationToken)

	return nil
}
