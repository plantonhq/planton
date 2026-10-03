package module

import (
	"fmt"
	"strconv"

	kubernetessupersetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessuperset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// secretEnvRef is one environment variable delivered as a secretKeyRef —
// the chart's extraEnvRaw shape (explicit env beats the envFrom Secret).
type secretEnvRef struct {
	EnvName    string
	SecretName string
	SecretKey  string
}

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetessupersetv1alpha1.KubernetesSupersetSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the module-owned Secrets — never injected into
	// the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace Superset installs into.
	Namespace string

	// Helm release name = metadata.name. The module PINS
	// fullnameOverride to the same value, so child names are
	// deterministic: `<name>` (web), `<name>-worker`,
	// `<name>-celerybeat`, `<name>-init-db`, and the `<name>-env` /
	// `<name>-config` Secrets the chart consumes by name.
	ReleaseName string

	// The module-owned environment Secret (`<name>-env`) — the chart's
	// runtime-credential contract (every component envFroms it; the
	// chart's own copy is turned OFF via secretEnv.create=false). It
	// carries the NON-SECRET connection facts plus module-GENERATED
	// material; REFERENCED material (the database/cache passwords, any
	// bring-your-own keys) arrives as extraEnvRaw secretKeyRef entries
	// — never copied, never read at apply time.
	EnvSecretName string
	// Plain (non-secret) env keys the module composes into the env
	// Secret: DB_* connection facts, REDIS_* facts, admin identity.
	EnvPlain map[string]string

	// Whether the cache/broker is declared (drives Celery components
	// and the config's Redis blocks).
	CacheEnabled bool
	// The cache password reference, when the store is authed.
	CachePasswordSecret string
	CachePasswordKey    string

	// The metadata-database password reference (always external).
	DbPasswordSecret string
	DbPasswordKey    string

	// SECRET_KEY resolution: module-generated into `<name>-secret-key`
	// (the exported handle) or a bring-your-own reference.
	SecretKeyModuleOwned  bool
	SecretKeySecretName   string
	SecretKeySecretKeyKey string

	// Admin bootstrap resolution: module-generated into
	// `<name>-admin-auth` or a bring-your-own reference. The password
	// reaches the init step as the ADMIN_PASSWORD env var — the
	// chart's literal-rendering path (createAdmin) is never used.
	AdminUsername       string
	AdminEmail          string
	AdminModuleOwned    bool
	AdminPasswordSecret string
	AdminPasswordKey    string

	// Component toggles resolved once (worker default: on when the
	// cache is declared).
	WorkerEnabled     bool
	BeatEnabled       bool
	FlowerEnabled     bool
	WebsocketsEnabled bool
	McpEnabled        bool

	// extraEnvRaw entries: the referenced credentials + the user's own.
	SecretEnvRefs []secretEnvRef

	// Output handles.
	Service            string
	Endpoint           string
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetessupersetv1alpha1.KubernetesSupersetIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSuperset.String(),
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
	locals.EnvSecretName = locals.ReleaseName + vars.EnvSecretSuffix

	// --------------------------- metadata database ------------------------
	db := spec.GetMetadataDatabase()
	dbPort := 5432
	if db.Port != nil {
		dbPort = int(db.GetPort())
	}
	locals.EnvPlain = map[string]string{
		"DB_HOST": db.GetHost().GetValue(),
		"DB_PORT": strconv.Itoa(dbPort),
		"DB_USER": defaultString(db.GetUsername(), "superset"),
		"DB_NAME": defaultString(db.GetDatabaseName(), "superset"),
	}
	locals.DbPasswordSecret = db.GetPasswordSecret().GetSecretName().GetValue()
	locals.DbPasswordKey = defaultString(db.GetPasswordSecret().GetSecretKey(), "password")

	// -------------------------------- cache -------------------------------
	if cache := spec.GetCache(); cache != nil {
		locals.CacheEnabled = true
		cachePort := 6379
		if cache.Port != nil {
			cachePort = int(cache.GetPort())
		}
		locals.EnvPlain["REDIS_HOST"] = cache.GetHost().GetValue()
		locals.EnvPlain["REDIS_PORT"] = strconv.Itoa(cachePort)
		if cache.GetUsername() != "" {
			locals.EnvPlain["REDIS_USER"] = cache.GetUsername()
		}
		if secret := cache.GetPasswordSecret(); secret != nil {
			locals.CachePasswordSecret = secret.GetSecretName().GetValue()
			locals.CachePasswordKey = defaultString(secret.GetSecretKey(), "password")
		}
	}

	// ------------------------------ SECRET_KEY -----------------------------
	if byo := spec.GetSecretKeySecret(); byo != nil {
		locals.SecretKeySecretName = byo.GetSecretName()
		locals.SecretKeySecretKeyKey = byo.GetSecretKey()
	} else {
		locals.SecretKeyModuleOwned = true
		locals.SecretKeySecretName = locals.ReleaseName + vars.SecretKeySuffix
		locals.SecretKeySecretKeyKey = vars.SecretKeyKey
	}

	// ------------------------------ admin user -----------------------------
	admin := spec.GetInit().GetAdmin()
	locals.AdminUsername = defaultString(admin.GetUsername(), "admin")
	locals.AdminEmail = defaultString(admin.GetEmail(), "admin@superset.local")
	locals.EnvPlain["ADMIN_USERNAME"] = locals.AdminUsername
	locals.EnvPlain["ADMIN_EMAIL"] = locals.AdminEmail
	if byo := admin.GetPasswordSecret(); byo != nil {
		locals.AdminPasswordSecret = byo.GetSecretName()
		locals.AdminPasswordKey = byo.GetSecretKey()
	} else {
		locals.AdminModuleOwned = true
		locals.AdminPasswordSecret = locals.ReleaseName + vars.AdminAuthSuffix
		locals.AdminPasswordKey = vars.AdminPasswordKey
	}

	// ------------------------- component toggles --------------------------
	worker := spec.GetWorker()
	locals.WorkerEnabled = locals.CacheEnabled
	if worker != nil && worker.Enabled != nil {
		locals.WorkerEnabled = worker.GetEnabled() && locals.CacheEnabled
	}
	locals.BeatEnabled = spec.GetBeat().GetEnabled()
	locals.FlowerEnabled = spec.GetFlower().GetEnabled()
	locals.WebsocketsEnabled = spec.GetWebsockets().GetEnabled()
	locals.McpEnabled = spec.GetMcp().GetEnabled()

	// -------------------- referenced-credential env refs ------------------
	// Explicit env entries (extraEnvRaw) override the envFrom Secret —
	// the chart's own bring-your-own-credential mechanism. Secret NAMES
	// only ever render.
	locals.SecretEnvRefs = append(locals.SecretEnvRefs, secretEnvRef{
		EnvName:    "DB_PASS",
		SecretName: locals.DbPasswordSecret,
		SecretKey:  locals.DbPasswordKey,
	})
	if locals.CachePasswordSecret != "" {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, secretEnvRef{
			EnvName:    "REDIS_PASSWORD",
			SecretName: locals.CachePasswordSecret,
			SecretKey:  locals.CachePasswordKey,
		})
	}
	if !locals.SecretKeyModuleOwned {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, secretEnvRef{
			EnvName:    "SUPERSET_SECRET_KEY",
			SecretName: locals.SecretKeySecretName,
			SecretKey:  locals.SecretKeySecretKeyKey,
		})
	}
	if !locals.AdminModuleOwned {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, secretEnvRef{
			EnvName:    "ADMIN_PASSWORD",
			SecretName: locals.AdminPasswordSecret,
			SecretKey:  locals.AdminPasswordKey,
		})
	}
	for envName, ref := range spec.GetExtraEnvFromSecret() {
		locals.SecretEnvRefs = append(locals.SecretEnvRefs, secretEnvRef{
			EnvName:    envName,
			SecretName: ref.GetSecretName(),
			SecretKey:  ref.GetSecretKey(),
		})
	}

	// ------------------------------- outputs ------------------------------
	locals.Service = locals.ReleaseName
	locals.Endpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.Service, locals.Namespace, vars.HttpPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s 8088:%d",
		locals.Service, locals.Namespace, vars.HttpPort)

	return locals
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
