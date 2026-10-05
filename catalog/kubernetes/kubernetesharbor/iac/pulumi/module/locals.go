package module

import (
	"fmt"
	"strconv"

	kubernetesharborv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesharbor/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input. Every
// resolution here has an exact twin in the Terraform module's locals.tf
// — keep them in lockstep.
type Locals struct {
	Spec *kubernetesharborv1alpha1.KubernetesHarborSpec

	// Resource-identity labels stamped on module-created satellites
	// (namespace, credential Secrets) — never injected into the
	// chart's own resources; Helm owns those.
	Labels map[string]string

	// Namespace Harbor installs into (resolved literal).
	Namespace string

	// ReleaseName is metadata.name; fullnameOverride AND the
	// front-door Service name are pinned to it, so every chart-derived
	// name hangs off this value (the chart's default front-door name
	// is the literal "harbor" — two installs in one namespace would
	// collide without the pin).
	ReleaseName string

	ChartVersion string

	// The chart's expose.type value (clusterIP / nodePort /
	// loadBalancer). The chart's `ingress` and `route` types are
	// never rendered — exposure composes.
	ExposeType string

	// http/80 or https/443 at the nginx front door, following
	// expose.tls.enabled — drives the exported endpoint and
	// port-forward command.
	FrontDoorScheme string
	FrontDoorPort   int

	// The Secret holding the admin password: the module-owned
	// `<name>-admin-auth` on the generated arm, or the user's
	// declared Secret/key.
	AdminSecretName string
	AdminSecretKey  string
	// True when the module generates the admin password (admin_auth
	// unset or without existing_secret_name).
	AdminGenerated bool

	// Name of the module-owned Secret carrying every generated
	// inter-component credential (encryption key, core/jobservice/
	// registry HTTP secrets, registry basic-auth credential, CSRF
	// key) under the chart's per-site contract keys.
	InternalAuthSecretName string

	// Name of the module-owned Secret materializing a DECLARED
	// external-redis password ("" when the arm is not in play — the
	// user brought their own Secret, or the cache is internal /
	// unauthenticated).
	RedisAuthSecretName string

	// Name of the module-owned Secret materializing DECLARED
	// s3/gcs/azure storage credentials ("" when keyless/ambient or
	// user-provided).
	StorageAuthSecretName string

	// True when the Trivy scanner deploys (spec unset = chart truth:
	// enabled).
	TrivyEnabled bool

	// True on the internal database / internal redis arms.
	InternalDatabase bool
	InternalRedis    bool
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesharborv1alpha1.KubernetesHarborIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesHarbor.String(),
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

	// Map the spec's Kubernetes-conventional Service type onto the
	// chart's camel-case expose.type values.
	exposeType := "clusterIP"
	switch spec.GetExpose().GetServiceType() {
	case "NodePort":
		exposeType = "nodePort"
	case "LoadBalancer":
		exposeType = "loadBalancer"
	}

	frontDoorScheme := "http"
	frontDoorPort := 80
	if spec.GetExpose().GetTls().GetEnabled() {
		frontDoorScheme = "https"
		frontDoorPort = 443
	}

	adminSecretName := target.Metadata.Name + vars.AdminAuthSecretSuffix
	adminSecretKey := vars.AdminPasswordSecretKey
	adminGenerated := true
	if spec.GetAdminAuth().GetExistingSecretName() != "" {
		adminSecretName = spec.GetAdminAuth().GetExistingSecretName()
		adminGenerated = false
		if spec.GetAdminAuth().GetExistingSecretKey() != "" {
			adminSecretKey = spec.GetAdminAuth().GetExistingSecretKey()
		}
	}

	redisAuthSecretName := ""
	if ext := spec.GetCache().GetExternal(); ext != nil && ext.GetPassword() != "" {
		redisAuthSecretName = target.Metadata.Name + vars.RedisAuthSecretSuffix
	}

	storageAuthSecretName := ""
	if s3 := spec.GetStorage().GetS3(); s3 != nil && s3.GetCredentials().GetAccessKey() != "" {
		storageAuthSecretName = target.Metadata.Name + vars.StorageAuthSecretSuffix
	}
	if gcs := spec.GetStorage().GetGcs(); gcs != nil && gcs.GetKeyData() != "" {
		storageAuthSecretName = target.Metadata.Name + vars.StorageAuthSecretSuffix
	}
	if azure := spec.GetStorage().GetAzure(); azure != nil && azure.GetAccountKey() != "" {
		storageAuthSecretName = target.Metadata.Name + vars.StorageAuthSecretSuffix
	}

	// Trivy: unset block = chart truth (enabled).
	trivyEnabled := true
	if spec.GetTrivy() != nil && spec.GetTrivy().Enabled != nil {
		trivyEnabled = spec.GetTrivy().GetEnabled()
	}

	return &Locals{
		Spec:                   spec,
		Labels:                 labels,
		Namespace:              spec.Namespace.GetValue(),
		ReleaseName:            target.Metadata.Name,
		ChartVersion:           chartVersion,
		ExposeType:             exposeType,
		FrontDoorScheme:        frontDoorScheme,
		FrontDoorPort:          frontDoorPort,
		AdminSecretName:        adminSecretName,
		AdminSecretKey:         adminSecretKey,
		AdminGenerated:         adminGenerated,
		InternalAuthSecretName: target.Metadata.Name + vars.InternalAuthSecretSuffix,
		RedisAuthSecretName:    redisAuthSecretName,
		StorageAuthSecretName:  storageAuthSecretName,
		TrivyEnabled:           trivyEnabled,
		InternalDatabase:       spec.GetDatabase().GetInternal() != nil,
		InternalRedis:          spec.GetCache().GetInternal() != nil,
	}
}

// kubeEndpoint renders the in-cluster front-door URL.
func (l *Locals) kubeEndpoint() string {
	return fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
		l.FrontDoorScheme, l.ReleaseName, l.Namespace, l.FrontDoorPort)
}

// portForwardCommand renders the workstation access recipe.
func (l *Locals) portForwardCommand() string {
	localPort := 8080
	if l.FrontDoorScheme == "https" {
		localPort = 8443
	}
	return fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d",
		l.Namespace, l.ReleaseName, localPort, l.FrontDoorPort)
}
