package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesrbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// agentReader is the resolved spec.agent_reader: agent teammates' read-only
// way into Grafana. Grafana cannot provision service accounts from files,
// so a Job run after the release is Ready keeps the account and its one
// token (scripts.go). Twin of the agent_reader locals in the Terraform
// module's locals.tf.
type agentReader struct {
	// `<name>-agent-reader`: the ServiceAccount the Job runs as, its Role
	// and RoleBinding, and the token Secret the Job writes and the
	// ServiceAccount owns.
	Name string
	// The Grafana service account and the token's generation, defaulted.
	ServiceAccountName string
	TokenGeneration    int32
	Disabled           bool
	// The scripts ConfigMap and the Job (`<name>-agent-reader-<8 hex>`).
	ScriptName string
	JobName    string
	// The Job's image reference and optional pull Secret.
	Image          string
	PullSecretName string
	// Where the Job reads the admin's basic credentials.
	AdminSecretName  string
	AdminUserKey     string
	AdminPasswordKey string
	// The token Secret's labels, as the JSON object the script stamps.
	SecretLabelsJSON string
}

// buildAgentReader resolves spec.agent_reader; nil when the block is not
// declared.
func buildAgentReader(spec *kubernetesgrafanav1alpha1.KubernetesGrafanaSpec, releaseName,
	adminSecretName string, labels map[string]string) *agentReader {
	declared := spec.GetAgentReader()
	if declared == nil {
		return nil
	}

	serviceAccountName := declared.GetServiceAccountName()
	if serviceAccountName == "" {
		serviceAccountName = vars.DefaultAgentReaderServiceAccount
	}
	tokenGeneration := declared.GetTokenGeneration()
	if tokenGeneration == 0 {
		tokenGeneration = vars.DefaultAgentReaderTokenGeneration
	}

	// The admin keys follow the credential arm: the declared Secret's keys
	// (defaulting to the chart's names) or the chart-generated Secret's.
	userKey, passwordKey := vars.AdminUserKey, vars.AdminPasswordKey
	if existing := spec.GetAdminSecret(); existing != nil {
		if existing.GetUserKey() != "" {
			userKey = existing.GetUserKey()
		}
		if existing.GetPasswordKey() != "" {
			passwordKey = existing.GetPasswordKey()
		}
	}

	// encoding/json sorts map keys, as OpenTofu's jsonencode does.
	secretLabels, _ := json.Marshal(labels)

	repo, tag := vars.DefaultAgentReaderImageRepo, vars.DefaultAgentReaderImageTag
	if declared.GetImage().GetRepo() != "" {
		repo = declared.GetImage().GetRepo()
	}
	if declared.GetImage().GetTag() != "" {
		tag = declared.GetImage().GetTag()
	}

	return &agentReader{
		Name:               releaseName + vars.AgentReaderSuffix,
		ServiceAccountName: serviceAccountName,
		TokenGeneration:    tokenGeneration,
		Disabled:           declared.GetDisabled(),
		ScriptName:         releaseName + vars.AgentReaderScriptSuffix,
		JobName:            agentReaderJobName(releaseName, serviceAccountName, tokenGeneration, declared.GetDisabled()),
		Image:              repo + ":" + tag,
		PullSecretName:     declared.GetImage().GetPullSecretName(),
		AdminSecretName:    adminSecretName,
		AdminUserKey:       userKey,
		AdminPasswordKey:   passwordKey,
		SecretLabelsJSON:   string(secretLabels),
	}
}

// agentReaderJobName names the Job by a hash of what it reconciles: the
// account, the generation, whether it is disabled, and the script's own
// text. The Job is a RUN, so its name is its identity: an unchanged
// declaration keeps the name and is a no-op on every apply, and a changed
// one (a raised generation, `disabled`, a fixed script) is a new Job and a
// new run. Re-running is always safe — the script changes nothing when the
// stored token answers — which is why the script's text is hashed too, so
// a module upgrade that fixes the script reaches every Grafana on its next
// apply. The Terraform twin (local.agent_reader_job_name in locals.tf)
// hashes the identical canonical string.
func agentReaderJobName(releaseName, serviceAccountName string, tokenGeneration int32, disabled bool) string {
	scriptSum := sha256.Sum256([]byte(agentReaderScript))
	parts := []string{
		serviceAccountName,
		strconv.Itoa(int(tokenGeneration)),
		strconv.FormatBool(disabled),
		hex.EncodeToString(scriptSum[:]),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return releaseName + vars.AgentReaderSuffix + "-" + hex.EncodeToString(sum[:])[:8]
}

// agentReaderResources renders the ServiceAccount the Job runs as, a Role
// and RoleBinding granting it exactly the token Secret, the scripts
// ConfigMap, and the Job. They are created after the release (dependsOn),
// because the Job talks to a running Grafana.
//
// The token Secret is NOT a resource here. The Job writes it with an owner
// reference to the ServiceAccount: Kubernetes deletes it when the
// ServiceAccount goes (the block removed, or the resource destroyed), and
// the token never passes through deploy state, where a module-owned
// Secret's data would land on every refresh.
//
// The deploy waits for the Job (Pulumi's default await of a Job's Complete
// condition; Terraform `wait_for_completion`): it depends on nothing but a
// Ready Grafana, so a token that fails to mint fails this deploy rather
// than an agent's first call. It carries no ttlSecondsAfterFinished: a
// finished Job that vanished would be recreated on the next apply as
// drift, and the Job object is where its log stays readable.
func agentReaderResources(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.Resource) error {
	reader := locals.AgentReader
	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider), pulumi.DependsOn(dependsOn)}
	meta := func(name string) kubernetesmeta.ObjectMetaPtrInput {
		return kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name:      pulumi.String(name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		})
	}

	serviceAccount, err := kubernetescorev1.NewServiceAccount(ctx, reader.Name,
		&kubernetescorev1.ServiceAccountArgs{Metadata: meta(reader.Name)}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create agent-reader service account")
	}

	// Kubernetes cannot scope `create` to a name (the name is not known
	// before the object exists), so create is namespace-wide; everything
	// else names the one Secret, and `get` names the one ServiceAccount
	// whose uid the owner reference carries.
	role, err := kubernetesrbacv1.NewRole(ctx, reader.Name, &kubernetesrbacv1.RoleArgs{
		Metadata: meta(reader.Name),
		Rules: kubernetesrbacv1.PolicyRuleArray{
			kubernetesrbacv1.PolicyRuleArgs{
				ApiGroups:     pulumi.StringArray{pulumi.String("")},
				Resources:     pulumi.StringArray{pulumi.String("secrets")},
				ResourceNames: pulumi.StringArray{pulumi.String(reader.Name)},
				Verbs:         pulumi.ToStringArray([]string{"get", "update", "delete"}),
			},
			kubernetesrbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("")},
				Resources: pulumi.StringArray{pulumi.String("secrets")},
				Verbs:     pulumi.StringArray{pulumi.String("create")},
			},
			kubernetesrbacv1.PolicyRuleArgs{
				ApiGroups:     pulumi.StringArray{pulumi.String("")},
				Resources:     pulumi.StringArray{pulumi.String("serviceaccounts")},
				ResourceNames: pulumi.StringArray{pulumi.String(reader.Name)},
				Verbs:         pulumi.StringArray{pulumi.String("get")},
			},
		},
	}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create agent-reader role")
	}

	binding, err := kubernetesrbacv1.NewRoleBinding(ctx, reader.Name, &kubernetesrbacv1.RoleBindingArgs{
		Metadata: meta(reader.Name),
		RoleRef: kubernetesrbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String(reader.Name),
		},
		Subjects: kubernetesrbacv1.SubjectArray{
			kubernetesrbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String(reader.Name),
				Namespace: pulumi.String(locals.Namespace),
			},
		},
	}, append(opts, pulumi.DependsOn([]pulumi.Resource{role}))...)
	if err != nil {
		return errors.Wrap(err, "failed to create agent-reader role binding")
	}

	scripts, err := kubernetescorev1.NewConfigMap(ctx, reader.ScriptName, &kubernetescorev1.ConfigMapArgs{
		Metadata: meta(reader.ScriptName),
		Data:     pulumi.StringMap{agentReaderScriptKey: pulumi.String(agentReaderScript)},
	}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create agent-reader script config map")
	}

	_, err = batchv1.NewJob(ctx, reader.JobName, &batchv1.JobArgs{
		Metadata: meta(reader.JobName),
		Spec: &batchv1.JobSpecArgs{
			// The script bounds its own wait for Grafana; these bound the
			// run, retries included. A failed Job stays with its log.
			BackoffLimit:          pulumi.Int(vars.AgentReaderJobBackoffLimit),
			ActiveDeadlineSeconds: pulumi.Int(vars.AgentReaderJobDeadlineSeconds),
			Template: &kubernetescorev1.PodTemplateSpecArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Labels: pulumi.ToStringMap(locals.Labels),
				}),
				Spec: agentReaderPodSpec(locals),
			},
		},
	}, append(opts,
		pulumi.DependsOn([]pulumi.Resource{serviceAccount, role, binding, scripts}),
		// Job specs are immutable after creation; a drifted field under
		// the same name (a new image or placement) is replaced, never
		// patched, and the old object goes first because the name is
		// fixed (the kubernetesjob shape).
		pulumi.ReplaceOnChanges([]string{"spec"}),
		pulumi.DeleteBeforeReplace(true),
	)...)
	if err != nil {
		return errors.Wrap(err, "failed to create agent-reader job")
	}
	return nil
}

// agentReaderPodSpec is the Job's pod: one container running the script
// as `nobody` on a read-only root filesystem, the scripts mount, and an
// emptyDir HOME for kubectl's cache. It is placed like the Grafana pods
// (spec.scheduling).
func agentReaderPodSpec(locals *Locals) *kubernetescorev1.PodSpecArgs {
	reader := locals.AgentReader
	env := kubernetescorev1.EnvVarArray{}
	for _, kv := range agentReaderEnv(locals) {
		env = append(env, &kubernetescorev1.EnvVarArgs{Name: pulumi.String(kv[0]), Value: pulumi.String(kv[1])})
	}
	for _, sk := range [][2]string{
		{"GRAFANA_ADMIN_USER", reader.AdminUserKey},
		{"GRAFANA_ADMIN_PASSWORD", reader.AdminPasswordKey},
	} {
		env = append(env, &kubernetescorev1.EnvVarArgs{
			Name: pulumi.String(sk[0]),
			ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
				SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
					Name: pulumi.String(reader.AdminSecretName),
					Key:  pulumi.String(sk[1]),
				},
			},
		})
	}

	podSpec := &kubernetescorev1.PodSpecArgs{
		ServiceAccountName: pulumi.String(reader.Name),
		RestartPolicy:      pulumi.String("OnFailure"),
		SecurityContext: &kubernetescorev1.PodSecurityContextArgs{
			RunAsUser:    pulumi.Int(vars.AgentReaderRunAsUser),
			RunAsGroup:   pulumi.Int(vars.AgentReaderRunAsUser),
			RunAsNonRoot: pulumi.Bool(true),
			SeccompProfile: &kubernetescorev1.SeccompProfileArgs{
				Type: pulumi.String("RuntimeDefault"),
			},
		},
		Containers: kubernetescorev1.ContainerArray{
			&kubernetescorev1.ContainerArgs{
				Name:    pulumi.String("agent-reader"),
				Image:   pulumi.String(reader.Image),
				Command: pulumi.ToStringArray([]string{"sh", vars.AgentReaderScriptsPath + "/" + agentReaderScriptKey}),
				Env:     env,
				Resources: &kubernetescorev1.ResourceRequirementsArgs{
					Requests: pulumi.StringMap{
						"cpu":    pulumi.String(vars.AgentReaderCpuRequest),
						"memory": pulumi.String(vars.AgentReaderMemoryRequest),
					},
					Limits: pulumi.StringMap{
						"cpu":    pulumi.String(vars.AgentReaderCpuLimit),
						"memory": pulumi.String(vars.AgentReaderMemoryLimit),
					},
				},
				SecurityContext: &kubernetescorev1.SecurityContextArgs{
					AllowPrivilegeEscalation: pulumi.Bool(false),
					ReadOnlyRootFilesystem:   pulumi.Bool(true),
					Capabilities: &kubernetescorev1.CapabilitiesArgs{
						Drop: pulumi.StringArray{pulumi.String("ALL")},
					},
				},
				VolumeMounts: kubernetescorev1.VolumeMountArray{
					&kubernetescorev1.VolumeMountArgs{Name: pulumi.String("scripts"), MountPath: pulumi.String(vars.AgentReaderScriptsPath)},
					&kubernetescorev1.VolumeMountArgs{Name: pulumi.String("home"), MountPath: pulumi.String(vars.AgentReaderHomePath)},
				},
			},
		},
		Volumes: kubernetescorev1.VolumeArray{
			&kubernetescorev1.VolumeArgs{
				Name: pulumi.String("scripts"),
				ConfigMap: &kubernetescorev1.ConfigMapVolumeSourceArgs{
					Name:        pulumi.String(reader.ScriptName),
					DefaultMode: pulumi.Int(0555),
				},
			},
			&kubernetescorev1.VolumeArgs{
				Name:     pulumi.String("home"),
				EmptyDir: &kubernetescorev1.EmptyDirVolumeSourceArgs{},
			},
		},
	}
	if reader.PullSecretName != "" {
		podSpec.ImagePullSecrets = kubernetescorev1.LocalObjectReferenceArray{
			&kubernetescorev1.LocalObjectReferenceArgs{Name: pulumi.String(reader.PullSecretName)},
		}
	}
	if sched := locals.Spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			podSpec.NodeSelector = pulumi.ToStringMap(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			podSpec.Tolerations = tolerationArray(sched)
		}
		if sched.GetPriorityClassName() != "" {
			podSpec.PriorityClassName = pulumi.String(sched.GetPriorityClassName())
		}
	}
	return podSpec
}

// agentReaderEnv is the script's plain environment (scripts.go lists the
// contract), in a fixed order so the rendered Job is stable.
func agentReaderEnv(locals *Locals) [][2]string {
	reader := locals.AgentReader
	return [][2]string{
		{"GRAFANA_URL", locals.Endpoint},
		{"ADMIN_SECRET", reader.AdminSecretName},
		{"NAMESPACE", locals.Namespace},
		{"RELEASE_NAME", locals.ReleaseName},
		{"SERVICE_ACCOUNT", reader.ServiceAccountName},
		{"TOKEN_GENERATION", strconv.Itoa(int(reader.TokenGeneration))},
		{"DISABLED", strconv.FormatBool(reader.Disabled)},
		{"TOKEN_SECRET", reader.Name},
		{"OWNER_SERVICE_ACCOUNT", reader.Name},
		{"SECRET_LABELS", reader.SecretLabelsJSON},
		{"HOME", vars.AgentReaderHomePath},
	}
}

// tolerationArray renders the shared WorkloadToleration list onto the
// Job's pod.
func tolerationArray(sched *kubernetesgrafanav1alpha1.KubernetesGrafanaScheduling) kubernetescorev1.TolerationArray {
	out := kubernetescorev1.TolerationArray{}
	for _, t := range sched.GetTolerations() {
		tol := &kubernetescorev1.TolerationArgs{}
		if t.GetKey() != "" {
			tol.Key = pulumi.String(t.GetKey())
		}
		if t.GetOperator() != "" {
			tol.Operator = pulumi.String(t.GetOperator())
		}
		if t.GetValue() != "" {
			tol.Value = pulumi.String(t.GetValue())
		}
		if t.GetEffect() != "" {
			tol.Effect = pulumi.String(t.GetEffect())
		}
		if t.TolerationSeconds != nil {
			tol.TolerationSeconds = pulumi.Int(int(t.GetTolerationSeconds()))
		}
		out = append(out, tol)
	}
	return out
}
