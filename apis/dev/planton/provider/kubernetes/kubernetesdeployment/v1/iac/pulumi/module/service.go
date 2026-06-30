package module

import (
	"github.com/pkg/errors"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func service(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, createdDeployment *appsv1.Deployment, namespaceDeps []pulumi.ResourceOption) error {

	//if the service ports are empty, we don't need to create a service
	if len(locals.KubernetesDeployment.Spec.Container.App.Ports) == 0 {
		return nil
	}

	portsArray := make(kubernetescorev1.ServicePortArray, 0)
	for _, p := range locals.KubernetesDeployment.Spec.Container.App.Ports {
		portsArray = append(portsArray, &kubernetescorev1.ServicePortArgs{
			Name:        pulumi.String(p.Name),
			Protocol:    pulumi.String(p.NetworkProtocol),
			Port:        pulumi.Int(p.ServicePort),
			TargetPort:  pulumi.Int(p.ContainerPort),
			AppProtocol: pulumi.String(p.AppProtocol),
		})
	}

	serviceArgs := &kubernetescorev1.ServiceArgs{
		Metadata: kubernetesmetav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.KubeServiceName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: &kubernetescorev1.ServiceSpecArgs{
			Type:     pulumi.String("ClusterIP"),
			Selector: pulumi.ToStringMap(locals.SelectorLabels),
			Ports:    portsArray,
		},
	}

	svcOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{createdDeployment}),
	}, namespaceDeps...)
	_, err := kubernetescorev1.NewService(ctx,
		locals.KubeServiceName,
		serviceArgs,
		svcOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to add service")
	}
	return nil
}
