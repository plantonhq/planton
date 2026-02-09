package module

import (
	"github.com/pkg/errors"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// headlessService creates a headless service for the StatefulSet.
// Headless services are required for StatefulSets to provide stable network identity.
// Pod DNS: <pod-name>.<headless-service>.<namespace>.svc.cluster.local
func headlessService(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (*kubernetescorev1.Service, error) {

	portsArray := make(kubernetescorev1.ServicePortArray, 0)
	for _, p := range locals.KubernetesStatefulSet.Spec.Container.App.Ports {
		portsArray = append(portsArray, &kubernetescorev1.ServicePortArgs{
			Name:        pulumi.String(p.Name),
			Protocol:    pulumi.String(p.NetworkProtocol),
			Port:        pulumi.Int(p.ServicePort),
			TargetPort:  pulumi.Int(p.ContainerPort),
			AppProtocol: pulumi.String(p.AppProtocol),
		})
	}

	// Headless service has ClusterIP: None
	serviceArgs := &kubernetescorev1.ServiceArgs{
		Metadata: kubernetesmetav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.HeadlessServiceName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Spec: &kubernetescorev1.ServiceSpecArgs{
			Type:                     pulumi.String("ClusterIP"),
			ClusterIP:                pulumi.String("None"), // Makes it headless
			Selector:                 pulumi.ToStringMap(locals.SelectorLabels),
			Ports:                    portsArray,
			PublishNotReadyAddresses: pulumi.Bool(true), // Important for StatefulSets
		},
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdService, err := kubernetescorev1.NewService(ctx,
		locals.HeadlessServiceName,
		serviceArgs,
		opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add headless service")
	}

	return createdService, nil
}

// clientService creates a ClusterIP service for client access to the StatefulSet.
// This provides load-balanced access to the pods.
func clientService(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, createdStatefulSet *appsv1.StatefulSet, namespaceDeps []pulumi.ResourceOption) error {

	// If no ports are defined, skip creating the client service
	if len(locals.KubernetesStatefulSet.Spec.Container.App.Ports) == 0 {
		return nil
	}

	portsArray := make(kubernetescorev1.ServicePortArray, 0)
	for _, p := range locals.KubernetesStatefulSet.Spec.Container.App.Ports {
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

	clientOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{createdStatefulSet}),
	}, namespaceDeps...)
	_, err := kubernetescorev1.NewService(ctx,
		locals.KubeServiceName,
		serviceArgs,
		clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to add client service")
	}

	return nil
}
