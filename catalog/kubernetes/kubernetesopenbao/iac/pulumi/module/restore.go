package module

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// A restore Job is a RUN, not a state: it fetches one snapshot and
// installs it exactly once. Declarative semantics therefore hinge on the
// object's NAME — `<name>-restore-<8 hex>` where the suffix hashes the
// declaration (which snapshot). An unchanged declaration keeps the same
// name and is a no-op on every apply; a changed declaration (a different
// key, or latest instead of a key) is a NEW Job and a new run. The
// Terraform twin (local.restore_job_name in locals.tf) hashes the
// identical canonical string, so both engines name the same run the
// same way.
//
// Two engine defaults are deliberately overridden, in both engines:
//   - the deploy never WAITS on this Job (`pulumi.com/skipAwait`;
//     Terraform `wait_for_completion = false`). The Job waits for the
//     operator's one manual step — creating the root-token Secret after
//     `bao operator init` — and an awaited Job would hang the deploy
//     until timeout and report a failure nobody caused;
//   - the Job never EXPIRES (no ttlSecondsAfterFinished). A vanished Job
//     would be recreated on the next apply and restore AGAIN over live
//     data; the one-shot posture depends on the object persisting until
//     the operator removes `restore` from the spec.
// It also carries no activeDeadlineSeconds: the token wait is unbounded
// by design, and the bound that matters lives inside restore.sh.

// restoreJobName derives the Job's name from the declaration.
func restoreJobName(releaseName string, restore *kubernetesopenbaov1alpha1.KubernetesOpenBaoRestore) string {
	parts := []string{restore.GetSnapshotKey(), strconv.FormatBool(restore.GetLatest())}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return releaseName + vars.RestoreJobPrefix + hex.EncodeToString(sum[:])[:8]
}

// restoreJob renders the one-shot restore: an init container fetches the
// declared snapshot (or the newest under the prefix) with rclone, the
// main container installs it with the bao CLI using the initial root
// token the operator placed in the `root_token` Secret. The main
// container cannot start until that Secret exists — Kubernetes holds the
// pod in CreateContainerConfigError — which is the declared shape of
// "waiting for your one manual step".
func restoreJob(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.Resource) error {
	backup := locals.Spec.GetBackup()
	restore := locals.Spec.GetRestore()

	credentials, err := backupCredentialsSecretData(backup)
	if err != nil {
		return err
	}

	fetchExtra := map[string]string{}
	if restore.GetLatest() {
		fetchExtra["RESTORE_LATEST"] = "true"
	} else {
		fetchExtra["RESTORE_SNAPSHOT_KEY"] = strings.TrimLeft(restore.GetSnapshotKey(), "/")
	}

	restoreEnv := baoClientEnv(locals)
	restoreEnv["ROOT_TOKEN_SECRET"] = restore.GetRootToken().GetName()
	restoreEnv["ROOT_TOKEN_KEY"] = restore.GetRootToken().GetKey()
	restoreEnv["RELEASE_NAME"] = locals.ReleaseName
	restoreEnv["RELEASE_NAMESPACE"] = locals.Namespace
	restoreContainerEnv := plainEnv(restoreEnv)
	restoreContainerEnv = append(restoreContainerEnv, &kubernetescorev1.EnvVarArgs{
		Name: pulumi.String("BAO_TOKEN"),
		ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
			SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
				Name: pulumi.String(restore.GetRootToken().GetName()),
				Key:  pulumi.String(restore.GetRootToken().GetKey()),
			},
		},
	})

	resources := resourcesBlock(backup.GetResources())

	podSpec := &kubernetescorev1.PodSpecArgs{
		ServiceAccountName: pulumi.String(locals.BackupName),
		RestartPolicy:      pulumi.String("OnFailure"),
		SecurityContext:    jobPodSecurityContext(),
		InitContainers: kubernetescorev1.ContainerArray{
			&kubernetescorev1.ContainerArgs{
				Name:            pulumi.String("fetch"),
				Image:           pulumi.String(locals.RcloneImage),
				Command:         pulumi.ToStringArray([]string{"sh", vars.ScriptsMountPath + "/" + scriptFetch}),
				Env:             rcloneEnv(locals, fetchExtra),
				VolumeMounts:    rcloneContainerMounts(credentials),
				SecurityContext: jobContainerSecurityContext(),
				Resources:       resourceRequirements(resources),
			},
		},
		Containers: kubernetescorev1.ContainerArray{
			&kubernetescorev1.ContainerArgs{
				Name:            pulumi.String("restore"),
				Image:           pulumi.String(locals.OpenBaoImage),
				Command:         pulumi.ToStringArray([]string{"sh", vars.ScriptsMountPath + "/" + scriptRestore}),
				Env:             restoreContainerEnv,
				VolumeMounts:    baoContainerMounts(locals),
				SecurityContext: jobContainerSecurityContext(),
				Resources:       resourceRequirements(resources),
			},
		},
		Volumes: jobVolumes(locals, credentials),
	}
	if sched := locals.Spec.GetServer().GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			podSpec.NodeSelector = pulumi.ToStringMap(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			podSpec.Tolerations = tolerationArray(sched.GetTolerations())
		}
	}

	_, err = batchv1.NewJob(ctx, locals.RestoreJobName,
		&batchv1.JobArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.RestoreJobName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				// skipAwait rides the object annotations but is Pulumi
				// engine metadata: without it the provider waits for the
				// Job's Complete condition, i.e. for the operator.
				Annotations: pulumi.ToStringMap(map[string]string{"pulumi.com/skipAwait": "true"}),
			}),
			Spec: &batchv1.JobSpecArgs{
				// The preflight in restore.sh may meet a target still
				// unsealing and exit to retry; the install itself is
				// attempted at most this many times, then the Job stays
				// failed and visible with its log.
				BackoffLimit: pulumi.Int(vars.RestoreJobBackoffLimit),
				Template: &kubernetescorev1.PodTemplateSpecArgs{
					Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
						Labels: pulumi.ToStringMap(jobPodLabels(locals)),
					}),
					Spec: podSpec,
				},
			},
		},
		append([]pulumi.ResourceOption{
			pulumi.Provider(kubernetesProvider),
			// Job specs are immutable after creation; a drifted field is
			// replaced, never patched, and the old object goes first
			// because the name is fixed (the kubernetesjob shape).
			pulumi.ReplaceOnChanges([]string{"spec"}),
			pulumi.DeleteBeforeReplace(true),
		}, pulumi.DependsOn(dependsOn))...)
	if err != nil {
		return errors.Wrap(err, "failed to create restore job")
	}
	return nil
}
