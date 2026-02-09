package module

import (
	"github.com/pkg/errors"
	kubernetesistiov1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesistio/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespaces conditionally creates the istio-system and istio-ingress namespaces
// based on the create_namespace flag. Returns the namespace names as StringOutputs
// (from the created resources or from string constants) plus a dependency list that
// downstream Helm releases should include in their resource options.
func namespaces(ctx *pulumi.Context,
	stackInput *kubernetesistiov1.KubernetesIstioStackInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (sysNSName pulumi.StringOutput, gwNSName pulumi.StringOutput, namespaceDeps []pulumi.ResourceOption, err error) {
	if !stackInput.Target.Spec.CreateNamespace {
		sysNSName = pulumi.String(locals.SystemNamespace).ToStringOutput()
		gwNSName = pulumi.String(locals.GatewayNamespace).ToStringOutput()
		return
	}

	sysNS, nsErr := kubernetescorev1.NewNamespace(ctx, locals.SystemNamespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name: pulumi.String(locals.SystemNamespace),
			},
		},
		pulumi.Provider(kubernetesProvider))
	if nsErr != nil {
		err = errors.Wrap(nsErr, "failed to create istio-system namespace")
		return
	}
	sysNSName = sysNS.Metadata.Name().Elem()

	gwNS, nsErr := kubernetescorev1.NewNamespace(ctx, locals.GatewayNamespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name: pulumi.String(locals.GatewayNamespace),
			},
		},
		pulumi.Provider(kubernetesProvider))
	if nsErr != nil {
		err = errors.Wrap(nsErr, "failed to create istio-ingress namespace")
		return
	}
	gwNSName = gwNS.Metadata.Name().Elem()

	namespaceDeps = []pulumi.ResourceOption{
		pulumi.DependsOn([]pulumi.Resource{sysNS, gwNS}),
	}
	return
}
