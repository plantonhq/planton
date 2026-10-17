package module

import (
	"github.com/pkg/errors"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// helmRelease installs Langfuse from the official chart as a real Helm
// release (helm.v3 Release - Helm's own lifecycle and rollback; never the
// client-side Chart resource).
func helmRelease(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) (*helmv3.Release, error) {
	values, err := buildHelmValues(locals)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build helm values")
	}

	releaseArgs := &helmv3.ReleaseArgs{
		Name:      pulumi.String(locals.ReleaseName),
		Namespace: pulumi.String(locals.Namespace),
		Chart:     pulumi.String(vars.HelmChartName),
		Version:   pulumi.String(locals.ChartVersion),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String(vars.HelmChartRepo),
		},
		Values: pulumi.ToMap(values),
		// The module owns namespace creation (create_namespace flag).
		CreateNamespace: pulumi.Bool(false),
		// Wait for both workloads to become Ready (Helm --wait), stated
		// explicitly to mirror the Terraform twin's `wait = true`. The first
		// start runs every Postgres and ClickHouse migration before the web
		// server turns ready, hence the generous timeout.
		SkipAwait:     pulumi.Bool(false),
		Atomic:        pulumi.Bool(true),
		CleanupOnFail: pulumi.Bool(true),
		Timeout:       pulumi.Int(900),
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)
	release, err := helmv3.NewRelease(ctx, locals.ReleaseName, releaseArgs, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to install the langfuse helm release")
	}
	return release, nil
}

// buildHelmValues merges, in order: the typed values, the spec's
// helm_values escape hatch (Helm `-f` semantics: maps deep-merge with the
// later document winning, lists replace), and the re-pins - the fullname
// every output derives from, and the four bundled stores held off, so an
// override can never deploy a private database beside the composed ones.
//
// PARITY: the Terraform module reaches the same result natively - its
// helm_release passes values = [yamlencode(typed values), helm_values,
// yamlencode(re-pins)].
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	values := buildTypedValues(locals)
	if hv := locals.Spec.GetHelmValues(); hv != "" {
		override := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(hv), &override); err != nil {
			return nil, errors.Wrap(err, "spec.helm_values is not a valid YAML document")
		}
		values = mergeMaps(values, override)
	}
	return mergeMaps(values, rePinnedValues(locals)), nil
}

// rePinnedValues are merged last (Terraform twin: local.re_pinned_values).
func rePinnedValues(locals *Locals) map[string]interface{} {
	return map[string]interface{}{
		"fullnameOverride": locals.ReleaseName,
		"postgresql":       map[string]interface{}{"deploy": false},
		"redis":            map[string]interface{}{"deploy": false},
		"clickhouse":       map[string]interface{}{"deploy": false},
		"s3":               map[string]interface{}{"deploy": false},
	}
}

// mergeMaps deep-merges override into base with Helm `-f` semantics: maps
// merge recursively, every other value (lists included) replaces.
func mergeMaps(base, override map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(base))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		if bv, ok := out[k].(map[string]interface{}); ok {
			if ov, ok := v.(map[string]interface{}); ok {
				out[k] = mergeMaps(bv, ov)
				continue
			}
		}
		out[k] = v
	}
	return out
}
