package module

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificate provisions the Civo certificate.
// NOTE: As of 2025, the Civo Pulumi/Terraform provider does not expose a certificate resource.
// This function is a placeholder for future implementation when provider support is added.
func certificate(
	ctx *pulumi.Context,
	locals *Locals,
) error {
	// Log informational message about provider limitation
	ctx.Log.Warn(fmt.Sprintf(
		"Certificate '%s' specification is valid but cannot be provisioned. "+
			"The Civo Pulumi/Terraform provider does not currently support certificate resources. "+
			"Certificates must be managed manually via the Civo dashboard or API until provider support is added. "+
			"Refer to: https://registry.terraform.io/providers/civo/civo/latest/docs",
		locals.CivoCertificate.Spec.CertificateName,
	), nil)

	// Log certificate type and configuration
	if locals.CivoCertificate.Spec.Type == 1 { // letsEncrypt
		if locals.CivoCertificate.Spec.GetLetsEncrypt() != nil {
			domains := locals.CivoCertificate.Spec.GetLetsEncrypt().Domains
			ctx.Log.Info(fmt.Sprintf(
				"Let's Encrypt certificate requested for domains: %v (auto-renew: %v)",
				domains,
				!locals.CivoCertificate.Spec.GetLetsEncrypt().DisableAutoRenew,
			), nil)
		}
	} else if locals.CivoCertificate.Spec.Type == 2 { // custom
		ctx.Log.Info(fmt.Sprintf(
			"Custom certificate provided for '%s'",
			locals.CivoCertificate.Spec.CertificateName,
		), nil)
	}

	// Export placeholder outputs
	// These would be populated by actual certificate resources when provider support is added
	ctx.Export(OpCertificateId, pulumi.String(""))
	ctx.Export(OpExpiryRfc3339, pulumi.String(""))

	return nil
}
