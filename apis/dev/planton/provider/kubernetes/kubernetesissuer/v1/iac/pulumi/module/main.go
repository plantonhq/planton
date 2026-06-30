package module

import (
	"github.com/pkg/errors"
	kubernetesissuerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesissuer/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	certmanagerv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/certmanager/kubernetes/cert_manager/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesissuerv1.KubernetesIssuerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createIssuer(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create issuer")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpIssuerName, pulumi.String(locals.IssuerName))

	return nil
}

// createIssuer creates a namespace-scoped cert-manager Issuer using the typed
// crd2pulumi SDK (certmanagerv1.NewIssuer). The issuer_type oneof in the proto
// spec maps to mutually exclusive branches here: CA or SelfSigned.
//
// Unlike KubernetesClusterIssuer (which is cluster-scoped and always uses ACME),
// an Issuer is namespace-scoped and only supports CA or SelfSigned modes.
// No namespace creation is performed -- the Issuer targets an EXISTING namespace
// (the referenced namespace must already exist on the cluster).
func createIssuer(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	issuerSpec := buildIssuerSpec(locals)

	// Metadata uses value ObjectMetaArgs (not pointer) -- this satisfies
	// metav1.ObjectMetaPtrInput and matches the pattern established by all
	// 17 ingress components that use certmanagerv1.NewCertificate.
	_, err := certmanagerv1.NewIssuer(ctx, locals.IssuerName, &certmanagerv1.IssuerArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.IssuerName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: issuerSpec,
	}, pulumi.Provider(kubeProvider))

	return err
}

// buildIssuerSpec maps the proto issuer_type oneof to the typed cert-manager
// IssuerSpec. The proto uses CaIssuerConfig.ca_secret_name but the CRD field
// is spec.ca.secretName -- the typed Pulumi SDK exposes this as
// IssuerSpecCaArgs.SecretName (NOT CaSecretName).
func buildIssuerSpec(locals *Locals) certmanagerv1.IssuerSpecArgs {
	if locals.IsCa {
		return certmanagerv1.IssuerSpecArgs{
			Ca: certmanagerv1.IssuerSpecCaArgs{
				SecretName: pulumi.String(locals.CaSecretName),
			},
		}
	}

	return certmanagerv1.IssuerSpecArgs{
		SelfSigned: certmanagerv1.IssuerSpecSelfSignedArgs{},
	}
}
