package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func job(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (*batchv1.Job, error) {
	target := locals.KubernetesJob

	envVarInputs := containerenv.BuildEnvVars(target.Spec.Env, locals.EnvSecretsSecretName)
	envFromInputs := containerenv.BuildEnvFrom(target.Spec.Env)

	// Build volume mounts and volumes from spec
	volumeMounts, volumes := buildVolumeMountsAndVolumes(target.Spec.VolumeMounts)

	mainContainer := &corev1.ContainerArgs{
		Name: pulumi.String("job-container"),
		Image: pulumi.String(fmt.Sprintf("%s:%s",
			target.Spec.Image.Repo,
			target.Spec.Image.Tag)),
		Env:          corev1.EnvVarArray(envVarInputs),
		EnvFrom:      envFromInputs,
		VolumeMounts: volumeMounts,
	}

	if target.Spec.Resources != nil {
		res := corev1.ResourceRequirementsArgs{}
		if target.Spec.Resources.Limits != nil {
			res.Limits = pulumi.ToStringMap(map[string]string{
				"cpu":    target.Spec.Resources.Limits.Cpu,
				"memory": target.Spec.Resources.Limits.Memory,
			})
		}
		if target.Spec.Resources.Requests != nil {
			res.Requests = pulumi.ToStringMap(map[string]string{
				"cpu":    target.Spec.Resources.Requests.Cpu,
				"memory": target.Spec.Resources.Requests.Memory,
			})
		}
		mainContainer.Resources = res
	}

	if len(target.Spec.Command) > 0 {
		mainContainer.Command = pulumi.ToStringArray(target.Spec.Command)
	}
	if len(target.Spec.Args) > 0 {
		mainContainer.Args = pulumi.ToStringArray(target.Spec.Args)
	}

	podSpecArgs := &corev1.PodSpecArgs{
		RestartPolicy: pulumi.String(target.Spec.GetRestartPolicy()),
		Containers: corev1.ContainerArray{
			mainContainer,
		},
		Volumes: volumes,
	}

	if locals.ImagePullSecretData != nil {
		podSpecArgs.ImagePullSecrets = corev1.LocalObjectReferenceArray{
			corev1.LocalObjectReferenceArgs{
				Name: pulumi.String(locals.ImagePullSecretName),
			},
		}
	}

	// Build JobSpec
	jobSpec := &batchv1.JobSpecArgs{
		Template: &corev1.PodTemplateSpecArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: podSpecArgs,
		},
	}

	// Set parallelism if specified
	if target.Spec.Parallelism != nil && *target.Spec.Parallelism > 0 {
		jobSpec.Parallelism = pulumi.IntPtr(int(*target.Spec.Parallelism))
	}

	// Set completions if specified
	if target.Spec.Completions != nil && *target.Spec.Completions > 0 {
		jobSpec.Completions = pulumi.IntPtr(int(*target.Spec.Completions))
	}

	// Set backoff limit
	if target.Spec.BackoffLimit != nil {
		jobSpec.BackoffLimit = pulumi.IntPtr(int(*target.Spec.BackoffLimit))
	}

	// Set active deadline seconds if specified and non-zero
	if target.Spec.ActiveDeadlineSeconds != nil && *target.Spec.ActiveDeadlineSeconds > 0 {
		jobSpec.ActiveDeadlineSeconds = pulumi.IntPtr(int(*target.Spec.ActiveDeadlineSeconds))
	}

	// Set TTL seconds after finished if specified and non-zero
	if target.Spec.TtlSecondsAfterFinished != nil && *target.Spec.TtlSecondsAfterFinished > 0 {
		jobSpec.TtlSecondsAfterFinished = pulumi.IntPtr(int(*target.Spec.TtlSecondsAfterFinished))
	}

	// Set completion mode if specified
	if target.Spec.CompletionMode != nil && *target.Spec.CompletionMode != "" {
		jobSpec.CompletionMode = pulumi.String(*target.Spec.CompletionMode)
	}

	// Set suspend if specified
	if target.Spec.Suspend != nil {
		jobSpec.Suspend = pulumi.BoolPtr(*target.Spec.Suspend)
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdJob, err := batchv1.NewJob(ctx,
		target.Metadata.Name,
		&batchv1.JobArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: jobSpec,
		},
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create job")
	}

	return createdJob, nil
}
