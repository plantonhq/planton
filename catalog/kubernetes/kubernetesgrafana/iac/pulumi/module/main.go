package module

import (
	"github.com/pkg/errors"
	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Grafana from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); admin
// credentials stay chart-owned (generated once via the chart's lookup, or
// read from an existing Secret); database, datasource and sign-in
// credentials ride environment variables sourced from Secrets (sign-in's
// from the module-owned `<name>-sso` Secret), so no credential ever lands
// in the chart's rendered configuration; the helm_values escape hatch
// merges last with Helm -f semantics — the exact semantic twin of the Terraform
// module's helm_release with values = [typed, helm_values].
func Resources(ctx *pulumi.Context, iacInput *kubernetesgrafanav1alpha1.KubernetesGrafanaIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// NAME BUDGET: the agent reader's Job is `<name>-agent-reader-<8 hex>`
	// and Kubernetes caps a Job's name at 63 characters. Refused before
	// anything is created (twin of the release's precondition in main.tf).
	if locals.AgentReader != nil && len(locals.ReleaseName) > vars.AgentReaderNameBudget {
		return errors.Errorf("metadata.name %q is %d characters, and with agent_reader declared it is at most %d: "+
			"the agent reader's Job is <name>-agent-reader-<8 hex> under Kubernetes' 63-character Job name cap. "+
			"Use a shorter name", locals.ReleaseName, len(locals.ReleaseName), vars.AgentReaderNameBudget)
	}

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

	var releaseDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------ sign-in credentials Secret -------------------
	// Created BEFORE the release: Grafana's GF_AUTH_*_CLIENT_SECRET
	// variables read it through secretKeyRef, and a pod whose referenced
	// key is missing never starts. The values arrive resolved from
	// managed-secret references; the provider stores Secret data as
	// Pulumi secrets.
	if s := locals.SignIn; s != nil {
		createdSecret, err := kubernetescorev1.NewSecret(ctx,
			locals.SsoSecretName,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.SsoSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
				StringData: pulumi.ToSecret(pulumi.ToStringMap(s.SecretData)).(pulumi.StringMapOutput),
			},
			append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)...)
		if err != nil {
			return errors.Wrap(err, "failed to create sign-in credentials secret")
		}
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdSecret}))
	}

	// ------------------------------ helm release --------------------------
	mergedValues, err := buildHelmValues(locals)
	if err != nil {
		return errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(mergedValues),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for Grafana to become Ready — a UI that never starts (bad
		// plugin ID, unreachable database, unbindable volume) should fail
		// THIS deploy, not the first login attempt. Plugin downloads at
		// startup are what the generous budget covers. SkipAwait false is
		// Helm --wait, stated explicitly to mirror the Terraform twin's
		// `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	release, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install grafana helm release")
	}

	// ------------------------ agent teammates' way in ---------------------
	// After the release: the Job talks to a Ready Grafana (agentreader.go).
	if locals.AgentReader != nil {
		if err := agentReaderResources(ctx, locals, kubernetesProvider, []pulumi.Resource{release}); err != nil {
			return err
		}
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The service name is the
// chart's ClusterIP Service — grafana.fullname, pinned to the resource name
// via fullnameOverride. The admin Secret name follows the credential arm:
// the chart-owned `<name>` Secret for the generate arm, the referenced
// Secret's own name for the existing arm.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpService, pulumi.String(locals.ServiceName))
	ctx.Export(OpEndpoint, pulumi.String(locals.Endpoint))
	ctx.Export(OpAdminSecretName, pulumi.String(locals.AdminSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))

	// The token Secret exists while the reader is on; the Job's name
	// whenever the block is declared (a disabled reader still runs one).
	tokenSecretName, tokenSecretKey, jobName := "", "", ""
	if reader := locals.AgentReader; reader != nil {
		jobName = reader.JobName
		if !reader.Disabled {
			tokenSecretName = reader.Name
			tokenSecretKey = vars.AgentReaderTokenKey
		}
	}
	ctx.Export(OpAgentReaderTokenSecretName, pulumi.String(tokenSecretName))
	ctx.Export(OpAgentReaderTokenSecretKey, pulumi.String(tokenSecretKey))
	ctx.Export(OpAgentReaderJobName, pulumi.String(jobName))
}
