package module

import (
	"github.com/pkg/errors"
	kubernetesserviceentryv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesserviceentry/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istionetworkingv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/istio/kubernetes/networking/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesserviceentryv1.KubernetesServiceEntryStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createServiceEntry(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create service entry")
	}

	ctx.Export(OpServiceEntryName, pulumi.String(locals.ServiceEntryName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createServiceEntry creates the namespaced Istio ServiceEntry using the typed crd2pulumi
// SDK (istionetworkingv1.NewServiceEntry), consistent with every other OpenMCF Istio
// component. The typed approach catches field-name and structure errors at compile time.
// Only `hosts` is always set (it is required upstream); every other block is attached only
// when present, so unset fields fall through to istiod's defaults.
func createServiceEntry(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesServiceEntry.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value itself
	// (not the SpecPtr() wrapper, which marshals to the wrong element type); assigned
	// directly below, mirroring the PeerAuthentication/RequestAuthentication components.
	seSpec := istionetworkingv1.ServiceEntrySpecArgs{
		Hosts: pulumi.ToStringArray(spec.GetHosts()),
	}

	if addresses := spec.GetAddresses(); len(addresses) > 0 {
		seSpec.Addresses = pulumi.ToStringArray(addresses)
	}
	if exportTo := spec.GetExportTo(); len(exportTo) > 0 {
		seSpec.ExportTo = pulumi.ToStringArray(exportTo)
	}
	if subjectAltNames := spec.GetSubjectAltNames(); len(subjectAltNames) > 0 {
		seSpec.SubjectAltNames = pulumi.ToStringArray(subjectAltNames)
	}
	if spec.Location != nil {
		seSpec.Location = pulumi.String(spec.GetLocation())
	}
	if spec.Resolution != nil {
		seSpec.Resolution = pulumi.String(spec.GetResolution())
	}

	if ports := spec.GetPorts(); len(ports) > 0 {
		portArgs := istionetworkingv1.ServiceEntrySpecPortsArray{}
		for _, port := range ports {
			portArgs = append(portArgs, buildPortArgs(port))
		}
		seSpec.Ports = portArgs
	}

	if endpoints := spec.GetEndpoints(); len(endpoints) > 0 {
		endpointArgs := istionetworkingv1.ServiceEntrySpecEndpointsArray{}
		for _, endpoint := range endpoints {
			endpointArgs = append(endpointArgs, buildEndpointArgs(endpoint))
		}
		seSpec.Endpoints = endpointArgs
	}

	if selector := spec.GetWorkloadSelector(); selector != nil && len(selector.GetLabels()) > 0 {
		seSpec.WorkloadSelector = istionetworkingv1.ServiceEntrySpecWorkloadSelectorArgs{
			Labels: pulumi.ToStringMap(selector.GetLabels()),
		}
	}

	_, err := istionetworkingv1.NewServiceEntry(ctx, locals.ServiceEntryName,
		&istionetworkingv1.ServiceEntryArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ServiceEntryName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: seSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildPortArgs maps one OpenMCF service port to the typed SDK args. The proto uint32 port
// numbers are converted to the SDK's int inputs; optional fields are attached only when
// present so unset fields are omitted from the CR.
func buildPortArgs(port *kubernetesserviceentryv1.KubernetesServiceEntryPort) istionetworkingv1.ServiceEntrySpecPortsArgs {
	args := istionetworkingv1.ServiceEntrySpecPortsArgs{
		Number: pulumi.Int(int(port.GetNumber())),
		Name:   pulumi.String(port.GetName()),
	}
	if port.Protocol != nil {
		args.Protocol = pulumi.String(port.GetProtocol())
	}
	if port.TargetPort != nil {
		args.TargetPort = pulumi.Int(int(port.GetTargetPort()))
	}
	return args
}

// buildEndpointArgs maps one OpenMCF endpoint to the typed SDK args. The endpoint port map
// (proto map[string]uint32) is converted to the SDK's pulumi.IntMap; optional scalar fields
// are attached only when present in the proto (the proto3 `optional` pointer distinguishes
// unset from empty), so unset fields are omitted and upstream defaults apply.
func buildEndpointArgs(endpoint *kubernetesserviceentryv1.KubernetesServiceEntryEndpoint) istionetworkingv1.ServiceEntrySpecEndpointsArgs {
	args := istionetworkingv1.ServiceEntrySpecEndpointsArgs{}
	if endpoint.Address != nil {
		args.Address = pulumi.String(endpoint.GetAddress())
	}
	if ports := endpoint.GetPorts(); len(ports) > 0 {
		portMap := pulumi.IntMap{}
		for name, number := range ports {
			portMap[name] = pulumi.Int(int(number))
		}
		args.Ports = portMap
	}
	if labels := endpoint.GetLabels(); len(labels) > 0 {
		args.Labels = pulumi.ToStringMap(labels)
	}
	if endpoint.Network != nil {
		args.Network = pulumi.String(endpoint.GetNetwork())
	}
	if endpoint.Locality != nil {
		args.Locality = pulumi.String(endpoint.GetLocality())
	}
	if endpoint.Weight != nil {
		args.Weight = pulumi.Int(int(endpoint.GetWeight()))
	}
	if endpoint.ServiceAccount != nil {
		args.ServiceAccount = pulumi.String(endpoint.GetServiceAccount())
	}
	return args
}
