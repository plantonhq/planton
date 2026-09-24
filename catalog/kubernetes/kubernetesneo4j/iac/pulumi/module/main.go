package module

import (
	"github.com/pkg/errors"
	kubernetesneo4jv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesneo4j/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Neo4j from the official Helm chart as a real Helm
// release. The typed spec renders into chart values (values.go); a declared
// admin password materializes as the "<name>-auth" Kubernetes Secret
// (secrets.go) the chart consumes via neo4j.passwordFromSecret; the
// helm_values escape hatch merges last with Helm -f semantics — the exact
// semantic twin of the Terraform module's helm_release with
// values = [typed, helm_values].
//
// ORDERING IS LOAD-BEARING: the chart looks the passwordFromSecret Secret
// up AT TEMPLATE TIME and fails the install when it is missing, so the auth
// Secret is an explicit dependency of the release — it exists first, always.
func Resources(ctx *pulumi.Context, stackInput *kubernetesneo4jv1alpha1.KubernetesNeo4JStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ auth secret ---------------------------
	createdAuthSecret, err := authSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return err
	}

	releaseDeps := namespaceDeps
	if createdAuthSecret != nil {
		releaseDeps = append(releaseDeps, pulumi.DependsOn([]pulumi.Resource{createdAuthSecret}))
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
		// Wait for the server to become Ready — a database that never
		// starts (bad image, unschedulable pod, unbindable volume, a JVM
		// that OOMs on boot) should fail THIS deploy, not the first driver
		// connection. Neo4j recovers/upgrades store files on startup, so
		// the budget is generous. SkipAwait false is Helm --wait, stated
		// explicitly to mirror the Terraform twin's `wait = true`.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(600),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, releaseDeps...)

	_, err = helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to install neo4j helm release")
	}

	exportOutputs(ctx, locals)
	return nil
}

// exportOutputs publishes the composition handles. The service name is the
// chart's always-created ClusterIP Service — neo4j.fullname = the release
// name (templates/neo4j-svc.yaml). auth_secret_name exports the
// module-materialized "<name>-auth" (declared or generated password) or
// the referenced existing Secret, never empty; password_secret names the
// bare-password key only for the Secret this module owns — an existing
// Secret's layout is the owner's, and the module never assumes it carries
// a `password` key.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpBoltEndpoint, pulumi.String(locals.BoltEndpoint))
	ctx.Export(OpHttpEndpoint, pulumi.String(locals.HttpEndpoint))
	ctx.Export(OpAuthSecretName, pulumi.String(locals.AuthSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	passwordSecretName := ""
	passwordSecretKey := ""
	if locals.CreateAuthSecret {
		passwordSecretName = locals.AuthSecretName
		passwordSecretKey = authSecretPasswordKey
	}
	ctx.Export(OpPasswordSecretName, pulumi.String(passwordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(passwordSecretKey))
}
