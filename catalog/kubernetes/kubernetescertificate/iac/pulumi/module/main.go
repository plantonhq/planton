package module

import (
	"github.com/pkg/errors"
	kubernetescertificatev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one cert-manager Certificate using the typed crd2pulumi
// SDK (regenerated from the pinned cert-manager CRDs). The full spec mapping
// lives in spec_builder.go; the Terraform module renders the identical CR
// through its locals — keep the two in lockstep.
//
// Neither engine waits for the certificate to become Ready: issuance time
// belongs to the issuer (an ACME order can take minutes; an unreachable CA
// would block forever) — the same never-block-on-a-controller posture as
// Ingress. Consumers that need the TLS Secret express the dependency through
// composition; the E2E lanes verify issuance by polling the live cluster.
// Terraform equivalent: kubectl_manifest without a wait_for block.
func Resources(ctx *pulumi.Context, iacInput *kubernetescertificatev1alpha1.KubernetesCertificateIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	certificateSpec, err := buildCertificateSpec(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build certificate spec")
	}

	_, err = certmanagerv1.NewCertificate(ctx, locals.CertificateName,
		&certmanagerv1.CertificateArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.CertificateName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: certificateSpec,
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpCertificateName, pulumi.String(locals.CertificateName))
	ctx.Export(OpSecretName, pulumi.String(locals.SecretName))

	return nil
}
