package module

import (
	"github.com/pkg/errors"
	kubernetesenvoyfilterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesenvoyfilter/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istionetworkingv1alpha3 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/networking/v1alpha3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/types/known/structpb"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesenvoyfilterv1.KubernetesEnvoyFilterStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createEnvoyFilter(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create envoy filter")
	}

	ctx.Export(OpEnvoyFilterName, pulumi.String(locals.EnvoyFilterName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createEnvoyFilter creates the namespaced Istio EnvoyFilter using the typed crd2pulumi SDK
// (istionetworkingv1alpha3.NewEnvoyFilter), consistent with every other Planton Istio
// component. EnvoyFilter is the only Istio API component still served at networking/v1alpha3
// (it has not graduated to v1). Every block is attached only when present so unset fields
// fall through to istiod's defaults.
func createEnvoyFilter(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesEnvoyFilter.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value itself
	// (not the SpecPtr() wrapper, which marshals to the wrong element type and panics at
	// `pulumi up`); assigned directly below, mirroring the sibling Istio components.
	efSpec := istionetworkingv1alpha3.EnvoyFilterSpecArgs{}

	if patches := spec.GetConfigPatches(); len(patches) > 0 {
		patchArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesArray{}
		for _, patch := range patches {
			patchArgs = append(patchArgs, buildConfigPatchArgs(patch))
		}
		efSpec.ConfigPatches = patchArgs
	}

	if spec.Priority != nil {
		efSpec.Priority = pulumi.Int(int(spec.GetPriority()))
	}

	if selector := spec.GetWorkloadSelector(); selector != nil && len(selector.GetLabels()) > 0 {
		efSpec.WorkloadSelector = istionetworkingv1alpha3.EnvoyFilterSpecWorkloadSelectorArgs{
			Labels: pulumi.ToStringMap(selector.GetLabels()),
		}
	}

	if targetRefs := spec.GetTargetRefs(); len(targetRefs) > 0 {
		refArgs := istionetworkingv1alpha3.EnvoyFilterSpecTargetRefsArray{}
		for _, ref := range targetRefs {
			args := istionetworkingv1alpha3.EnvoyFilterSpecTargetRefsArgs{
				Kind: pulumi.String(ref.GetKind()),
				Name: pulumi.String(ref.GetName()),
			}
			if ref.GetGroup() != "" {
				args.Group = pulumi.String(ref.GetGroup())
			}
			if ref.GetNamespace() != "" {
				args.Namespace = pulumi.String(ref.GetNamespace())
			}
			refArgs = append(refArgs, args)
		}
		efSpec.TargetRefs = refArgs
	}

	_, err := istionetworkingv1alpha3.NewEnvoyFilter(ctx, locals.EnvoyFilterName,
		&istionetworkingv1alpha3.EnvoyFilterArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.EnvoyFilterName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: efSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildConfigPatchArgs maps one Planton config patch to the typed SDK args. apply_to, match,
// and patch are each attached only when present.
func buildConfigPatchArgs(patch *kubernetesenvoyfilterv1.KubernetesEnvoyFilterConfigPatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesArgs{}
	if patch.ApplyTo != nil {
		args.ApplyTo = pulumi.String(patch.GetApplyTo())
	}
	if match := patch.GetMatch(); match != nil {
		args.Match = buildMatchArgs(match)
	}
	if p := patch.GetPatch(); p != nil {
		args.Patch = buildPatchArgs(p)
	}
	return args
}

// buildMatchArgs maps the match conditions. context is attached only when present; the proxy
// match and the (at-most-one) listener/route_configuration/cluster branches are built only
// when set.
func buildMatchArgs(match *kubernetesenvoyfilterv1.KubernetesEnvoyFilterEnvoyConfigObjectMatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchArgs{}
	if match.Context != nil {
		args.Context = pulumi.String(match.GetContext())
	}
	if proxy := match.GetProxy(); proxy != nil {
		proxyArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchProxyArgs{}
		if proxy.ProxyVersion != nil {
			proxyArgs.ProxyVersion = pulumi.String(proxy.GetProxyVersion())
		}
		if md := proxy.GetMetadata(); len(md) > 0 {
			proxyArgs.Metadata = pulumi.ToStringMap(md)
		}
		args.Proxy = proxyArgs
	}
	if listener := match.GetListener(); listener != nil {
		args.Listener = buildListenerMatchArgs(listener)
	}
	if rc := match.GetRouteConfiguration(); rc != nil {
		args.RouteConfiguration = buildRouteConfigurationMatchArgs(rc)
	}
	if cluster := match.GetCluster(); cluster != nil {
		args.Cluster = buildClusterMatchArgs(cluster)
	}
	return args
}

func buildClusterMatchArgs(cluster *kubernetesenvoyfilterv1.KubernetesEnvoyFilterClusterMatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchClusterArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchClusterArgs{}
	if cluster.PortNumber != nil {
		args.PortNumber = pulumi.Int(int(cluster.GetPortNumber()))
	}
	if cluster.Service != nil {
		args.Service = pulumi.String(cluster.GetService())
	}
	if cluster.Subset != nil {
		args.Subset = pulumi.String(cluster.GetSubset())
	}
	if cluster.Name != nil {
		args.Name = pulumi.String(cluster.GetName())
	}
	return args
}

func buildRouteConfigurationMatchArgs(rc *kubernetesenvoyfilterv1.KubernetesEnvoyFilterRouteConfigurationMatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchRouteConfigurationArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchRouteConfigurationArgs{}
	if rc.PortNumber != nil {
		args.PortNumber = pulumi.Int(int(rc.GetPortNumber()))
	}
	if rc.PortName != nil {
		args.PortName = pulumi.String(rc.GetPortName())
	}
	if rc.Gateway != nil {
		args.Gateway = pulumi.String(rc.GetGateway())
	}
	if vhost := rc.GetVhost(); vhost != nil {
		vhostArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchRouteConfigurationVhostArgs{}
		if vhost.Name != nil {
			vhostArgs.Name = pulumi.String(vhost.GetName())
		}
		if vhost.DomainName != nil {
			vhostArgs.DomainName = pulumi.String(vhost.GetDomainName())
		}
		if route := vhost.GetRoute(); route != nil {
			routeArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchRouteConfigurationVhostRouteArgs{}
			if route.Name != nil {
				routeArgs.Name = pulumi.String(route.GetName())
			}
			if route.Action != nil {
				routeArgs.Action = pulumi.String(route.GetAction())
			}
			vhostArgs.Route = routeArgs
		}
		args.Vhost = vhostArgs
	}
	if rc.Name != nil {
		args.Name = pulumi.String(rc.GetName())
	}
	return args
}

func buildListenerMatchArgs(listener *kubernetesenvoyfilterv1.KubernetesEnvoyFilterListenerMatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerArgs{}
	if listener.PortNumber != nil {
		args.PortNumber = pulumi.Int(int(listener.GetPortNumber()))
	}
	if fc := listener.GetFilterChain(); fc != nil {
		args.FilterChain = buildFilterChainMatchArgs(fc)
	}
	if listener.ListenerFilter != nil {
		args.ListenerFilter = pulumi.String(listener.GetListenerFilter())
	}
	if listener.Name != nil {
		args.Name = pulumi.String(listener.GetName())
	}
	return args
}

func buildFilterChainMatchArgs(fc *kubernetesenvoyfilterv1.KubernetesEnvoyFilterFilterChainMatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerFilterChainArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerFilterChainArgs{}
	if fc.Name != nil {
		args.Name = pulumi.String(fc.GetName())
	}
	if fc.Sni != nil {
		args.Sni = pulumi.String(fc.GetSni())
	}
	if fc.TransportProtocol != nil {
		args.TransportProtocol = pulumi.String(fc.GetTransportProtocol())
	}
	if fc.ApplicationProtocols != nil {
		args.ApplicationProtocols = pulumi.String(fc.GetApplicationProtocols())
	}
	if filter := fc.GetFilter(); filter != nil {
		filterArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerFilterChainFilterArgs{}
		if filter.Name != nil {
			filterArgs.Name = pulumi.String(filter.GetName())
		}
		if sub := filter.GetSubFilter(); sub != nil {
			subArgs := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesMatchListenerFilterChainFilterSubFilterArgs{}
			if sub.Name != nil {
				subArgs.Name = pulumi.String(sub.GetName())
			}
			filterArgs.SubFilter = subArgs
		}
		args.Filter = filterArgs
	}
	if fc.DestinationPort != nil {
		args.DestinationPort = pulumi.Int(int(fc.GetDestinationPort()))
	}
	return args
}

// buildPatchArgs maps the patch. operation and filter_class are attached only when present;
// the free-form Struct `value` is converted to a Pulumi map preserving nested structure.
func buildPatchArgs(p *kubernetesenvoyfilterv1.KubernetesEnvoyFilterPatch) istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesPatchArgs {
	args := istionetworkingv1alpha3.EnvoyFilterSpecConfigPatchesPatchArgs{}
	if p.Operation != nil {
		args.Operation = pulumi.String(p.GetOperation())
	}
	if value := p.GetValue(); value != nil && len(value.GetFields()) > 0 {
		args.Value = structToPulumiMap(value)
	}
	if p.FilterClass != nil {
		args.FilterClass = pulumi.String(p.GetFilterClass())
	}
	return args
}

// structToPulumiMap converts a google.protobuf.Struct (the free-form xDS patch value) into a
// pulumi.Map, recursively preserving nested objects, arrays, and scalars. The crd2pulumi SDK
// types `patch.value` as pulumi.MapInput (the upstream CRD marks it preserveUnknownFields), so
// the arbitrary JSON is passed through unmodified to the manifest.
func structToPulumiMap(s *structpb.Struct) pulumi.Map {
	if s == nil {
		return nil
	}
	return goMapToPulumiMap(s.AsMap())
}

func goMapToPulumiMap(m map[string]interface{}) pulumi.Map {
	out := pulumi.Map{}
	for k, v := range m {
		out[k] = goValueToPulumi(v)
	}
	return out
}

func goValueToPulumi(v interface{}) pulumi.Input {
	switch val := v.(type) {
	case nil:
		return pulumi.Any(nil)
	case bool:
		return pulumi.Bool(val)
	case float64:
		return pulumi.Float64(val)
	case string:
		return pulumi.String(val)
	case map[string]interface{}:
		return goMapToPulumiMap(val)
	case []interface{}:
		arr := pulumi.Array{}
		for _, e := range val {
			arr = append(arr, goValueToPulumi(e))
		}
		return arr
	default:
		return pulumi.Any(val)
	}
}
