package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func cronJob(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) (*batchv1.CronJob, error) {
	target := locals.KubernetesCronJob

	envVarInputs := containerenv.BuildEnvVars(target.Spec.Env, locals.EnvSecretsSecretName)
	envFromInputs := containerenv.BuildEnvFrom(target.Spec.Env)

	// Build volume mounts and volumes from spec
	volumeMounts, volumes := buildVolumeMountsAndVolumes(target.Spec.VolumeMounts)

	mainContainer := &corev1.ContainerArgs{
		Name: pulumi.String("cronjob-container"),
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

	cronJobSpec := &batchv1.CronJobSpecArgs{
		Schedule:                   pulumi.String(target.Spec.Schedule),
		ConcurrencyPolicy:          pulumi.String(target.Spec.GetConcurrencyPolicy()),
		Suspend:                    pulumi.BoolPtr(target.Spec.GetSuspend()),
		SuccessfulJobsHistoryLimit: pulumi.IntPtr(int(target.Spec.GetSuccessfulJobsHistoryLimit())),
		FailedJobsHistoryLimit:     pulumi.IntPtr(int(target.Spec.GetFailedJobsHistoryLimit())),
		JobTemplate: &batchv1.JobTemplateSpecArgs{
			Spec: &batchv1.JobSpecArgs{
				BackoffLimit: pulumi.IntPtr(int(target.Spec.GetBackoffLimit())),
				Template: &corev1.PodTemplateSpecArgs{
					Spec: podSpecArgs,
				},
			},
		},
	}

	if target.Spec.StartingDeadlineSeconds != nil && *target.Spec.StartingDeadlineSeconds > 0 {
		cronJobSpec.StartingDeadlineSeconds = pulumi.IntPtr(int(*target.Spec.StartingDeadlineSeconds))
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdCronJob, err := batchv1.NewCronJob(ctx,
		target.Metadata.Name,
		&batchv1.CronJobArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: cronJobSpec,
		},
		opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cronjob")
	}

	return createdCronJob, nil
}
