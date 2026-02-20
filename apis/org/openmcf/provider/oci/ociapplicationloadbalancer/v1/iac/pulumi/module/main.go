package module

import (
	"github.com/pkg/errors"
	ociapplicationloadbalancerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociapplicationloadbalancer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *ociapplicationloadbalancerv1.OciApplicationLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdLb, err := createLoadBalancer(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create load balancer")
	}

	createdCerts, err := createCertificates(ctx, locals, ociProvider, createdLb)
	if err != nil {
		return errors.Wrap(err, "failed to create certificates")
	}

	createdBackendSets, err := createBackendSets(ctx, locals, ociProvider, createdLb)
	if err != nil {
		return errors.Wrap(err, "failed to create backend sets")
	}

	createdHostnames, err := createHostnames(ctx, locals, ociProvider, createdLb)
	if err != nil {
		return errors.Wrap(err, "failed to create hostnames")
	}

	createdRuleSets, err := createRuleSets(ctx, locals, ociProvider, createdLb)
	if err != nil {
		return errors.Wrap(err, "failed to create rule sets")
	}

	if err := createListeners(ctx, locals, ociProvider, createdLb,
		createdBackendSets, createdHostnames, createdRuleSets, createdCerts); err != nil {
		return errors.Wrap(err, "failed to create listeners")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}

func pulumiDependsOn(deps ...pulumi.Resource) pulumi.ResourceOption {
	return pulumi.DependsOn(deps)
}

func collectDeps(
	backendSets []*loadbalancer.BackendSet,
	hostnames []*loadbalancer.Hostname,
	ruleSets []*loadbalancer.RuleSet,
	certificates []*loadbalancer.Certificate,
) []pulumi.Resource {
	var deps []pulumi.Resource
	for _, bs := range backendSets {
		deps = append(deps, bs)
	}
	for _, h := range hostnames {
		deps = append(deps, h)
	}
	for _, rs := range ruleSets {
		deps = append(deps, rs)
	}
	for _, c := range certificates {
		deps = append(deps, c)
	}
	return deps
}
