package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetestrinov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestrino/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// catalogEnvRef is one environment variable sourced from an existing
// Secret — the delivery vehicle for `${ENV:VAR}` references in rendered
// catalog/config properties.
type catalogEnvRef struct {
	EnvName    string
	SecretName string
	SecretKey  string
}

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetestrinov1alpha1.KubernetesTrinoSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the module-owned Secrets — never injected into
	// the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace Trino installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name = metadata.name. The module PINS
	// fullnameOverride to the same value, so child names are
	// deterministic: `<name>-coordinator`, `<name>-worker`,
	// `<name>-catalog`, `<name>-schemas-volume-coordinator`.
	ReleaseName string

	// Authentication resolution. AuthEnabled is the secured default
	// (true unless explicitly disabled). ModuleOwnedPasswordDb is true
	// when the module generates the admin credential + password file;
	// false on the bring-your-own arm.
	AuthEnabled           bool
	AdminUsername         string
	ModuleOwnedPasswordDb bool
	PasswordDbSecretName  string
	GroupsSecretName      string

	// The internal-communication shared secret Secret (module-owned,
	// only when auth is enabled).
	InternalSecretName string

	// Environment variables delivered to every Trino container from
	// existing Secrets: the per-catalog password refs plus the user's
	// extra_env_from_secret entries. The shared-secret ref is appended
	// separately (its Secret is module-owned).
	SecretEnvRefs []catalogEnvRef

	// Output handles.
	CoordinatorService  string
	WorkerService       string
	CoordinatorEndpoint string
	PortForwardCommand  string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetestrinov1alpha1.KubernetesTrinoIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesTrino.String(),
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

	locals := &Locals{
		Spec:        spec,
		Labels:      labels,
		Namespace:   spec.Namespace.GetValue(),
		ReleaseName: target.Metadata.Name,
	}

	// --------------------------- authentication ---------------------------
	// The secured default: an ABSENT auth block means PASSWORD auth
	// with a module-generated admin — the chart's own default (no
	// authentication at all) never ships.
	auth := spec.GetAuth()
	locals.AuthEnabled = true
	if auth != nil && auth.Enabled != nil && !auth.GetEnabled() {
		locals.AuthEnabled = false
	}
	if locals.AuthEnabled {
		locals.AdminUsername = auth.GetAdminUsername()
		if locals.AdminUsername == "" {
			locals.AdminUsername = "trino"
		}
		if byo := auth.GetExistingPasswordDbSecret(); byo != nil {
			locals.PasswordDbSecretName = byo.GetSecretName()
			// Bring-your-own file: no module-generated admin exists,
			// so the exported credential handles stay EMPTY (honest —
			// the module cannot know the file's users).
			locals.AdminUsername = ""
		} else {
			locals.PasswordDbSecretName = locals.ReleaseName + vars.AuthSecretSuffix
			locals.ModuleOwnedPasswordDb = true
		}
		locals.GroupsSecretName = auth.GetGroupsSecret().GetSecretName()
		locals.InternalSecretName = locals.ReleaseName + vars.InternalSecretSuffix
	}

	// ------------------- secret-sourced environment refs ------------------
	// Per-catalog password env vars (referenced from the rendered
	// catalog properties via ${ENV:...}) plus the user's own entries.
	for _, catalog := range spec.GetCatalogs().GetPostgres() {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, catalogEnvRef{
			EnvName:    catalogPasswordEnvVar(catalog.GetName()),
			SecretName: catalog.GetPasswordSecret().GetSecretName().GetValue(),
			SecretKey:  defaultString(catalog.GetPasswordSecret().GetSecretKey(), "password"),
		})
	}
	for _, catalog := range spec.GetCatalogs().GetMysql() {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, catalogEnvRef{
			EnvName:    catalogPasswordEnvVar(catalog.GetName()),
			SecretName: catalog.GetPasswordSecret().GetSecretName().GetValue(),
			SecretKey:  defaultString(catalog.GetPasswordSecret().GetSecretKey(), "password"),
		})
	}
	for envName, ref := range spec.GetExtraEnvFromSecret() {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, catalogEnvRef{
			EnvName:    envName,
			SecretName: ref.GetSecretName(),
			SecretKey:  ref.GetSecretKey(),
		})
	}

	// ------------------------------- outputs ------------------------------
	locals.CoordinatorService = locals.ReleaseName + vars.CoordinatorSuffix
	locals.WorkerService = locals.ReleaseName + vars.WorkerSuffix
	locals.CoordinatorEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.CoordinatorService, locals.Namespace, vars.HttpPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s 8080:%d",
		locals.CoordinatorService, locals.Namespace, vars.HttpPort)

	return locals
}

// catalogPasswordEnvVar derives the env-var name a catalog's rendered
// properties reference — catalog names are [a-z][a-z0-9_]* by CEL, so the
// upper-cased form is env-var safe.
func catalogPasswordEnvVar(catalogName string) string {
	return vars.CatalogPasswordEnvPrefix + strings.ToUpper(catalogName) + "_PASSWORD"
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
