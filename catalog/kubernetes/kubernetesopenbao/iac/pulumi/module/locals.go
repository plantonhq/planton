package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Server modes (the spec's mode oneof; unset = standalone — the chart's
// own default).
const (
	modeDev        = "dev"
	modeStandalone = "standalone"
	modeHa         = "ha"
)

// Locals holds computed values derived from the stack input. Every
// resolution here has an exact twin in the Terraform module's locals.tf —
// keep them in lockstep.
type Locals struct {
	Spec *kubernetesopenbaov1alpha1.KubernetesOpenBaoSpec

	// Resource-identity labels stamped on module-created satellites
	// (namespace, seal-credentials Secret) — never injected into the
	// chart's own resources; Helm owns those.
	Labels map[string]string

	// Namespace the server installs into (resolved literal).
	Namespace string

	// ReleaseName is metadata.name; fullnameOverride is pinned to it,
	// so every chart-derived name hangs off this value.
	ReleaseName string

	ChartVersion string

	// Mode resolved from the spec oneof (dev / standalone / ha).
	Mode string

	// Raft peer count (ha mode; 1 otherwise).
	Replicas int

	// http or https, following tls.enabled — drives the synthesized
	// listener config, retry_join addresses, and the exported endpoint.
	Scheme string

	// True when a TLS certificate Secret is mounted.
	TlsEnabled bool

	// Resolved TLS Secret name ("" when TLS is off).
	TlsSecretName string

	// Name of the module-owned seal-credentials Secret ("" when the
	// declared seal arm carries no credential material — the keyless
	// posture).
	SealCredentialsSecretName string

	// The synthesized server config HCL (empty in dev mode — dev
	// ignores config).
	BaoConfigHcl string

	// ------------------------------ backup --------------------------------
	// True when `backup` is declared: the module renders the job
	// ServiceAccount, the scripts ConfigMap, the credentials Secret
	// (when an arm declares keys), and the CronJob.
	BackupEnabled bool

	// True when `restore` is declared: the module renders the one-shot
	// restore Job and renders the backup CronJob SUSPENDED (a fresh
	// target must not snapshot an empty vault into the shared prefix
	// while its restore waits for the operator's token).
	RestoreDeclared bool

	// `<name>-backup`: the job ServiceAccount, the OpenBao policy, and
	// the CronJob share one name so the login recipe is one noun.
	BackupName string

	// The Kubernetes-auth role the job logs in with — `<name>-backup`
	// unless `backup.auth.role` names an existing one.
	BackupAuthRole string

	// The Kubernetes-auth mount path (`kubernetes` unless declared).
	BackupAuthMountPath string

	// The scripts ConfigMap and the credentials Secret names; the
	// Secret name is "" when the declared arm carries no key material
	// (the keyless postures need no Secret).
	BackupScriptsName           string
	BackupCredentialsSecretName string

	// The address the jobs' bao CLI talks to: the active-leader Service
	// in HA (the only mode backups are legal on), scheme following TLS.
	BaoAddr string

	// `<name>-restore-<8 hex>`, "" when no restore is declared. The hex
	// hashes the declaration so an unchanged restore is a no-op on
	// every apply and a changed one is a new Job (Mongo's shape).
	RestoreJobName string

	// The two job images after `backup.images` overrides.
	OpenBaoImage string
	RcloneImage  string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesopenbaov1alpha1.KubernetesOpenBaoStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesOpenBao.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	chartVersion := spec.GetChartVersion()
	if chartVersion == "" {
		chartVersion = vars.DefaultChartVersion
	}

	// Resolve the mode oneof; unset = standalone (the chart default).
	mode := modeStandalone
	replicas := 1
	if spec.GetServer() != nil {
		switch {
		case spec.GetServer().GetDev() != nil:
			mode = modeDev
		case spec.GetServer().GetHa() != nil:
			mode = modeHa
			replicas = 3
			if spec.GetServer().GetHa().Replicas != nil {
				replicas = int(spec.GetServer().GetHa().GetReplicas())
			}
		}
	}

	tlsEnabled := spec.GetTls().GetEnabled()
	tlsSecretName := ""
	if tlsEnabled {
		tlsSecretName = spec.GetTls().GetCertSecretName().GetValue()
	}
	scheme := "http"
	if tlsEnabled {
		scheme = "https"
	}

	sealCredentialsSecretName := ""
	if sealSecretData(spec) != nil {
		sealCredentialsSecretName = target.Metadata.Name + vars.SealCredentialsSecretSuffix
	}

	locals := &Locals{
		Spec:                      spec,
		Labels:                    labels,
		Namespace:                 spec.Namespace.GetValue(),
		ReleaseName:               target.Metadata.Name,
		ChartVersion:              chartVersion,
		Mode:                      mode,
		Replicas:                  replicas,
		Scheme:                    scheme,
		TlsEnabled:                tlsEnabled,
		TlsSecretName:             tlsSecretName,
		SealCredentialsSecretName: sealCredentialsSecretName,
	}

	locals.BaoConfigHcl = renderBaoConfigHcl(locals)

	if backup := spec.GetBackup(); backup != nil {
		locals.BackupEnabled = true
		locals.BackupName = target.Metadata.Name + vars.BackupSuffix
		locals.BackupAuthRole = locals.BackupName
		if backup.GetAuth().GetRole() != "" {
			locals.BackupAuthRole = backup.GetAuth().GetRole()
		}
		locals.BackupAuthMountPath = "kubernetes"
		if backup.GetAuth() != nil && backup.GetAuth().MountPath != nil && backup.GetAuth().GetMountPath() != "" {
			locals.BackupAuthMountPath = backup.GetAuth().GetMountPath()
		}
		locals.BackupScriptsName = target.Metadata.Name + vars.BackupScriptsSuffix
		if data, _ := backupCredentialsSecretData(backup); len(data) > 0 {
			locals.BackupCredentialsSecretName = target.Metadata.Name + vars.BackupCredentialsSecretSuffix
		}
		// The active-leader Service selects exactly the unsealed leader;
		// the main Service round-robins sealed pods too (by design, for
		// init). The chart's own snapshot agent targets the same name.
		locals.BaoAddr = fmt.Sprintf("%s://%s-active.%s.svc:%d", scheme, target.Metadata.Name, locals.Namespace, vars.ApiPort)
		locals.OpenBaoImage = imageRef(backup.GetImages().GetOpenbao(), vars.DefaultOpenBaoImage)
		locals.RcloneImage = imageRef(backup.GetImages().GetRclone(), vars.DefaultRcloneImage)
	}
	if restore := spec.GetRestore(); restore != nil {
		locals.RestoreDeclared = true
		locals.RestoreJobName = restoreJobName(target.Metadata.Name, restore)
	}

	return locals
}

// imageRef renders a ContainerImage override, or the module default when
// the override is absent. A repo without a tag keeps the default's tag so
// a mirror override never silently floats to `latest`.
func imageRef(override *kubernetesprovider.ContainerImage, def string) string {
	if override == nil || override.GetRepo() == "" {
		return def
	}
	tag := override.GetTag()
	if tag == "" {
		if i := strings.LastIndex(def, ":"); i > 0 {
			tag = def[i+1:]
		}
	}
	return override.GetRepo() + ":" + tag
}

// apiEndpoint is the in-cluster URL clients (external-secrets stores,
// cert-manager Vault issuers) point at.
func (l *Locals) apiEndpoint() string {
	return fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", l.Scheme, l.ReleaseName, l.Namespace, vars.ApiPort)
}
