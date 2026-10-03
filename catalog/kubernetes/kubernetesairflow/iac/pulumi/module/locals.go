package module

import (
	"fmt"
	"strconv"
	"strings"

	kubernetesairflowv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesairflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesairflowv1alpha1.KubernetesAirflowSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the module-owned Secrets — never injected into
	// the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace Airflow installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name:
	// several Airflow installations can coexist in one Kubernetes
	// cluster. At the chart's default useStandardNaming=false the
	// fullname IS the release name, so every child is `<name>-<suffix>`.
	ReleaseName string

	// Chart and Airflow versions resolved to the pinned defaults when
	// unset, so both engines install the same chart whether or not the
	// platform's defaulting middleware ran.
	ChartVersion   string
	AirflowVersion string

	// Executor resolved to the KubernetesExecutor default. CeleryEnabled
	// is the chart's own pairing test (substring match on the two
	// Celery family names) — it gates the broker, result-backend and
	// worker surfaces.
	Executor      string
	CeleryEnabled bool

	// The metadata database resolved to literals: driver protocol
	// ("postgresql" or "mysql" — the chart's urlJoin scheme values),
	// endpoint, database, user, sslmode (postgres only), and the
	// referenced password Secret.
	DbProtocol          string
	DbHost              string
	DbPort              int
	DbName              string
	DbUser              string
	DbSslMode           string
	DbPasswordSecret    string
	DbPasswordSecretKey string

	// PgBouncer routing (mirrors the chart's own connection rewriting
	// EXACTLY): when enabled, Airflow's connection URIs point at the
	// pooler Service on the pooler port, and the DATABASE segment
	// becomes the pgbouncer.ini alias (`<release>-metadata` /
	// `<release>-result-backend`) while pgbouncer.ini maps the alias
	// back to the real host/database.
	PgbouncerEnabled bool
	// Host/port/database the AIRFLOW COMPONENTS dial (pooler when
	// enabled, the real database otherwise).
	EffectiveDbHost           string
	EffectiveDbPort           int
	EffectiveMetadataDbName   string
	EffectiveResultBackendDb  string
	PgbouncerConfigSecretName string
	PgbouncerStatsSecretName  string

	// Broker arm resolution (Celery only). BundledRedis and the
	// composed arm both end in a module-owned broker-url Secret; the
	// existing-secret arm passes the user's Secret name through.
	BundledRedisEnabled        bool
	BrokerHost                 string
	BrokerPort                 int
	BrokerUser                 string
	BrokerDb                   int
	BrokerPasswordSecret       string
	BrokerPasswordSecretKey    string
	BrokerUrlSecretName        string
	BrokerUrlSecretModuleOwned bool
	RedisPasswordSecretName    string

	// Log read-path backend ("" when none): "elasticsearch" or
	// "opensearch", plus the connection pieces the module composes
	// into the log-read connection Secret.
	LogBackend                  string
	LogBackendScheme            string
	LogBackendHost              string
	LogBackendPort              int
	LogBackendUser              string
	LogBackendPasswordSecret    string
	LogBackendPasswordSecretKey string
	LogReadConnSecretName       string

	// Module-owned Secret names (BYO names respected where the spec
	// allows them).
	MetadataConnSecretName      string
	ResultBackendConnSecretName string
	FernetKeySecretName         string
	FernetKeyModuleOwned        bool
	ApiSecretKeySecretName      string
	ApiSecretKeyModuleOwned     bool
	WebserverSecretKeyName      string
	JwtSecretName               string
	JwtSecretModuleOwned        bool

	// Admin bootstrap user.
	AdminCreate            bool
	AdminUsername          string
	AdminEmail             string
	AdminSecretName        string
	AdminSecretKey         string
	AdminSecretModuleOwned bool

	// Output handles.
	ApiServerServiceName string
	ApiServerEndpoint    string
	PortForwardCommand   string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesairflowv1alpha1.KubernetesAirflowIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesAirflow.String(),
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

	locals.ChartVersion = spec.GetChartVersion()
	if locals.ChartVersion == "" {
		locals.ChartVersion = vars.DefaultChartVersion
	}
	locals.AirflowVersion = spec.GetAirflowVersion()
	if locals.AirflowVersion == "" {
		locals.AirflowVersion = vars.DefaultAirflowVersion
	}
	locals.Executor = spec.GetExecutor()
	if locals.Executor == "" {
		locals.Executor = vars.DefaultExecutor
	}
	// The chart's own pairing test (check-values.yaml): a Celery family
	// member anywhere in the (possibly comma-separated) executor list.
	locals.CeleryEnabled = containsAny(locals.Executor, "CeleryExecutor", "CeleryKubernetesExecutor")

	// ------------------------------ database ------------------------------
	if pg := spec.GetDatabase().GetPostgres(); pg != nil {
		locals.DbProtocol = "postgresql"
		locals.DbHost = pg.GetHost().GetValue()
		locals.DbPort = int(pg.GetPort())
		if locals.DbPort == 0 {
			locals.DbPort = 5432
		}
		locals.DbName = pg.GetDatabaseName()
		if locals.DbName == "" {
			locals.DbName = "airflow"
		}
		locals.DbUser = pg.GetUsername()
		if locals.DbUser == "" {
			locals.DbUser = "airflow"
		}
		locals.DbSslMode = pg.GetSslMode()
		if locals.DbSslMode == "" {
			locals.DbSslMode = "disable"
		}
		locals.DbPasswordSecret = pg.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = pg.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	} else if my := spec.GetDatabase().GetMysql(); my != nil {
		locals.DbProtocol = "mysql"
		locals.DbHost = my.GetHost().GetValue()
		locals.DbPort = int(my.GetPort())
		if locals.DbPort == 0 {
			locals.DbPort = 3306
		}
		locals.DbName = my.GetDatabaseName()
		if locals.DbName == "" {
			locals.DbName = "airflow"
		}
		locals.DbUser = my.GetUsername()
		if locals.DbUser == "" {
			locals.DbUser = "airflow"
		}
		locals.DbPasswordSecret = my.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = my.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	}

	// ------------------------------ pgbouncer -----------------------------
	// Mirrors metadata-connection-secret.yaml at the pin: host swaps to
	// `<fullname>-pgbouncer`, port to ports.pgbouncer, and the database
	// segment to the pgbouncer.ini [databases] alias.
	locals.PgbouncerEnabled = spec.GetPgbouncer().GetEnabled()
	locals.EffectiveDbHost = locals.DbHost
	locals.EffectiveDbPort = locals.DbPort
	locals.EffectiveMetadataDbName = locals.DbName
	locals.EffectiveResultBackendDb = locals.DbName
	if locals.PgbouncerEnabled {
		locals.EffectiveDbHost = fmt.Sprintf("%s-pgbouncer", locals.ReleaseName)
		locals.EffectiveDbPort = vars.PgbouncerPort
		locals.EffectiveMetadataDbName = fmt.Sprintf("%s-metadata", locals.ReleaseName)
		locals.EffectiveResultBackendDb = fmt.Sprintf("%s-result-backend", locals.ReleaseName)
		locals.PgbouncerConfigSecretName = locals.ReleaseName + vars.PgbouncerConfigSecretSuffix
		locals.PgbouncerStatsSecretName = locals.ReleaseName + vars.PgbouncerStatsSecretSuffix
	}

	// ------------------------------- broker -------------------------------
	if locals.CeleryEnabled {
		broker := spec.GetBroker()
		switch {
		case broker.GetBundledRedis() != nil:
			locals.BundledRedisEnabled = true
			locals.BrokerHost = fmt.Sprintf("%s-redis", locals.ReleaseName)
			locals.BrokerPort = vars.RedisPort
			locals.BrokerDb = 0
			locals.RedisPasswordSecretName = locals.ReleaseName + vars.RedisPasswordSecretSuffix
			locals.BrokerUrlSecretName = locals.ReleaseName + vars.BrokerUrlSecretSuffix
			locals.BrokerUrlSecretModuleOwned = true
		case broker.GetValkey() != nil:
			vk := broker.GetValkey()
			locals.BrokerHost = vk.GetHost().GetValue()
			locals.BrokerPort = int(vk.GetPort())
			if locals.BrokerPort == 0 {
				locals.BrokerPort = vars.RedisPort
			}
			locals.BrokerUser = vk.GetUsername()
			locals.BrokerDb = int(vk.GetDatabaseNumber())
			if vk.GetPasswordSecret() != nil {
				locals.BrokerPasswordSecret = vk.GetPasswordSecret().GetSecretName().GetValue()
				locals.BrokerPasswordSecretKey = vk.GetPasswordSecret().GetSecretKey()
				if locals.BrokerPasswordSecretKey == "" {
					locals.BrokerPasswordSecretKey = "password"
				}
			}
			locals.BrokerUrlSecretName = locals.ReleaseName + vars.BrokerUrlSecretSuffix
			locals.BrokerUrlSecretModuleOwned = true
		case broker.GetExistingBrokerUrlSecret() != nil:
			locals.BrokerUrlSecretName = broker.GetExistingBrokerUrlSecret().GetSecretName()
		}
	}

	// ----------------------------- log read path --------------------------
	logBackend := (*kubernetesairflowv1alpha1.KubernetesAirflowLogSearchBackend)(nil)
	if es := spec.GetLogging().GetElasticsearch(); es != nil {
		locals.LogBackend = "elasticsearch"
		logBackend = es
	} else if os := spec.GetLogging().GetOpensearch(); os != nil {
		locals.LogBackend = "opensearch"
		logBackend = os
	}
	if logBackend != nil {
		locals.LogBackendScheme = logBackend.GetScheme()
		if locals.LogBackendScheme == "" {
			locals.LogBackendScheme = "http"
		}
		locals.LogBackendHost = logBackend.GetHost().GetValue()
		locals.LogBackendPort = int(logBackend.GetPort())
		if locals.LogBackendPort == 0 {
			locals.LogBackendPort = 9200
		}
		locals.LogBackendUser = logBackend.GetUsername()
		if logBackend.GetPasswordSecret() != nil {
			locals.LogBackendPasswordSecret = logBackend.GetPasswordSecret().GetSecretName().GetValue()
			locals.LogBackendPasswordSecretKey = logBackend.GetPasswordSecret().GetSecretKey()
			if locals.LogBackendPasswordSecretKey == "" {
				locals.LogBackendPasswordSecretKey = "password"
			}
		}
		locals.LogReadConnSecretName = locals.ReleaseName + vars.LogReadConnSecretSuffix
	}

	// ----------------------------- key secrets ----------------------------
	locals.MetadataConnSecretName = locals.ReleaseName + vars.MetadataConnSecretSuffix
	locals.ResultBackendConnSecretName = locals.ReleaseName + vars.ResultBackendConnSecretSuffix

	security := spec.GetSecurity()
	locals.FernetKeySecretName = security.GetFernetKeySecretName()
	locals.FernetKeyModuleOwned = locals.FernetKeySecretName == ""
	if locals.FernetKeyModuleOwned {
		locals.FernetKeySecretName = locals.ReleaseName + vars.FernetKeySecretSuffix
	}
	locals.ApiSecretKeySecretName = security.GetApiSecretKeySecretName()
	locals.ApiSecretKeyModuleOwned = locals.ApiSecretKeySecretName == ""
	if locals.ApiSecretKeyModuleOwned {
		locals.ApiSecretKeySecretName = locals.ReleaseName + vars.ApiSecretKeySecretSuffix
	}
	// The FAB webserver session key has no BYO field: always
	// module-owned (the chart would regenerate it every upgrade render,
	// logging out every UI session).
	locals.WebserverSecretKeyName = locals.ReleaseName + vars.WebserverSecretKeySuffix
	locals.JwtSecretName = security.GetJwtSecretName()
	locals.JwtSecretModuleOwned = locals.JwtSecretName == ""
	if locals.JwtSecretModuleOwned {
		locals.JwtSecretName = locals.ReleaseName + vars.JwtSecretSuffix
	}

	// ------------------------------ admin user ----------------------------
	admin := spec.GetAdminUser()
	locals.AdminCreate = true
	if admin != nil && admin.Create != nil {
		locals.AdminCreate = admin.GetCreate()
	}
	locals.AdminUsername = admin.GetUsername()
	if locals.AdminUsername == "" {
		locals.AdminUsername = "admin"
	}
	locals.AdminEmail = admin.GetEmail()
	if locals.AdminEmail == "" {
		locals.AdminEmail = "admin@example.com"
	}
	if admin.GetPasswordSecret() != nil {
		locals.AdminSecretName = admin.GetPasswordSecret().GetSecretName()
		locals.AdminSecretKey = admin.GetPasswordSecret().GetSecretKey()
		if locals.AdminSecretKey == "" {
			locals.AdminSecretKey = "password"
		}
	} else {
		locals.AdminSecretName = locals.ReleaseName + vars.AdminAuthSecretSuffix
		locals.AdminSecretKey = vars.AdminPasswordKey
		locals.AdminSecretModuleOwned = true
	}

	// ------------------------------- outputs ------------------------------
	locals.ApiServerServiceName = fmt.Sprintf("%s-api-server", locals.ReleaseName)
	locals.ApiServerEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.ApiServerServiceName, locals.Namespace, vars.ApiServerPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		locals.ApiServerServiceName, locals.Namespace, vars.ApiServerPort, vars.ApiServerPort)

	return locals
}

// containsAny reports whether s contains any of the given substrings —
// the chart's own Celery detection shape (Go template `contains`).
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
