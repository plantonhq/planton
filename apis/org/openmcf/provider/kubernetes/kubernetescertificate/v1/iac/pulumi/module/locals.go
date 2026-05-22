package module

import (
	kubernetescertificatev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetescertificate/v1"
	certmanagerv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesCertificate *kubernetescertificatev1.KubernetesCertificate
	Namespace             string
	CertificateName       string
	Labels                map[string]string
	DnsNames              []string
	SecretName            string
	IsCa                  bool
	IssuerRefKind         string
	IssuerRefName         string
	Duration              pulumi.StringPtrInput
	RenewBefore           pulumi.StringPtrInput
	// nil means "use cert-manager defaults" -- the CRD applies sensible defaults
	// (RSA 2048, PKCS1, Always) when no privateKey block is present at all.
	PrivateKey *certmanagerv1.CertificateSpecPrivateKeyArgs
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetescertificatev1.KubernetesCertificateStackInput) *Locals {
	locals := &Locals{}
	locals.KubernetesCertificate = stackInput.Target

	target := stackInput.Target
	spec := target.Spec

	locals.Namespace = spec.Namespace.GetValue()
	locals.CertificateName = target.Metadata.Name
	locals.Labels = map[string]string{
		"app.kubernetes.io/managed-by": "openmcf",
		"resource.openmcf.org/id":      target.Metadata.Name,
	}
	locals.DnsNames = spec.DnsNames
	locals.SecretName = spec.SecretName
	locals.IsCa = spec.IsCa

	// The proto oneof generates wrapper types: CertificateIssuerRef_ClusterIssuer
	// and CertificateIssuerRef_Issuer. The type switch determines the cert-manager
	// issuerRef.kind ("ClusterIssuer" vs "Issuer") and extracts the name from the
	// StringValueOrRef FK field.
	switch ref := spec.IssuerRef.IssuerType.(type) {
	case *kubernetescertificatev1.CertificateIssuerRef_ClusterIssuer:
		locals.IssuerRefKind = "ClusterIssuer"
		locals.IssuerRefName = ref.ClusterIssuer.Name.GetValue()
	case *kubernetescertificatev1.CertificateIssuerRef_Issuer:
		locals.IssuerRefKind = "Issuer"
		locals.IssuerRefName = ref.Issuer.Name.GetValue()
	}

	// Duration and RenewBefore are optional. When the duration_config submessage is
	// nil the Pulumi args remain nil, causing cert-manager to use its own defaults
	// (90d duration, 30d renew-before). Middleware populates defaults from the proto
	// options annotation, but we still guard against nil for direct IaC usage.
	if dc := spec.GetDurationConfig(); dc != nil {
		if d := dc.GetDuration(); d != "" {
			locals.Duration = pulumi.StringPtr(d)
		}
		if rb := dc.GetRenewBefore(); rb != "" {
			locals.RenewBefore = pulumi.StringPtr(rb)
		}
	}

	if pk := spec.GetPrivateKey(); pk != nil {
		locals.PrivateKey = buildPrivateKeyArgs(pk)
	}

	return locals
}

// buildPrivateKeyArgs maps the proto CertificatePrivateKey to the typed Pulumi
// CertificateSpecPrivateKeyArgs. Each field is conditionally set only when the
// proto optional field is present.
func buildPrivateKeyArgs(pk *kubernetescertificatev1.CertificatePrivateKey) *certmanagerv1.CertificateSpecPrivateKeyArgs {
	args := &certmanagerv1.CertificateSpecPrivateKeyArgs{}

	if pk.Algorithm != nil {
		args.Algorithm = pulumi.StringPtr(mapAlgorithm(pk.GetAlgorithm()))
	}
	if pk.Size != nil {
		args.Size = pulumi.IntPtr(int(pk.GetSize()))
	}
	if pk.Encoding != nil {
		args.Encoding = pulumi.StringPtr(mapEncoding(pk.GetEncoding()))
	}
	if pk.RotationPolicy != nil {
		args.RotationPolicy = pulumi.StringPtr(mapRotationPolicy(pk.GetRotationPolicy()))
	}

	return args
}

// Proto enum values use lowercase names (rsa, ecdsa, ed25519) defined in
// spec.proto. The cert-manager CRD expects PascalCase: RSA, ECDSA, Ed25519.
// These mapping functions bridge the two naming conventions.

func mapAlgorithm(a kubernetescertificatev1.CertificatePrivateKey_PrivateKeyAlgorithm) string {
	switch a {
	case kubernetescertificatev1.CertificatePrivateKey_rsa:
		return "RSA"
	case kubernetescertificatev1.CertificatePrivateKey_ecdsa:
		return "ECDSA"
	case kubernetescertificatev1.CertificatePrivateKey_ed25519:
		return "Ed25519"
	default:
		return "RSA"
	}
}

// Proto encoding enums (pkcs1, pkcs8) map to cert-manager's uppercase PKCS1, PKCS8.
func mapEncoding(e kubernetescertificatev1.CertificatePrivateKey_PrivateKeyEncoding) string {
	switch e {
	case kubernetescertificatev1.CertificatePrivateKey_pkcs1:
		return "PKCS1"
	case kubernetescertificatev1.CertificatePrivateKey_pkcs8:
		return "PKCS8"
	default:
		return "PKCS1"
	}
}

// Proto rotation enums (always, never) map to cert-manager's PascalCase Always, Never.
func mapRotationPolicy(r kubernetescertificatev1.CertificatePrivateKey_PrivateKeyRotationPolicy) string {
	switch r {
	case kubernetescertificatev1.CertificatePrivateKey_always:
		return "Always"
	case kubernetescertificatev1.CertificatePrivateKey_never:
		return "Never"
	default:
		return "Always"
	}
}
