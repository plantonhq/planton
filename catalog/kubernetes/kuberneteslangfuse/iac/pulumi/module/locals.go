package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf - keep them in lockstep.
type Locals struct {
	Spec *kuberneteslangfusev1alpha1.KubernetesLangfuseSpec

	// Resource-identity labels: stamped on the module-created objects
	// (namespace, instance Secret) and on both workloads' pods.
	Labels map[string]string

	// Namespace Langfuse installs into.
	Namespace string

	// ReleaseName is metadata.name. fullnameOverride pins the chart
	// fullname to it: the web Deployment and Service are `<name>-web`, the
	// worker `<name>-worker`, the ServiceAccount `<name>`.
	ReleaseName string

	// ChartVersion resolved to the pinned default when unset.
	ChartVersion string

	// ExternalUrl resolved to the spec default when unset.
	ExternalUrl string

	// WebServiceName / WebPort identify the chart's web Service and the
	// port the web server listens on; WebServicePort is the port the
	// Service exposes (external_port when set).
	WebServiceName string
	WebPort        int
	WebServicePort int

	// InstanceSecretName is the module-owned Secret holding every secret
	// value; it always exists (the salt, encryption key and NextAuth secret
	// live there).
	InstanceSecretName string

	// Seeding (headless initialization).
	InitEnabled      bool
	OrganizationId   string
	ProjectId        string
	AdminEmail       string
	AdminSeeded      bool
	DeclaredInitKeys bool

	// DeclaredSecrets holds every secret value the spec declares, keyed by
	// its instance Secret key. Generated values join them in secrets.go.
	DeclaredSecrets map[string]string

	// Generate records which credentials the module generates because the
	// spec leaves them empty.
	Generate GeneratedCredentials

	// SecretChecksum is the sha256 of the declared secret values (JSON,
	// sorted keys) - the pod annotation that rolls both workloads when a
	// declared value changes. Generated values never change after creation
	// and are left out.
	SecretChecksum string

	// The environment the chart's additionalEnv channels carry: CommonEnv
	// on both workloads, WebEnv on the web server only, WorkerEnv on the
	// worker only. Each sorted by variable name.
	CommonEnv []envVar
	WebEnv    []envVar
	WorkerEnv []envVar

	// Volumes / VolumeMounts land in the chart's extraVolumes and
	// extraVolumeMounts (both workloads).
	Volumes      []interface{}
	VolumeMounts []interface{}

	// File paths inside the containers ("" when not mounted).
	RedisTlsCaPath   string
	RedisTlsCertPath string
	RedisTlsKeyPath  string
	TrustedCaPath    string

	// Blob storage targets resolved per upload type.
	EventTarget *blobTarget
	MediaTarget *blobTarget
	BatchTarget *blobTarget

	// Composition handles (twin: outputs.tf).
	Endpoint           string
	OtelEndpoint       string
	PortForwardCommand string
}

// GeneratedCredentials flags each credential the module generates.
type GeneratedCredentials struct {
	Salt             bool
	EncryptionKey    bool
	NextauthSecret   bool
	ProjectKeys      bool
	AdminPassword    bool
	SlackStateSecret bool
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteslangfusev1alpha1.KubernetesLangfuseIacInput) (*Locals, error) {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesLangfuse.String(),
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
	externalUrl := spec.GetExternalUrl()
	if externalUrl == "" {
		externalUrl = vars.DefaultExternalUrl
	}

	releaseName := target.Metadata.Name
	namespace := spec.GetNamespace().GetValue()
	webPort := vars.DefaultWebPort
	if spec.GetWeb().GetService() != nil && spec.GetWeb().GetService().Port != nil {
		webPort = int(spec.GetWeb().GetService().GetPort())
	}
	// The Service exposes external_port when set, the container port
	// otherwise (the chart's `externalPort | default port`).
	servicePort := webPort
	if spec.GetWeb().GetService() != nil && spec.GetWeb().GetService().ExternalPort != nil {
		servicePort = int(spec.GetWeb().GetService().GetExternalPort())
	}
	webServiceName := releaseName + vars.WebComponentSuffix

	locals := &Locals{
		Spec:               spec,
		Labels:             labels,
		Namespace:          namespace,
		ReleaseName:        releaseName,
		ChartVersion:       chartVersion,
		ExternalUrl:        externalUrl,
		WebServiceName:     webServiceName,
		WebPort:            webPort,
		WebServicePort:     servicePort,
		InstanceSecretName: releaseName + vars.InstanceSecretSuffix,
		DeclaredSecrets:    map[string]string{},
		Endpoint:           fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", webServiceName, namespace, servicePort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d", namespace, webServiceName, servicePort, servicePort),
	}
	locals.OtelEndpoint = locals.Endpoint + "/api/public/otel"

	locals.resolveInstanceSecrets()
	locals.resolveInit(releaseName)
	locals.resolveBlobTargets()
	locals.resolveFiles()
	locals.buildEnvironment()

	checksum, err := checksumOf(locals.DeclaredSecrets)
	if err != nil {
		return nil, err
	}
	locals.SecretChecksum = checksum

	return locals, nil
}

// resolveInstanceSecrets records the declared instance secrets and flags
// the ones to generate.
func (l *Locals) resolveInstanceSecrets() {
	s := l.Spec.GetSecrets()
	l.declareOrGenerate(keySalt, s.GetSalt(), &l.Generate.Salt)
	l.declareOrGenerate(keyEncryptionKey, s.GetEncryptionKey(), &l.Generate.EncryptionKey)
	l.declareOrGenerate(keyNextauthSecret, s.GetNextauthSecret(), &l.Generate.NextauthSecret)

	l.declare(keyPostgresDirectUrl, l.Spec.GetPostgres().GetDirectUrl())
	l.declare(keyPostgresShadowUrl, l.Spec.GetPostgres().GetShadowDatabaseUrl())
	l.declare(keyRedisTlsKeyPassphrase, l.Spec.GetRedis().GetTls().GetClientKeyPassphrase())
	l.declare(keySmtpConnectionUrl, l.Spec.GetSmtp().GetConnectionUrl())
	l.declare(keyAiApiKey, l.Spec.GetAiFeatures().GetApiKey())
	l.declare(keyLicenseKey, l.Spec.GetEnterprise().GetLicenseKey())
	l.declare(keyAdminApiKey, l.Spec.GetEnterprise().GetAdminApiKey())
	if slack := l.Spec.GetSlack(); slack != nil {
		l.declare(keySlackClientSecret, slack.GetClientSecret())
		l.declareOrGenerate(keySlackStateSecret, slack.GetStateSecret(), &l.Generate.SlackStateSecret)
	}
	for _, p := range ssoProviders(l.Spec.GetAuth().GetSso()) {
		l.declare(fmt.Sprintf(keySsoClientSecretTmpl, p.Id), p.Client.GetClientSecret())
	}
}

// resolveInit resolves headless initialization: the seeded IDs, the key
// pair and the admin user.
func (l *Locals) resolveInit(releaseName string) {
	init := l.Spec.GetInit()
	l.InitEnabled = init == nil || init.Enabled == nil || init.GetEnabled()
	if !l.InitEnabled {
		return
	}
	l.OrganizationId = init.GetOrganization().GetId()
	if l.OrganizationId == "" {
		l.OrganizationId = releaseName
	}
	l.ProjectId = init.GetProject().GetId()
	if l.ProjectId == "" {
		l.ProjectId = releaseName + vars.ProjectIdSuffix
	}
	if pk := init.GetProject().GetPublicKey(); pk != "" {
		l.DeclaredInitKeys = true
		l.declare(keyInitProjectPublicKey, pk)
		l.declare(keyInitProjectSecretKey, init.GetProject().GetSecretKey())
	} else {
		l.Generate.ProjectKeys = true
	}
	if admin := init.GetAdmin(); admin != nil {
		l.AdminSeeded = true
		l.AdminEmail = admin.GetEmail()
		l.declareOrGenerate(keyInitUserPassword, admin.GetPassword(), &l.Generate.AdminPassword)
	}
}

func (l *Locals) declare(key, value string) {
	if value != "" {
		l.DeclaredSecrets[key] = value
	}
}

func (l *Locals) declareOrGenerate(key, value string, generate *bool) {
	if value != "" {
		l.DeclaredSecrets[key] = value
		return
	}
	*generate = true
}

// resolveFiles builds the volumes the module projects into both workloads,
// in this order in both engines: the Redis TLS material, the trusted CA
// bundle, then spec.extra_volumes.
func (l *Locals) resolveFiles() {
	tls := l.Spec.GetRedis().GetTls()
	if ca := tls.GetCa(); ca != nil {
		l.RedisTlsCaPath = l.projectCertificateFile("redis-tls-ca", vars.RedisTlsCaDir, ca)
	}
	if cert := tls.GetClientCertificate(); cert != nil {
		l.RedisTlsCertPath = l.projectSecretFile("redis-tls-client-cert", vars.RedisTlsClientCertDir, cert, "tls.crt")
	}
	if key := tls.GetClientKey(); key != nil {
		l.RedisTlsKeyPath = l.projectSecretFile("redis-tls-client-key", vars.RedisTlsClientKeyDir, key, "tls.key")
	}
	if ca := l.Spec.GetNetwork().GetTrustedCa(); ca != nil {
		l.TrustedCaPath = l.projectCertificateFile("trusted-ca", vars.TrustedCaDir, ca)
	}
	for _, v := range l.Spec.GetExtraVolumes() {
		l.Volumes = append(l.Volumes, volumeBody(v))
		l.VolumeMounts = append(l.VolumeMounts, volumeMountBody(v))
	}
}

// projectCertificateFile mounts one key of a Secret or ConfigMap as a file
// named after the key, in its own directory (a directory mount, never
// subPath, so a rotated certificate reaches the pod).
func (l *Locals) projectCertificateFile(volumeName, dir string, f *kuberneteslangfusev1alpha1.KubernetesLangfuseCertificateFile) string {
	key := f.GetKey()
	if key == "" {
		key = "ca.crt"
	}
	items := []interface{}{map[string]interface{}{"key": key, "path": key}}
	volume := map[string]interface{}{"name": volumeName}
	if f.GetSecretName() != "" {
		volume["secret"] = map[string]interface{}{"secretName": f.GetSecretName(), "items": items}
	} else {
		volume["configMap"] = map[string]interface{}{"name": f.GetConfigMapName(), "items": items}
	}
	l.Volumes = append(l.Volumes, volume)
	l.VolumeMounts = append(l.VolumeMounts, map[string]interface{}{"name": volumeName, "mountPath": dir, "readOnly": true})
	return dir + "/" + key
}

func (l *Locals) projectSecretFile(volumeName, dir string, f *kuberneteslangfusev1alpha1.KubernetesLangfuseSecretFile, defaultKey string) string {
	key := f.GetKey()
	if key == "" {
		key = defaultKey
	}
	l.Volumes = append(l.Volumes, map[string]interface{}{
		"name": volumeName,
		"secret": map[string]interface{}{
			"secretName": f.GetSecretName().GetValue(),
			"items":      []interface{}{map[string]interface{}{"key": key, "path": key}},
		},
	})
	l.VolumeMounts = append(l.VolumeMounts, map[string]interface{}{"name": volumeName, "mountPath": dir, "readOnly": true})
	return dir + "/" + key
}

// checksumOf hashes a string map as sorted-key JSON (Terraform twin:
// sha256(jsonencode(...)), which sorts keys and escapes identically).
func checksumOf(m map[string]string) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// csv joins a list for comma-separated Langfuse variables.
func csv(values []string) string {
	return strings.Join(values, ",")
}
