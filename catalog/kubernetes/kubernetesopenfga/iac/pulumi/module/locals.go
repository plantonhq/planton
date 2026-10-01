package module

import (
	"fmt"
	"strconv"

	kubernetesopenfgav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenfga/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the stack input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesopenfgav1alpha1.KubernetesOpenFgaSpec

	// Resource-identity labels: stamped on the module-created satellites
	// (the namespace and the authn-keys Secret) and on every OpenFGA pod
	// through the chart's own podExtraLabels -- so a log line, a metric or
	// an alert from it names its organization and environment. The chart's
	// selector is its own fixed labels; these never reach it.
	Labels map[string]string

	// Namespace the server installs into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// ReleaseName is metadata.name — several OpenFGA servers coexist in
	// one cluster. fullnameOverride pins the chart fullname to it, so
	// every chart child (Service, ServiceAccount, `-migrate` Job,
	// `-datastore-secret`) derives deterministically.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// ServiceName is the chart's ClusterIP Service — openfga.fullname,
	// pinned to the resource name via fullnameOverride.
	ServiceName string

	// Datastore rendering, resolved once (twin: locals.tf). The URI
	// NEVER carries userinfo: credentials ride separate env vars (the
	// server gives flag-supplied credentials precedence over
	// URI-embedded ones), so nothing credential-bearing lands in
	// rendered values.
	DatastoreEngine    string
	DatastoreUri       string // "" on the memory arm
	DatastoreUsername  string // "" on the memory arm
	PasswordSecretName string // existing Secret carrying the password; "" on memory
	PasswordSecretKey  string // key within that Secret (default "password")

	// AuthnKeysSecretName is the module-owned Secret materialized from
	// declared pre-shared keys (`<name>-authn-keys`); "" when authn is
	// unset, oidc, or rides an existing Secret. Exported as an output.
	AuthnKeysSecretName string

	// PresharedKeysSecretRef is the Secret NAME rendered into
	// authn.preshared.keysSecret — the module-owned Secret or the
	// user's existing one. "" when authn is not preshared.
	PresharedKeysSecretRef string

	// Composition handles (twin: outputs.tf).
	ApiHttpEndpoint    string
	ApiGrpcEndpoint    string
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, stackInput *kubernetesopenfgav1alpha1.KubernetesOpenFgaStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesOpenFga.String(),
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

	namespace := spec.Namespace.GetValue()
	releaseName := target.Metadata.Name
	serviceName := releaseName

	// ---- datastore arms ---------------------------------------------------
	// URIs are engine-native DSNs WITHOUT userinfo; username and password
	// reach the server as OPENFGA_DATASTORE_USERNAME / _PASSWORD env
	// vars, which the server prefers over any URI-embedded credentials.
	datastoreEngine := "memory"
	datastoreUri := ""
	datastoreUsername := ""
	passwordSecretName := ""
	passwordSecretKey := ""

	if pg := spec.GetDatastore().GetPostgres(); pg != nil {
		datastoreEngine = "postgres"
		port := int32(5432)
		if pg.Port != nil {
			port = pg.GetPort()
		}
		sslMode := pg.GetSslMode()
		if sslMode == "" {
			sslMode = "disable"
		}
		datastoreUri = fmt.Sprintf("postgres://%s:%d/%s?sslmode=%s",
			pg.GetHost().GetValue(), port, pg.GetDatabase(), sslMode)
		datastoreUsername = pg.GetUsername()
		passwordSecretName = pg.GetPasswordSecret().GetSecretName().GetValue()
		passwordSecretKey = pg.GetPasswordSecret().GetSecretKey()
	}

	if my := spec.GetDatastore().GetMysql(); my != nil {
		datastoreEngine = "mysql"
		port := int32(3306)
		if my.Port != nil {
			port = my.GetPort()
		}
		// Go MySQL DSN (host part only — no userinfo); parseTime=true is
		// required by the server's mysql storage adapter.
		datastoreUri = fmt.Sprintf("tcp(%s:%d)/%s?parseTime=true",
			my.GetHost().GetValue(), port, my.GetDatabase())
		datastoreUsername = my.GetUsername()
		passwordSecretName = my.GetPasswordSecret().GetSecretName().GetValue()
		passwordSecretKey = my.GetPasswordSecret().GetSecretKey()
	}

	if passwordSecretName != "" && passwordSecretKey == "" {
		passwordSecretKey = "password"
	}

	// ---- pre-shared authn keys ----------------------------------------------
	// Declared keys materialize into a module-owned Secret; an existing
	// Secret is referenced by name. Either way only a Secret NAME ever
	// renders into chart values.
	authnKeysSecretName := ""
	presharedKeysSecretRef := ""
	if preshared := spec.GetAuthn().GetPreshared(); preshared != nil {
		if len(preshared.GetKeys()) > 0 {
			authnKeysSecretName = releaseName + vars.AuthnKeysSecretSuffix
			presharedKeysSecretRef = authnKeysSecretName
		} else {
			presharedKeysSecretRef = preshared.GetExistingKeysSecretName()
		}
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              namespace,
		ReleaseName:            releaseName,
		ChartVersion:           chartVersion,
		ServiceName:            serviceName,
		DatastoreEngine:        datastoreEngine,
		DatastoreUri:           datastoreUri,
		DatastoreUsername:      datastoreUsername,
		PasswordSecretName:     passwordSecretName,
		PasswordSecretKey:      passwordSecretKey,
		AuthnKeysSecretName:    authnKeysSecretName,
		PresharedKeysSecretRef: presharedKeysSecretRef,
		ApiHttpEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
			serviceName, namespace, vars.HttpPort),
		ApiGrpcEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			serviceName, namespace, vars.GrpcPort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d",
			namespace, serviceName, vars.HttpPort, vars.HttpPort),
	}
}
