package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	kubernetescronjobv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetescronjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesCronJob *kubernetescronjobv1alpha1.KubernetesCronJob
	Namespace         string

	// Labels are the full governance label set stamped on every created object
	// and on the pod template of every scheduled run's Job (workload identity +
	// Planton resource tracking).
	Labels map[string]string

	// Computed satellite resource names, prefixed with metadata.name so multiple
	// instances sharing a namespace never collide.
	EnvSecretName       string
	ImagePullSecretName string
	ImagePullSecretData map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *kubernetescronjobv1alpha1.KubernetesCronJobIacInput) (*Locals, error) {
	locals := &Locals{}

	locals.KubernetesCronJob = iacInput.Target
	target := iacInput.Target

	if target.Spec.JobTemplate == nil || target.Spec.JobTemplate.Container == nil ||
		target.Spec.JobTemplate.Container.App == nil || target.Spec.JobTemplate.Container.App.Image == nil {
		return nil, errors.New("spec.job_template.container.app.image is required")
	}

	locals.Labels = map[string]string{
		"app":                            target.Metadata.Name,
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesCronJob.String(),
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

	locals.Namespace = target.Spec.Namespace.GetValue()
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpCronJobName, pulumi.String(target.Metadata.Name))

	// The schedule is exported so dependents and audits read the deployed truth
	// from outputs rather than from the spec.
	ctx.Export(OpSchedule, pulumi.String(target.Spec.Schedule))

	locals.EnvSecretName = fmt.Sprintf("%s-env-secrets", target.Metadata.Name)
	locals.ImagePullSecretName = fmt.Sprintf("%s-image-pull", target.Metadata.Name)

	// The image-pull Secret's data comes from the workload's OWN spec — the registry
	// logins declared on pod.image_registries — and from nowhere else. Nil when the
	// pod declares none: a public image, or a same-cloud registry the cluster's own
	// identity reaches, or a Secret declared beside the workload and named in
	// pod.image_pull_secrets, all need no Secret from this module.
	imagePullSecretData, err := workloadpod.BuildImagePullSecretData(target.Spec.JobTemplate.Pod)
	if err != nil {
		return nil, err
	}
	locals.ImagePullSecretData = imagePullSecretData

	return locals, nil
}
