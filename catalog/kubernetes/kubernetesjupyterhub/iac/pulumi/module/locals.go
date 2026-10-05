package module

import (
	"fmt"
	"strconv"

	kubernetesjupyterhubv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesjupyterhub/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesjupyterhubv1alpha1.KubernetesJupyterHubSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the module-owned Secrets — never injected into
	// the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace JupyterHub installs into (resolved literal from the
	// spec's value-or-ref). The chart's resource names are FIXED (hub,
	// proxy, proxy-public…), so the namespace is the multi-instance
	// boundary.
	Namespace string

	// Helm release name — metadata.name. With the chart's default
	// fullnameOverride "" the release name never prefixes resource
	// names; it only names the release record and the module-owned
	// Secrets.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Hub database arm resolution. DbType is the chart's own
	// hub.db.type vocabulary: "sqlite-pvc" (default), "postgres",
	// "mysql". On the external arms DbUrl is the CREDENTIAL-FREE
	// SQLAlchemy URL (the hub reads the password separately — see
	// HubSecret*) and the referenced password Secret is read at apply
	// time.
	DbType              string
	DbUrl               string
	DbPasswordSecret    string
	DbPasswordSecretKey string

	// Module-owned hub Secret (`<name>-hub-secret`) — mounted by the
	// hub through the chart's hub.existingSecret seam; carries the
	// external database password under the chart's contract key
	// `hub.db.password`. Only exists on the external database arms.
	HubSecretName    string
	HubSecretEnabled bool

	// Authentication arm resolution. AuthMethod: "shared_password"
	// (the secured default), "native", "github", "google", "oidc".
	// On the shared-password arm the module generates (or passes
	// through) the password Secret; on OAuth arms the client secret
	// Secret is wired through an env var.
	AuthMethod                  string
	SharedPasswordSecretName    string
	SharedPasswordSecretKey     string
	SharedPasswordModuleOwned   bool
	OauthClientSecretSecretName string
	OauthClientSecretSecretKey  string

	// Output handles.
	ProxyPublicEndpoint string
	PortForwardCommand  string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesjupyterhubv1alpha1.KubernetesJupyterHubIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesJupyterHub.String(),
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

	// ------------------------------ database ------------------------------
	// The chart's own vocabulary: sqlite-pvc is the default arm; the
	// external arms carry a CREDENTIAL-FREE url (the hub composes the
	// password from the mounted existing-secret at startup — chart
	// truth: jupyterhub_config.py exports PGPASSWORD/MYSQL_PWD from
	// the `hub.db.password` secret key).
	locals.DbType = "sqlite-pvc"
	if pg := spec.GetHub().GetDatabase().GetPostgres(); pg != nil {
		locals.DbType = "postgres"
		port := int(pg.GetPort())
		if port == 0 {
			port = 5432
		}
		databaseName := pg.GetDatabaseName()
		if databaseName == "" {
			databaseName = "jupyterhub"
		}
		username := pg.GetUsername()
		if username == "" {
			username = "jupyterhub"
		}
		locals.DbUrl = fmt.Sprintf("postgresql+psycopg2://%s@%s:%d/%s",
			username, pg.GetHost().GetValue(), port, databaseName)
		locals.DbPasswordSecret = pg.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = pg.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	} else if my := spec.GetHub().GetDatabase().GetMysql(); my != nil {
		locals.DbType = "mysql"
		port := int(my.GetPort())
		if port == 0 {
			port = 3306
		}
		databaseName := my.GetDatabaseName()
		if databaseName == "" {
			databaseName = "jupyterhub"
		}
		username := my.GetUsername()
		if username == "" {
			username = "jupyterhub"
		}
		locals.DbUrl = fmt.Sprintf("mysql+pymysql://%s@%s:%d/%s",
			username, my.GetHost().GetValue(), port, databaseName)
		locals.DbPasswordSecret = my.GetPasswordSecret().GetSecretName().GetValue()
		locals.DbPasswordSecretKey = my.GetPasswordSecret().GetSecretKey()
		if locals.DbPasswordSecretKey == "" {
			locals.DbPasswordSecretKey = "password"
		}
	}
	locals.HubSecretEnabled = locals.DbType != "sqlite-pvc"
	if locals.HubSecretEnabled {
		locals.HubSecretName = locals.ReleaseName + vars.HubSecretSuffix
	}

	// ---------------------------- authentication --------------------------
	// The secured default: an ABSENT authentication block (or an empty
	// oneof) is the shared-password arm with a module-generated
	// password — the chart's own default (any username, NO password)
	// never ships.
	auth := spec.GetAuthentication()
	locals.AuthMethod = "shared_password"
	switch {
	case auth.GetNative() != nil:
		locals.AuthMethod = "native"
	case auth.GetGithub() != nil:
		locals.AuthMethod = "github"
		locals.OauthClientSecretSecretName = auth.GetGithub().GetClientSecretSecret().GetSecretName()
		locals.OauthClientSecretSecretKey = auth.GetGithub().GetClientSecretSecret().GetSecretKey()
	case auth.GetGoogle() != nil:
		locals.AuthMethod = "google"
		locals.OauthClientSecretSecretName = auth.GetGoogle().GetClientSecretSecret().GetSecretName()
		locals.OauthClientSecretSecretKey = auth.GetGoogle().GetClientSecretSecret().GetSecretKey()
	case auth.GetOidc() != nil:
		locals.AuthMethod = "oidc"
		locals.OauthClientSecretSecretName = auth.GetOidc().GetClientSecretSecret().GetSecretName()
		locals.OauthClientSecretSecretKey = auth.GetOidc().GetClientSecretSecret().GetSecretKey()
	}
	if locals.OauthClientSecretSecretKey == "" {
		locals.OauthClientSecretSecretKey = "password"
	}

	if locals.AuthMethod == "shared_password" {
		if byo := auth.GetSharedPassword().GetPasswordSecret(); byo != nil {
			locals.SharedPasswordSecretName = byo.GetSecretName()
			locals.SharedPasswordSecretKey = byo.GetSecretKey()
			if locals.SharedPasswordSecretKey == "" {
				locals.SharedPasswordSecretKey = "password"
			}
		} else {
			locals.SharedPasswordSecretName = locals.ReleaseName + vars.AuthSecretSuffix
			locals.SharedPasswordSecretKey = vars.SharedPasswordKey
			locals.SharedPasswordModuleOwned = true
		}
	}

	// ------------------------------- outputs ------------------------------
	locals.ProxyPublicEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		vars.ProxyPublicServiceName, locals.Namespace, vars.ProxyPublicPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s 8080:%d",
		vars.ProxyPublicServiceName, locals.Namespace, vars.ProxyPublicPort)

	return locals
}
