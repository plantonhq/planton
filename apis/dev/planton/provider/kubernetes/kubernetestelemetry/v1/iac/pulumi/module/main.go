package module

import (
	"github.com/pkg/errors"
	istioapi "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetestelemetryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestelemetry/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetestelemetryv1.KubernetesTelemetryStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createTelemetry(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create telemetry")
	}

	ctx.Export(OpTelemetryName, pulumi.String(locals.TelemetryName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createTelemetry creates the namespaced Istio Telemetry resource.
//
// Unlike every other typed Istio component, Telemetry is provisioned via the generic
// apiextensions.CustomResource rather than a crd2pulumi-generated typed resource. The
// reason is a concrete crd2pulumi limitation: the Telemetry CRD's
// `spec.tracing[].customTags` field is a map whose values are nested objects with a
// `oneOf` over `{literal, environment, header}`. crd2pulumi (confirmed through v1.6.0)
// degrades that map to `map[string]map[string]string`, which structurally cannot carry
// the nested `{literal: {value: "..."}}` object, so the typed SDK cannot express a
// real custom tag at all. Using the typed resource here would silently drop a valid,
// upstream-supported configuration -- a fidelity break. The spec is therefore assembled
// from the strongly-typed proto getters below (so the input side stays type-safe) and
// passed through the generic CustomResource, which preserves the nested shape exactly.
//
// Every other field crd2pulumi types correctly, but mixing a typed resource with one
// untyped field is not possible, so the whole resource is built this way. If a future
// crd2pulumi gains support for object-valued additionalProperties maps, this can move
// to the typed istio telemetry SDK like its siblings.
func createTelemetry(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesTelemetry.Spec

	_, err := apiextensions.NewCustomResource(ctx, locals.TelemetryName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("telemetry.istio.io/v1"),
			Kind:       pulumi.String("Telemetry"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.TelemetryName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": buildTelemetrySpec(spec),
			},
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildTelemetrySpec maps the typed Planton spec to the Istio Telemetry CR `spec`
// (camelCase CRD JSON keys). Every block is attached only when present, so unset
// fields fall through to istiod's defaults.
func buildTelemetrySpec(spec *kubernetestelemetryv1.KubernetesTelemetrySpec) map[string]interface{} {
	out := map[string]interface{}{}

	if selector := spec.GetSelector(); selector != nil && len(selector.GetMatchLabels()) > 0 {
		out["selector"] = map[string]interface{}{"matchLabels": selector.GetMatchLabels()}
	}

	if refs := spec.GetTargetRefs(); len(refs) > 0 {
		out["targetRefs"] = buildTargetRefs(refs)
	}

	if tracing := spec.GetTracing(); len(tracing) > 0 {
		out["tracing"] = buildTracingList(tracing)
	}

	if metrics := spec.GetMetrics(); len(metrics) > 0 {
		out["metrics"] = buildMetricsList(metrics)
	}

	if accessLogging := spec.GetAccessLogging(); len(accessLogging) > 0 {
		out["accessLogging"] = buildAccessLoggingList(accessLogging)
	}

	return out
}

func buildTargetRefs(refs []*istioapi.KubernetesIstioApiPolicyTargetReference) []interface{} {
	out := make([]interface{}, 0, len(refs))
	for _, ref := range refs {
		m := map[string]interface{}{
			"kind": ref.GetKind(),
			"name": ref.GetName(),
		}
		if ref.GetGroup() != "" {
			m["group"] = ref.GetGroup()
		}
		if ref.GetNamespace() != "" {
			m["namespace"] = ref.GetNamespace()
		}
		out = append(out, m)
	}
	return out
}

// buildProviderRefs maps the shared ProviderRef list (reused by tracing, metrics, and
// access logging) to its CRD JSON shape.
func buildProviderRefs(refs []*kubernetestelemetryv1.KubernetesTelemetryProviderRef) []interface{} {
	out := make([]interface{}, 0, len(refs))
	for _, ref := range refs {
		out = append(out, map[string]interface{}{"name": ref.GetName()})
	}
	return out
}

func buildTracingList(list []*kubernetestelemetryv1.KubernetesTelemetryTracing) []interface{} {
	out := make([]interface{}, 0, len(list))
	for _, t := range list {
		m := map[string]interface{}{}
		if match := t.GetMatch(); match != nil && match.Mode != nil {
			m["match"] = map[string]interface{}{"mode": match.GetMode()}
		}
		if providers := t.GetProviders(); len(providers) > 0 {
			m["providers"] = buildProviderRefs(providers)
		}
		if t.RandomSamplingPercentage != nil {
			m["randomSamplingPercentage"] = t.GetRandomSamplingPercentage()
		}
		if t.DisableSpanReporting != nil {
			m["disableSpanReporting"] = t.GetDisableSpanReporting()
		}
		if tags := t.GetCustomTags(); len(tags) > 0 {
			m["customTags"] = buildCustomTags(tags)
		}
		if t.EnableIstioTags != nil {
			m["enableIstioTags"] = t.GetEnableIstioTags()
		}
		if t.UseRequestIdForTraceSampling != nil {
			m["useRequestIdForTraceSampling"] = t.GetUseRequestIdForTraceSampling()
		}
		out = append(out, m)
	}
	return out
}

// buildCustomTags maps the custom-tag map to its CRD JSON shape. Each tag carries
// exactly one source (literal/environment/header); only the set source is emitted, so
// the resulting object satisfies the CRD's oneOf. This nested shape is the precise
// reason this component uses an untyped CustomResource (see createTelemetry).
func buildCustomTags(tags map[string]*kubernetestelemetryv1.KubernetesTelemetryCustomTag) map[string]interface{} {
	out := make(map[string]interface{}, len(tags))
	for name, tag := range tags {
		entry := map[string]interface{}{}
		switch {
		case tag.GetLiteral() != nil:
			entry["literal"] = map[string]interface{}{"value": tag.GetLiteral().GetValue()}
		case tag.GetEnvironment() != nil:
			env := tag.GetEnvironment()
			e := map[string]interface{}{"name": env.GetName()}
			if env.GetDefaultValue() != "" {
				e["defaultValue"] = env.GetDefaultValue()
			}
			entry["environment"] = e
		case tag.GetHeader() != nil:
			hdr := tag.GetHeader()
			h := map[string]interface{}{"name": hdr.GetName()}
			if hdr.GetDefaultValue() != "" {
				h["defaultValue"] = hdr.GetDefaultValue()
			}
			entry["header"] = h
		}
		out[name] = entry
	}
	return out
}

func buildMetricsList(list []*kubernetestelemetryv1.KubernetesTelemetryMetrics) []interface{} {
	out := make([]interface{}, 0, len(list))
	for _, metric := range list {
		m := map[string]interface{}{}
		if providers := metric.GetProviders(); len(providers) > 0 {
			m["providers"] = buildProviderRefs(providers)
		}
		if overrides := metric.GetOverrides(); len(overrides) > 0 {
			m["overrides"] = buildMetricsOverrides(overrides)
		}
		if metric.ReportingInterval != nil {
			m["reportingInterval"] = metric.GetReportingInterval()
		}
		out = append(out, m)
	}
	return out
}

func buildMetricsOverrides(list []*kubernetestelemetryv1.KubernetesTelemetryMetricsOverride) []interface{} {
	out := make([]interface{}, 0, len(list))
	for _, o := range list {
		m := map[string]interface{}{}
		if match := o.GetMatch(); match != nil {
			if mm := buildMetricSelector(match); len(mm) > 0 {
				m["match"] = mm
			}
		}
		if o.Disabled != nil {
			m["disabled"] = o.GetDisabled()
		}
		if tagOverrides := o.GetTagOverrides(); len(tagOverrides) > 0 {
			m["tagOverrides"] = buildTagOverrides(tagOverrides)
		}
		out = append(out, m)
	}
	return out
}

func buildMetricSelector(match *kubernetestelemetryv1.KubernetesTelemetryMetricSelector) map[string]interface{} {
	m := map[string]interface{}{}
	switch {
	case match.Metric != nil:
		m["metric"] = match.GetMetric()
	case match.CustomMetric != nil:
		m["customMetric"] = match.GetCustomMetric()
	}
	if match.Mode != nil {
		m["mode"] = match.GetMode()
	}
	return m
}

func buildTagOverrides(tags map[string]*kubernetestelemetryv1.KubernetesTelemetryTagOverride) map[string]interface{} {
	out := make(map[string]interface{}, len(tags))
	for name, t := range tags {
		entry := map[string]interface{}{}
		if t.Operation != nil {
			entry["operation"] = t.GetOperation()
		}
		if t.Value != nil {
			entry["value"] = t.GetValue()
		}
		out[name] = entry
	}
	return out
}

func buildAccessLoggingList(list []*kubernetestelemetryv1.KubernetesTelemetryAccessLogging) []interface{} {
	out := make([]interface{}, 0, len(list))
	for _, al := range list {
		m := map[string]interface{}{}
		if match := al.GetMatch(); match != nil && match.Mode != nil {
			m["match"] = map[string]interface{}{"mode": match.GetMode()}
		}
		if providers := al.GetProviders(); len(providers) > 0 {
			m["providers"] = buildProviderRefs(providers)
		}
		if al.Disabled != nil {
			m["disabled"] = al.GetDisabled()
		}
		if filter := al.GetFilter(); filter != nil {
			m["filter"] = map[string]interface{}{"expression": filter.GetExpression()}
		}
		out = append(out, m)
	}
	return out
}
