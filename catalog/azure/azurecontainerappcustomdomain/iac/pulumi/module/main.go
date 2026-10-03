package module

import (
	"github.com/pkg/errors"
	azurecontainerappcustomdomainv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappcustomdomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/containerapp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificateBindingTypeWireValues maps the spec's binding-type enum
// value names to Azure's wire vocabulary -- the provider validates the
// mixed-case form case-sensitively.
var certificateBindingTypeWireValues = map[string]string{
	"SNI_ENABLED": "SniEnabled",
	"DISABLED":    "Disabled",
	"AUTO":        "Auto",
}

// Resources binds a custom domain to a Container App.
//
// Lifecycle notes worth knowing before operating this resource:
//   - CREATE BLOCKS ON DOMAIN VALIDATION: Azure verifies ownership of the
//     hostname during the operation, so the asuid TXT record (carrying the
//     app's custom_domain_verification_id) and the CNAME/A routing record
//     must resolve publicly BEFORE this resource deploys, or the create
//     fails. The app must have ingress enabled.
//   - Every field replaces the binding when changed -- Azure models it as
//     an entry in the app's ingress configuration with no update surface.
//   - MANAGED-CERTIFICATE FLOW (certificate fields unset): Azure attaches
//     the issued certificate to this binding asynchronously, out of band.
//     The module ignores drift on the certificate fields in that flow ONLY
//     -- applying the ignore unconditionally would swallow a legitimate
//     certificate change on a bring-your-own binding.
func Resources(ctx *pulumi.Context, iacInput *azurecontainerappcustomdomainv1alpha1.AzureContainerAppCustomDomainIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureContainerAppCustomDomain.Spec

	domainArgs := &containerapp.CustomDomainArgs{
		Name:           pulumi.String(spec.DomainName),
		ContainerAppId: pulumi.String(locals.ContainerAppId),
	}

	resourceOptions := []pulumi.ResourceOption{pulumi.Provider(azureProvider)}

	if locals.CertificateId != "" {
		// Bring-your-own flow: certificate + binding type travel together
		// (spec validation guarantees the pairing).
		domainArgs.ContainerAppEnvironmentCertificateId = pulumi.String(locals.CertificateId)
		domainArgs.CertificateBindingType = pulumi.String(certificateBindingTypeWireValues[spec.CertificateBindingType.String()])
	} else {
		// Managed flow: deploy certificate-less; Azure fills the
		// certificate binding in asynchronously once the managed
		// certificate issues. Without the ignore, every refresh after
		// that attachment would read as drift and plan a replacement.
		resourceOptions = append(resourceOptions,
			pulumi.IgnoreChanges([]string{"certificateBindingType", "containerAppEnvironmentCertificateId"}))
	}

	createdDomain, err := containerapp.NewCustomDomain(ctx,
		"main",
		domainArgs,
		resourceOptions...)
	if err != nil {
		return errors.Wrapf(err, "failed to create container app custom domain %s", spec.DomainName)
	}

	ctx.Export(OpCustomDomainId, createdDomain.ID())
	// Empty for bring-your-own bindings, and empty until Azure attaches
	// the managed certificate (asynchronous).
	ctx.Export(OpManagedCertificateId, createdDomain.ContainerAppEnvironmentManagedCertificateId)

	return nil
}
