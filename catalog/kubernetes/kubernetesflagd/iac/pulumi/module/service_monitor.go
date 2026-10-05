package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceMonitor creates the optional ServiceMonitor scraping /metrics on
// the management port. Requires the Prometheus Operator CRDs. Terraform
// twin: kubectl_manifest.service_monitor with count.
func serviceMonitor(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) error {
	if !locals.Spec.GetMetrics().GetServiceMonitorEnabled() {
		return nil
	}
	labels := map[string]string{}
	for k, v := range locals.Spec.GetMetrics().GetServiceMonitorLabels() {
		labels[k] = v
	}
	for k, v := range locals.Labels {
		labels[k] = v
	}
	_, err := apiextensions.NewCustomResource(ctx, locals.ServiceMonitorName, &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("monitoring.coreos.com/v1"),
		Kind:       pulumi.String("ServiceMonitor"),
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name: pulumi.String(locals.ServiceMonitorName), Namespace: pulumi.String(locals.Namespace), Labels: pulumi.ToStringMap(labels),
		}),
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"selector":  map[string]interface{}{"matchLabels": locals.SelectorLabels},
				"endpoints": []interface{}{map[string]interface{}{"port": "management", "path": "/metrics", "interval": "30s"}},
			},
		},
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create service monitor")
	}
	return nil
}
