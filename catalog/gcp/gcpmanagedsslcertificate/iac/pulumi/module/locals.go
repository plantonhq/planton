package module

import (
	gcpmanagedsslcertificatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmanagedsslcertificate/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpManagedSslCertificate *gcpmanagedsslcertificatev1alpha1.GcpManagedSslCertificate

	// The cloud-side name defaults to metadata.name when the spec leaves
	// certificate_name empty — the same naming basis every kind uses.
	CertificateName string
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpmanagedsslcertificatev1alpha1.GcpManagedSslCertificateIacInput) *Locals {
	target := iacInput.Target

	certificateName := target.Spec.CertificateName
	if certificateName == "" {
		certificateName = target.Metadata.Name
	}

	return &Locals{
		GcpManagedSslCertificate: target,
		CertificateName:          certificateName,
	}
}
