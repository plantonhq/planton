package module

import (
	"strings"

	"github.com/pkg/errors"
	ocicontainerengineclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocicontainerenginecluster/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/containerengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cluster(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciContainerEngineCluster.Spec

	args := &containerengine.ClusterArgs{
		CompartmentId:     pulumi.String(spec.CompartmentId.GetValue()),
		VcnId:             pulumi.String(spec.VcnId.GetValue()),
		KubernetesVersion: pulumi.String(spec.KubernetesVersion),
		Name:              pulumi.StringPtr(locals.DisplayName),
		FreeformTags:      pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.Type != ocicontainerengineclusterv1.OciContainerEngineClusterSpec_unspecified {
		args.Type = pulumi.StringPtr(strings.ToUpper(spec.Type.String()))
	}

	if spec.CniType != ocicontainerengineclusterv1.OciContainerEngineClusterSpec_cni_unspecified {
		args.ClusterPodNetworkOptions = containerengine.ClusterClusterPodNetworkOptionArray{
			&containerengine.ClusterClusterPodNetworkOptionArgs{
				CniType: pulumi.String(strings.ToUpper(spec.CniType.String())),
			},
		}
	}

	if spec.EndpointConfig != nil {
		args.EndpointConfig = buildEndpointConfig(spec.EndpointConfig)
	}

	if spec.Options != nil {
		args.Options = buildClusterOptions(spec.Options)
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.ImagePolicyConfig != nil {
		args.ImagePolicyConfig = buildImagePolicyConfig(spec.ImagePolicyConfig)
	}

	createdCluster, err := containerengine.NewCluster(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci container engine cluster")
	}

	ctx.Export(OpClusterId, createdCluster.ID())
	ctx.Export(OpKubernetesVersion, createdCluster.KubernetesVersion)

	ctx.Export(OpKubernetesEndpoint, createdCluster.Endpoints.ApplyT(func(endpoints []containerengine.ClusterEndpoint) string {
		if len(endpoints) > 0 && endpoints[0].Kubernetes != nil {
			return *endpoints[0].Kubernetes
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpPrivateEndpoint, createdCluster.Endpoints.ApplyT(func(endpoints []containerengine.ClusterEndpoint) string {
		if len(endpoints) > 0 && endpoints[0].PrivateEndpoint != nil {
			return *endpoints[0].PrivateEndpoint
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpPublicEndpoint, createdCluster.Endpoints.ApplyT(func(endpoints []containerengine.ClusterEndpoint) string {
		if len(endpoints) > 0 && endpoints[0].PublicEndpoint != nil {
			return *endpoints[0].PublicEndpoint
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

func buildEndpointConfig(ec *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_EndpointConfig) *containerengine.ClusterEndpointConfigArgs {
	args := &containerengine.ClusterEndpointConfigArgs{
		SubnetId: pulumi.String(ec.SubnetId.GetValue()),
	}

	if ec.IsPublicIpEnabled != nil {
		args.IsPublicIpEnabled = pulumi.BoolPtr(*ec.IsPublicIpEnabled)
	}

	if len(ec.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(ec.NsgIds))
		for i, nsg := range ec.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	return args
}

func buildClusterOptions(opts *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_ClusterOptions) *containerengine.ClusterOptionsArgs {
	args := &containerengine.ClusterOptionsArgs{}

	if opts.KubernetesNetworkConfig != nil {
		args.KubernetesNetworkConfig = buildKubernetesNetworkConfig(opts.KubernetesNetworkConfig)
	}

	if len(opts.ServiceLbSubnetIds) > 0 {
		subnetIds := make(pulumi.StringArray, len(opts.ServiceLbSubnetIds))
		for i, s := range opts.ServiceLbSubnetIds {
			subnetIds[i] = pulumi.String(s.GetValue())
		}
		args.ServiceLbSubnetIds = subnetIds
	}

	if len(opts.IpFamilies) > 0 {
		families := make(pulumi.StringArray, len(opts.IpFamilies))
		for i, f := range opts.IpFamilies {
			families[i] = pulumi.String(ipFamilyString(f))
		}
		args.IpFamilies = families
	}

	if opts.ServiceLbConfig != nil {
		args.ServiceLbConfig = buildServiceLbConfig(opts.ServiceLbConfig)
	}

	if opts.PersistentVolumeConfig != nil {
		args.PersistentVolumeConfig = buildPersistentVolumeConfig(opts.PersistentVolumeConfig)
	}

	if opts.OpenIdConnectTokenAuthenticationConfig != nil {
		args.OpenIdConnectTokenAuthenticationConfig = buildOidcConfig(opts.OpenIdConnectTokenAuthenticationConfig)
	}

	if opts.IsOpenIdConnectDiscoveryEnabled {
		args.OpenIdConnectDiscovery = &containerengine.ClusterOptionsOpenIdConnectDiscoveryArgs{
			IsOpenIdConnectDiscoveryEnabled: pulumi.BoolPtr(true),
		}
	}

	return args
}

func buildKubernetesNetworkConfig(knc *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_KubernetesNetworkConfig) *containerengine.ClusterOptionsKubernetesNetworkConfigArgs {
	args := &containerengine.ClusterOptionsKubernetesNetworkConfigArgs{}

	if knc.PodsCidr != "" {
		args.PodsCidr = pulumi.StringPtr(knc.PodsCidr)
	}

	if knc.ServicesCidr != "" {
		args.ServicesCidr = pulumi.StringPtr(knc.ServicesCidr)
	}

	return args
}

func buildServiceLbConfig(slc *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_ServiceLbConfig) *containerengine.ClusterOptionsServiceLbConfigArgs {
	args := &containerengine.ClusterOptionsServiceLbConfigArgs{}

	if len(slc.BackendNsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(slc.BackendNsgIds))
		for i, nsg := range slc.BackendNsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.BackendNsgIds = nsgIds
	}

	if len(slc.FreeformTags) > 0 {
		args.FreeformTags = pulumi.ToStringMap(slc.FreeformTags)
	}

	if len(slc.DefinedTags) > 0 {
		args.DefinedTags = pulumi.ToStringMap(slc.DefinedTags)
	}

	return args
}

func buildPersistentVolumeConfig(pvc *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_PersistentVolumeConfig) *containerengine.ClusterOptionsPersistentVolumeConfigArgs {
	args := &containerengine.ClusterOptionsPersistentVolumeConfigArgs{}

	if len(pvc.FreeformTags) > 0 {
		args.FreeformTags = pulumi.ToStringMap(pvc.FreeformTags)
	}

	if len(pvc.DefinedTags) > 0 {
		args.DefinedTags = pulumi.ToStringMap(pvc.DefinedTags)
	}

	return args
}

func buildOidcConfig(oidc *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_OpenIdConnectTokenAuthenticationConfig) *containerengine.ClusterOptionsOpenIdConnectTokenAuthenticationConfigArgs {
	args := &containerengine.ClusterOptionsOpenIdConnectTokenAuthenticationConfigArgs{
		IsOpenIdConnectAuthEnabled: pulumi.Bool(oidc.IsOpenIdConnectAuthEnabled),
	}

	if oidc.ConfigurationFile != "" {
		args.ConfigurationFile = pulumi.StringPtr(oidc.ConfigurationFile)
	}

	if oidc.IssuerUrl != "" {
		args.IssuerUrl = pulumi.StringPtr(oidc.IssuerUrl)
	}

	if oidc.ClientId != "" {
		args.ClientId = pulumi.StringPtr(oidc.ClientId)
	}

	if oidc.CaCertificate != "" {
		args.CaCertificate = pulumi.StringPtr(oidc.CaCertificate)
	}

	if oidc.UsernameClaim != "" {
		args.UsernameClaim = pulumi.StringPtr(oidc.UsernameClaim)
	}

	if oidc.UsernamePrefix != "" {
		args.UsernamePrefix = pulumi.StringPtr(oidc.UsernamePrefix)
	}

	if oidc.GroupsClaim != "" {
		args.GroupsClaim = pulumi.StringPtr(oidc.GroupsClaim)
	}

	if oidc.GroupsPrefix != "" {
		args.GroupsPrefix = pulumi.StringPtr(oidc.GroupsPrefix)
	}

	if len(oidc.SigningAlgorithms) > 0 {
		algos := make(pulumi.StringArray, len(oidc.SigningAlgorithms))
		for i, a := range oidc.SigningAlgorithms {
			algos[i] = pulumi.String(a)
		}
		args.SigningAlgorithms = algos
	}

	if len(oidc.RequiredClaims) > 0 {
		claims := make(containerengine.ClusterOptionsOpenIdConnectTokenAuthenticationConfigRequiredClaimArray, len(oidc.RequiredClaims))
		for i, rc := range oidc.RequiredClaims {
			claims[i] = &containerengine.ClusterOptionsOpenIdConnectTokenAuthenticationConfigRequiredClaimArgs{
				Key:   pulumi.StringPtr(rc.Key),
				Value: pulumi.StringPtr(rc.Value),
			}
		}
		args.RequiredClaims = claims
	}

	return args
}

func buildImagePolicyConfig(ipc *ocicontainerengineclusterv1.OciContainerEngineClusterSpec_ImagePolicyConfig) *containerengine.ClusterImagePolicyConfigArgs {
	args := &containerengine.ClusterImagePolicyConfigArgs{
		IsPolicyEnabled: pulumi.BoolPtr(ipc.IsPolicyEnabled),
	}

	if len(ipc.KeyDetails) > 0 {
		keys := make(containerengine.ClusterImagePolicyConfigKeyDetailArray, len(ipc.KeyDetails))
		for i, kd := range ipc.KeyDetails {
			keys[i] = &containerengine.ClusterImagePolicyConfigKeyDetailArgs{
				KmsKeyId: pulumi.StringPtr(kd.KmsKeyId.GetValue()),
			}
		}
		args.KeyDetails = keys
	}

	return args
}

func ipFamilyString(f ocicontainerengineclusterv1.OciContainerEngineClusterSpec_IpFamily) string {
	switch f {
	case ocicontainerengineclusterv1.OciContainerEngineClusterSpec_ipv4:
		return "IPv4"
	case ocicontainerengineclusterv1.OciContainerEngineClusterSpec_ipv6:
		return "IPv6"
	default:
		return "IPv4"
	}
}
