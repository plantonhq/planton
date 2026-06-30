package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetesrookcephoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesrookcephoperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed configuration values from the stack input
type Locals struct {
	// KubernetesRookCephOperator is the target resource
	KubernetesRookCephOperator *kubernetesrookcephoperatorv1.KubernetesRookCephOperator

	// Namespace is the Kubernetes namespace to deploy to
	Namespace string

	// Labels are common labels applied to all resources
	Labels map[string]string

	// HelmReleaseName is the name of the Helm release
	HelmReleaseName string

	// ChartVersion is the Helm chart version to install (without 'v' prefix)
	ChartVersion string

	// HelmValues contains computed values for the Helm release
	HelmValues map[string]interface{}
}

// initializeLocals creates computed values from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesrookcephoperatorv1.KubernetesRookCephOperatorStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesRookCephOperator = stackInput.Target

	target := stackInput.Target

	// Build common labels
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesRookCephOperator.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Helm release name based on metadata name
	locals.HelmReleaseName = target.Metadata.Name
	ctx.Export(OpHelmReleaseName, pulumi.String(locals.HelmReleaseName))

	// Webhook service name
	webhookServiceName := fmt.Sprintf("%s-rook-ceph-operator", target.Metadata.Name)
	ctx.Export(OpWebhookService, pulumi.String(webhookServiceName))

	// Helm chart version without 'v' prefix
	// The default version (v1.16.6) is set in spec.proto via options.default
	operatorVersion := target.Spec.GetOperatorVersion()
	if operatorVersion == "" {
		operatorVersion = "v1.16.6"
	}
	locals.ChartVersion = strings.TrimPrefix(operatorVersion, "v")

	// Build Helm values
	locals.HelmValues = buildHelmValues(target)

	return locals
}

// buildHelmValues constructs the Helm values map from spec
func buildHelmValues(target *kubernetesrookcephoperatorv1.KubernetesRookCephOperator) map[string]interface{} {
	values := map[string]interface{}{}

	// CRDs configuration
	crdsEnabled := target.Spec.GetCrdsEnabled()
	values["crds"] = map[string]interface{}{
		"enabled": crdsEnabled,
	}

	// Resource configuration
	if target.Spec.Container != nil && target.Spec.Container.Resources != nil {
		resources := map[string]interface{}{}

		if target.Spec.Container.Resources.Requests != nil {
			requests := map[string]interface{}{}
			if target.Spec.Container.Resources.Requests.Cpu != "" {
				requests["cpu"] = target.Spec.Container.Resources.Requests.Cpu
			}
			if target.Spec.Container.Resources.Requests.Memory != "" {
				requests["memory"] = target.Spec.Container.Resources.Requests.Memory
			}
			if len(requests) > 0 {
				resources["requests"] = requests
			}
		}

		if target.Spec.Container.Resources.Limits != nil {
			limits := map[string]interface{}{}
			if target.Spec.Container.Resources.Limits.Cpu != "" {
				limits["cpu"] = target.Spec.Container.Resources.Limits.Cpu
			}
			if target.Spec.Container.Resources.Limits.Memory != "" {
				limits["memory"] = target.Spec.Container.Resources.Limits.Memory
			}
			if len(limits) > 0 {
				resources["limits"] = limits
			}
		}

		if len(resources) > 0 {
			values["resources"] = resources
		}
	}

	// CSI configuration
	if target.Spec.Csi != nil {
		csi := map[string]interface{}{}

		if target.Spec.Csi.EnableRbdDriver != nil {
			csi["enableRbdDriver"] = *target.Spec.Csi.EnableRbdDriver
		}
		if target.Spec.Csi.EnableCephfsDriver != nil {
			csi["enableCephfsDriver"] = *target.Spec.Csi.EnableCephfsDriver
		}
		if target.Spec.Csi.DisableCsiDriver != nil {
			if *target.Spec.Csi.DisableCsiDriver {
				csi["disableCsiDriver"] = "true"
			} else {
				csi["disableCsiDriver"] = "false"
			}
		}
		if target.Spec.Csi.EnableCsiHostNetwork != nil {
			csi["enableCSIHostNetwork"] = *target.Spec.Csi.EnableCsiHostNetwork
		}
		if target.Spec.Csi.ProvisionerReplicas != nil {
			csi["provisionerReplicas"] = *target.Spec.Csi.ProvisionerReplicas
		}
		if target.Spec.Csi.EnableCsiAddons != nil {
			csi["csiAddons"] = map[string]interface{}{
				"enabled": *target.Spec.Csi.EnableCsiAddons,
			}
		}
		if target.Spec.Csi.EnableNfsDriver != nil {
			csi["nfs"] = map[string]interface{}{
				"enabled": *target.Spec.Csi.EnableNfsDriver,
			}
		}

		if len(csi) > 0 {
			values["csi"] = csi
		}
	}

	return values
}
