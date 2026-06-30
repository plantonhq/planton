package module

import (
	"github.com/pkg/errors"
	kubernetescertificatev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetescertificate/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetescertificatev1.KubernetesCertificateStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// Create the cert-manager Certificate using the typed crd2pulumi SDK,
	// following the same pattern as KubernetesLocust ingress.go and all other
	// Planton components that create Certificate resources.
	certArgs := &certmanagerv1.CertificateArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.CertificateName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: certmanagerv1.CertificateSpecArgs{
			DnsNames:   pulumi.ToStringArray(locals.DnsNames),
			SecretName: pulumi.String(locals.SecretName),
			IssuerRef: certmanagerv1.CertificateSpecIssuerRefArgs{
				Kind: pulumi.String(locals.IssuerRefKind),
				Name: pulumi.String(locals.IssuerRefName),
			},
			IsCA:        pulumi.Bool(locals.IsCa),
			Duration:    locals.Duration,
			RenewBefore: locals.RenewBefore,
			PrivateKey:  locals.PrivateKey,
		},
	}

	_, err = certmanagerv1.NewCertificate(ctx, locals.CertificateName, certArgs,
		pulumi.Provider(kubeProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpCertificateName, pulumi.String(locals.CertificateName))
	ctx.Export(OpSecretName, pulumi.String(locals.SecretName))

	return nil
}
