package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createService creates the Kubernetes Service resource based on the computed locals.
func createService(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*kubernetescorev1.Service, error) {
	// Build the port specifications.
	servicePorts := buildServicePorts(locals)

	// Build the core service spec arguments.
	serviceSpecArgs := &kubernetescorev1.ServiceSpecArgs{
		Type:  pulumi.String(locals.ServiceType),
		Ports: servicePorts,
	}

	// Set selector if provided (not for ExternalName services or selectorless services).
	if len(locals.Selector) > 0 {
		serviceSpecArgs.Selector = pulumi.ToStringMap(locals.Selector)
	}

	// Handle headless services: set clusterIP to "None".
	if locals.Headless {
		serviceSpecArgs.ClusterIP = pulumi.String("None")
	}

	// Set ExternalName for ExternalName-type services.
	if locals.ServiceType == "ExternalName" && locals.ExternalDnsName != "" {
		serviceSpecArgs.ExternalName = pulumi.String(locals.ExternalDnsName)
	}

	// Set external traffic policy for NodePort and LoadBalancer services.
	if locals.ServiceType == "NodePort" || locals.ServiceType == "LoadBalancer" {
		serviceSpecArgs.ExternalTrafficPolicy = pulumi.String(locals.ExternalTrafficPolicy)
	}

	// Set session affinity.
	if locals.SessionAffinity != "None" {
		serviceSpecArgs.SessionAffinity = pulumi.String(locals.SessionAffinity)
	}

	// Set load balancer source ranges for LoadBalancer services.
	if locals.ServiceType == "LoadBalancer" && len(locals.LoadBalancerSourceRanges) > 0 {
		serviceSpecArgs.LoadBalancerSourceRanges = pulumi.ToStringArray(locals.LoadBalancerSourceRanges)
	}

	// Create the Service resource.
	service, err := kubernetescorev1.NewService(
		ctx,
		locals.Name,
		&kubernetescorev1.ServiceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.Name),
				Namespace:   pulumi.String(locals.Namespace),
				Labels:      pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.ToStringMap(locals.Annotations),
			},
			Spec: serviceSpecArgs,
		},
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create service %s/%s", locals.Namespace, locals.Name)
	}

	return service, nil
}

// buildServicePorts converts the protobuf port definitions to Pulumi service port arguments.
func buildServicePorts(locals *Locals) kubernetescorev1.ServicePortArray {
	var ports kubernetescorev1.ServicePortArray

	for _, p := range locals.Ports {
		portArgs := &kubernetescorev1.ServicePortArgs{
			Port:     pulumi.Int(p.GetPort()),
			Protocol: pulumi.String(resolveProtocol(p.GetProtocol())),
		}

		// Set port name if provided.
		if p.GetName() != "" {
			portArgs.Name = pulumi.String(p.GetName())
		}

		// Set target port. Can be a number or a named port.
		if p.GetTargetPort() != "" {
			// Try to parse as integer first; if it's a number, use IntOrString with int.
			if num, err := strconv.Atoi(p.GetTargetPort()); err == nil {
				portArgs.TargetPort = pulumi.Int(num)
			} else {
				portArgs.TargetPort = pulumi.String(p.GetTargetPort())
			}
		}

		// Set node port if specified (non-zero).
		if p.GetNodePort() > 0 {
			portArgs.NodePort = pulumi.Int(int(p.GetNodePort()))
		}

		ports = append(ports, portArgs)
	}

	return ports
}

// formatPort returns a human-readable port string for logging purposes.
func formatPort(port int32, name string) string {
	if name != "" {
		return fmt.Sprintf("%s/%d", name, port)
	}
	return fmt.Sprintf("%d", port)
}
