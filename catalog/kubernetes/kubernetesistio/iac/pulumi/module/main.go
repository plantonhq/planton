package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesistiov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesistio/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs the Istio control plane from the official Helm charts as
// real Helm releases, in upstream's own order:
//
//	CRDs (module-owned, server-side apply)
//	-> base (validation plumbing) -> istiod (the control plane)
//	-> [ambient or cni.enabled] istio-cni (node agent)
//	-> [ambient] ztunnel (per-node L4 proxy)
//
// The CRDs deliberately apply OUTSIDE the base release (the chart installs
// them with base.excludedCRDs covering the whole bundle): Helm refuses to
// adopt CRDs that already exist without ITS ownership metadata, so a cluster
// running the CRDs-only KubernetesIstioBaseCrds kind could never upgrade to
// the full mesh if the chart owned them — server-side-applied CRDs are
// co-ownable by both kinds, making that migration a plain redeploy.
//
// The typed spec renders into per-chart values (values.go); each release's
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release resources.
//
// Deliberately NO gateway release: istiod implements the Kubernetes Gateway
// API, so north-south gateways are composed from KubernetesGateway resources
// (gateway_class_name: istio) and istiod provisions their deployments itself.
func Resources(ctx *pulumi.Context, iacInput *kubernetesistiov1alpha1.KubernetesIstioIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// ------------------------------ CRDs ----------------------------------
	// Module-owned (never Helm-owned — see the function comment). The bundle
	// version is pinned to spec.version so the installed CRD schema matches
	// the control plane and the typed Istio kinds' generated SDK.
	crds, err := pulumiyaml.NewConfigFile(ctx, "istio-crds",
		&pulumiyaml.ConfigFileArgs{
			File: fmt.Sprintf(vars.CrdBundleURLTemplate, locals.Version),
		},
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to apply istio CRDs")
	}

	baseOpts := []pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{crds}),
	}
	if createdNamespace != nil {
		baseOpts = append(baseOpts, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ base ---------------------------------
	baseValues, err := buildBaseValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build base chart values")
	}

	baseRelease, err := helmv3.NewRelease(ctx, vars.BaseReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(vars.BaseReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.BaseChart),
		Version:   pulumi.String(locals.Version),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmRepo),
		},
		Values: pulumi.ToMap(baseValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		Atomic:          pulumi.Bool(true),
		CleanupOnFail:   pulumi.Bool(true),
		Timeout:         pulumi.Int(300),
	}, baseOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install istio base helm release")
	}

	// ------------------------------ istiod --------------------------------
	istiodValues, err := buildIstiodValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build istiod chart values")
	}

	istiodRelease, err := helmv3.NewRelease(ctx, locals.IstiodReleaseName, &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.IstiodReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.IstiodChart),
		Version:   pulumi.String(locals.Version),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmRepo),
		},
		Values:          pulumi.ToMap(istiodValues),
		CreateNamespace: pulumi.Bool(false),
		// Wait for istiod to be Ready: a control plane whose webhooks and
		// discovery service are not serving rejects every mesh-config apply
		// and every injection, so a premature "success" would just move the
		// failure downstream.
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{baseRelease}))
	if err != nil {
		return errors.Wrap(err, "failed to install istiod helm release")
	}

	// ------------------------------ cni (ambient or opt-in) ----------------
	if locals.InstallCni {
		cniValues, err := buildCniValues(locals)
		if err != nil {
			return errors.Wrap(err, "failed to build cni chart values")
		}

		_, err = helmv3.NewRelease(ctx, vars.CniReleaseName, &helmv3.ReleaseArgs{
			Name:      pulumi.String(vars.CniReleaseName),
			Namespace: pulumi.String(locals.Namespace),
			Chart:     pulumi.String(vars.CniChart),
			Version:   pulumi.String(locals.Version),
			RepositoryOpts: &helmv3.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmRepo),
			},
			Values:          pulumi.ToMap(cniValues),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			Timeout:         pulumi.Int(600),
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{istiodRelease}))
		if err != nil {
			return errors.Wrap(err, "failed to install istio-cni helm release")
		}
	}

	// ------------------------------ ztunnel (ambient only) -----------------
	if locals.Ambient {
		ztunnelValues, err := buildZtunnelValues(locals)
		if err != nil {
			return errors.Wrap(err, "failed to build ztunnel chart values")
		}

		_, err = helmv3.NewRelease(ctx, vars.ZtunnelReleaseName, &helmv3.ReleaseArgs{
			Name:      pulumi.String(vars.ZtunnelReleaseName),
			Namespace: pulumi.String(locals.Namespace),
			Chart:     pulumi.String(vars.ZtunnelChart),
			Version:   pulumi.String(locals.Version),
			RepositoryOpts: &helmv3.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmRepo),
			},
			Values:          pulumi.ToMap(ztunnelValues),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			Timeout:         pulumi.Int(600),
		}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{istiodRelease}))
		if err != nil {
			return errors.Wrap(err, "failed to install ztunnel helm release")
		}
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpIstiodServiceName, pulumi.String(locals.IstiodServiceName))
	ctx.Export(OpRevision, pulumi.String(locals.Revision))
	ctx.Export(OpGatewayClassName, pulumi.String(vars.GatewayClassName))
	ctx.Export(OpTrustDomain, pulumi.String(locals.TrustDomain))
	ctx.Export(OpDataplaneMode, pulumi.String(locals.DataplaneMode))

	return nil
}
