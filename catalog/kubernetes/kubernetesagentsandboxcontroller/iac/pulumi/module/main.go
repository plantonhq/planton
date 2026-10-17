package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesagentsandboxcontrollerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxcontroller/v1alpha1"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/keptcrds"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadowner"
	"github.com/plantonhq/planton/pkg/kubernetes/helmcrds"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs agent-sandbox from upstream's own release: the release
// asset the extensions setting selects (sandbox.yaml, or
// sandbox-with-extensions.yaml), and the router's deployment files at the
// same tag when a router is configured, each with the typed settings applied
// onto upstream's own objects (compose.go).
//
// ORDER: everything that can refuse runs before anything registers. A
// namespace other than upstream's is refused first, then a controller or
// router another installation created, and a namespace that is missing or
// owned by someone else (workloadowner), then the kept-CRD primitive derives
// the definitions from the same release asset and refuses foreign or newer
// ones. Then the Namespace when this resource owns it, then everything else,
// after the definitions exist, with the rollout awaited: the controller holds no finalizers and registers no
// webhook at the pinned release, so a destroy needs no live controller to
// drain anything, and nothing deadlocks waiting for the controller to start
// before its definitions. A finished deploy therefore means a serving
// controller (the module adds readiness probes on upstream's own health
// endpoints).
//
// DESTROY: the controller, the router and their RBAC are deleted, and the
// Namespace when this resource created it; the definitions are kept unless
// crds.keep_on_uninstall is false, so no Sandbox, pod or volume is lost with
// the controller.
func Resources(ctx *pulumi.Context, iacInput *kubernetesagentsandboxcontrollerv1alpha1.KubernetesAgentSandboxControllerIacInput) error {
	locals := initializeLocals(iacInput)
	s := resolveSettings(locals.Spec)
	s.identityLabels = locals.IdentityLabels
	if s.requestedNamespace != vars.Namespace {
		return namespaceNotUpstreams(s.requestedNamespace)
	}

	files, err := fetchUpstream(s)
	if err != nil {
		return err
	}
	composed, err := compose(s, files)
	if err != nil {
		return err
	}

	namespaces := []workloadowner.Namespace{{Name: vars.Namespace, Create: s.createNamespace}}
	if err := workloadowner.Refuse(iacInput.ProviderConfig, kindName, namespaces, ownedDeployments(s)); err != nil {
		return err
	}

	crds := locals.Spec.GetCrds()
	createdCrds, err := keptcrds.Apply(ctx, keptcrds.Args{
		Source: helmcrds.Source{
			BundleURL:  CRDBundleURLTemplate(s.extensions),
			BundleName: vars.BundleName,
			Version:    vars.Release,
		},
		Policy:          helmcrds.Policy{ExpectCRDs: true},
		ReleaseName:     locals.ResourceName,
		Namespace:       vars.Namespace,
		Install:         crds == nil || crds.Install == nil || crds.GetInstall(),
		KeepOnUninstall: crds == nil || crds.KeepOnUninstall == nil || crds.GetKeepOnUninstall(),
		ProviderConfig:  iacInput.ProviderConfig,
		ProviderName:    "kubernetes-crd-upsert",
	})
	if err != nil {
		return errors.Wrap(err, "failed to apply the agent-sandbox resource definitions")
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	dependsOn := createdCrds
	if composed.namespace != nil {
		namespaceGroup, err := pulumiyaml.NewConfigGroup(ctx, locals.ResourceName+"-namespace",
			&pulumiyaml.ConfigGroupArgs{Objs: []map[string]interface{}{composed.namespace}},
			pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrap(err, "failed to apply the agent-sandbox namespace")
		}
		dependsOn = append([]pulumi.Resource{namespaceGroup}, createdCrds...)
	}

	_, err = pulumiyaml.NewConfigGroup(ctx, locals.ResourceName,
		&pulumiyaml.ConfigGroupArgs{Objs: composed.workloads},
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn(dependsOn))
	if err != nil {
		return errors.Wrap(err, "failed to apply the agent-sandbox controller")
	}

	exportOutputs(ctx, s)
	return nil
}

// fetchUpstream downloads the pinned files the settings need.
func fetchUpstream(s *settings) (upstreamFiles, error) {
	release, err := fetchText(ReleaseAssetURL(s.extensions))
	if err != nil {
		return upstreamFiles{}, err
	}
	files := upstreamFiles{release: release, router: map[string]string{}}
	for _, file := range routerFilesFor(s) {
		text, err := fetchText(RouterFileURL(file))
		if err != nil {
			return upstreamFiles{}, err
		}
		files.router[file] = text
	}
	return files, nil
}

// namespaceNotUpstreams refuses a spec.namespace reference that resolved to
// another name; the schema refuses another literal, but cannot see what a
// reference resolves to. Terraform twin: the first preflight precondition.
func namespaceNotUpstreams(requested string) error {
	return &failure.Failure{
		Observed: fmt.Sprintf("spec.namespace resolved to %q", requested),
		Meaning:  fmt.Sprintf("agent-sandbox runs only in %s: upstream's manifests, RBAC and the default sandbox network policy name it", vars.Namespace),
		NextStep: fmt.Sprintf("reference a KubernetesNamespace named %s, or set spec.namespace to %s", vars.Namespace, vars.Namespace),
	}
}

// ownedDeployments are the Deployments the module applies under upstream's
// fixed names, which no other installation may already hold.
func ownedDeployments(s *settings) []workloadowner.Workload {
	deployments := []workloadowner.Workload{{Kind: workloadowner.Deployment, Namespace: vars.Namespace, Name: vars.ControllerDeploymentName}}
	if s.router != nil {
		deployments = append(deployments, workloadowner.Workload{Kind: workloadowner.Deployment, Namespace: vars.Namespace, Name: vars.RouterDeploymentName})
	}
	return deployments
}
