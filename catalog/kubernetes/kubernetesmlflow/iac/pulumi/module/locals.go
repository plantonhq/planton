package module

import (
	"fmt"
	"strconv"

	kubernetesmlflowv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesmlflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesmlflowv1alpha1.KubernetesMlflowSpec

	// Resource-identity labels stamped on every module-created object
	// (this is a module-owned-manifests kind — the module IS the
	// renderer, so unlike Helm kinds the labels reach everything).
	Labels map[string]string

	// Selector labels for the server Deployment/Service pairing —
	// immutable once deployed (a Deployment selector cannot change).
	SelectorLabels map[string]string

	// Namespace MLflow installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// metadata.name — the Deployment/Service name and the prefix for
	// every module-owned satellite.
	Name string

	// Server container resolution.
	Image    string
	Replicas int
	Workers  int

	// Backend-store arm resolution. BackendType: "sqlite" (default),
	// "postgres", "mysql". On the database arms the module composes
	// the SQLAlchemy URI at apply time from the referenced credential
	// Secret into `<name>-backend-uri`; on the sqlite arm the URI is a
	// literal file path on the data PVC.
	BackendType          string
	DbProtocol           string
	DbHost               string
	DbPort               int
	DbName               string
	DbUser               string
	DbPasswordSecret     string
	DbPasswordSecretKey  string
	BackendUriSecretName string
	SqliteBackendUri     string

	// Artifact-store arm resolution. ArtifactType: "pvc" (default),
	// "s3_compatible", "aws_s3", "gcs", "azure_blob". The server
	// PROXIES artifact traffic (--artifacts-destination), so clients
	// never carry store credentials.
	ArtifactType        string
	ArtifactDestination string

	// PVC needs derived from the arms (each drives a Recreate
	// deployment strategy — RWO volumes bind one pod).
	DataPvcEnabled      bool
	ArtifactsPvcEnabled bool

	// Authentication resolution. AuthEnabled defaults TRUE (the open
	// server never ships); the admin password is module-generated into
	// `<name>-admin-auth` unless a BYO Secret is referenced; the
	// basic-auth ini (composed into `<name>-auth-config`) points its
	// user database at the backend store.
	AuthEnabled            bool
	AdminUsername          string
	AdminSecretName        string
	AdminSecretKey         string
	AdminSecretModuleOwned bool
	AuthConfigSecretName   string
	DefaultPermission      string
	AuthDatabaseUriSqlite  string

	// Garbage collection.
	GcEnabled   bool
	GcSchedule  string
	GcOlderThan string

	// Metrics.
	MetricsEnabled        bool
	ServiceMonitorEnabled bool

	// Output handles.
	TrackingEndpoint   string
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesmlflowv1alpha1.KubernetesMlflowIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesMlflow.String(),
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
		Spec:      spec,
		Labels:    labels,
		Namespace: spec.Namespace.GetValue(),
		Name:      target.Metadata.Name,
		SelectorLabels: map[string]string{
			"app.kubernetes.io/name":     "mlflow",
			"app.kubernetes.io/instance": target.Metadata.Name,
		},
	}

	// ------------------------------ server --------------------------------
	server := spec.GetServer()
	repository := server.GetImage().GetRepository()
	if repository == "" {
		repository = vars.DefaultImageRepository
	}
	tag := server.GetImage().GetTag()
	if tag == "" {
		tag = vars.DefaultImageTag
	}
	locals.Image = fmt.Sprintf("%s:%s", repository, tag)
	locals.Replicas = 1
	if server != nil && server.Replicas != nil {
		locals.Replicas = int(server.GetReplicas())
	}
	locals.Workers = vars.DefaultWorkers
	if server != nil && server.Workers != nil {
		locals.Workers = int(server.GetWorkers())
	}

	// --------------------------- backend store ----------------------------
	locals.BackendType = "sqlite"
	if pg := spec.GetBackendStore().GetPostgres(); pg != nil {
		locals.BackendType = "postgres"
		locals.DbProtocol = "postgresql"
		locals.DbHost = pg.GetHost().GetValue()
		locals.DbPort = int(pg.GetPort())
		if locals.DbPort == 0 {
			locals.DbPort = 5432
		}
		locals.DbName = pg.GetDatabaseName()
		if locals.DbName == "" {
			locals.DbName = "mlflow"
		}
		locals.DbUser = pg.GetUsername()
		if locals.DbUser == "" {
			locals.DbUser = "mlflow"
		}
		locals.DbPasswordSecret = pg.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = pg.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	} else if my := spec.GetBackendStore().GetMysql(); my != nil {
		locals.BackendType = "mysql"
		locals.DbProtocol = "mysql+pymysql"
		locals.DbHost = my.GetHost().GetValue()
		locals.DbPort = int(my.GetPort())
		if locals.DbPort == 0 {
			locals.DbPort = 3306
		}
		locals.DbName = my.GetDatabaseName()
		if locals.DbName == "" {
			locals.DbName = "mlflow"
		}
		locals.DbUser = my.GetUsername()
		if locals.DbUser == "" {
			locals.DbUser = "mlflow"
		}
		locals.DbPasswordSecret = my.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = my.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	}
	if locals.BackendType == "sqlite" {
		locals.DataPvcEnabled = true
		// Four slashes: sqlite absolute-path URI form
		// (sqlite:/// + /mlflow/data/mlflow.db).
		locals.SqliteBackendUri = fmt.Sprintf("sqlite:///%s/mlflow.db", vars.DataMountPath)
		locals.AuthDatabaseUriSqlite = fmt.Sprintf("sqlite:///%s/basic_auth.db", vars.DataMountPath)
	} else {
		locals.BackendUriSecretName = locals.Name + vars.BackendUriSecretSuffix
	}

	// --------------------------- artifact store ---------------------------
	// The server proxies artifact traffic: --artifacts-destination is
	// the physical store; clients upload/download THROUGH the tracking
	// API and never need store credentials.
	locals.ArtifactType = "pvc"
	artifactStore := spec.GetArtifactStore()
	switch {
	case artifactStore.GetS3Compatible() != nil:
		locals.ArtifactType = "s3_compatible"
		s3 := artifactStore.GetS3Compatible()
		locals.ArtifactDestination = s3Uri(s3.GetBucket(), s3.GetPrefix())
	case artifactStore.GetAwsS3() != nil:
		locals.ArtifactType = "aws_s3"
		s3 := artifactStore.GetAwsS3()
		locals.ArtifactDestination = s3Uri(s3.GetBucket(), s3.GetPrefix())
	case artifactStore.GetGcs() != nil:
		locals.ArtifactType = "gcs"
		gcs := artifactStore.GetGcs()
		locals.ArtifactDestination = fmt.Sprintf("gs://%s", joinBucketPrefix(gcs.GetBucket(), gcs.GetPrefix()))
	case artifactStore.GetAzureBlob() != nil:
		locals.ArtifactType = "azure_blob"
		azure := artifactStore.GetAzureBlob()
		destination := fmt.Sprintf("wasbs://%s@%s.blob.core.windows.net",
			azure.GetContainer(), azure.GetStorageAccount())
		if azure.GetPrefix() != "" {
			destination = destination + "/" + azure.GetPrefix()
		}
		locals.ArtifactDestination = destination
	default:
		locals.ArtifactsPvcEnabled = true
		locals.ArtifactDestination = vars.ArtifactsMountPath
	}

	// ------------------------------- auth ----------------------------------
	auth := spec.GetAuth()
	locals.AuthEnabled = true
	if auth != nil && auth.Enabled != nil {
		locals.AuthEnabled = auth.GetEnabled()
	}
	if locals.AuthEnabled {
		locals.AdminUsername = auth.GetAdminUsername()
		if locals.AdminUsername == "" {
			locals.AdminUsername = "admin"
		}
		if byo := auth.GetAdminPasswordSecret(); byo != nil {
			locals.AdminSecretName = byo.GetSecretName()
			locals.AdminSecretKey = byo.GetSecretKey()
			if locals.AdminSecretKey == "" {
				locals.AdminSecretKey = "password"
			}
		} else {
			locals.AdminSecretName = locals.Name + vars.AdminAuthSecretSuffix
			locals.AdminSecretKey = vars.AdminPasswordKey
			locals.AdminSecretModuleOwned = true
		}
		locals.AuthConfigSecretName = locals.Name + vars.AuthConfigSecretSuffix
		locals.DefaultPermission = auth.GetDefaultPermission()
		if locals.DefaultPermission == "" {
			locals.DefaultPermission = "READ"
		}
	}

	// -------------------------------- gc -----------------------------------
	if gc := spec.GetGc(); gc.GetEnabled() {
		locals.GcEnabled = true
		locals.GcSchedule = gc.GetSchedule()
		if locals.GcSchedule == "" {
			locals.GcSchedule = "0 3 * * *"
		}
		locals.GcOlderThan = gc.GetOlderThan()
		if locals.GcOlderThan == "" {
			locals.GcOlderThan = "30d"
		}
	}

	// ------------------------------ metrics --------------------------------
	locals.MetricsEnabled = spec.GetMetrics().GetEnabled()
	locals.ServiceMonitorEnabled = spec.GetMetrics().GetServiceMonitorEnabled()

	// ------------------------------- outputs -------------------------------
	locals.TrackingEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.Name, locals.Namespace, vars.ServerPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		locals.Name, locals.Namespace, vars.ServerPort, vars.ServerPort)

	return locals
}

// s3Uri renders `s3://bucket[/prefix]`.
func s3Uri(bucket, prefix string) string {
	return "s3://" + joinBucketPrefix(bucket, prefix)
}

// joinBucketPrefix joins a bucket and optional key prefix.
func joinBucketPrefix(bucket, prefix string) string {
	if prefix == "" {
		return bucket
	}
	return bucket + "/" + prefix
}
