package module

import (
	"fmt"
	"strconv"

	kuberneteskarapacev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarapace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Target *kuberneteskarapacev1alpha1.KubernetesKarapace
	Spec   *kuberneteskarapacev1alpha1.KubernetesKarapaceSpec

	// Resource-identity labels stamped on every module-created object.
	// The per-role "app" label rides on top of these (see the selector
	// maps below).
	Labels map[string]string

	// Immutable pod-selection identity per role. The two Deployments run
	// the SAME image in the same namespace — the role-specific "app"
	// value is what keeps each Service from selecting the other role's
	// pods (selectors AND all their labels).
	RegistrySelectorLabels map[string]string
	RestSelectorLabels     map[string]string

	// Namespace the registry deploys into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// RegistryName is metadata.name — the registry Deployment and
	// Service name. RestName is "<metadata.name>-rest" for the optional
	// REST-proxy role.
	RegistryName string
	RestName     string

	// Image resolves spec.image, falling back to the module's pinned
	// upstream release (vars.KarapaceImage).
	Image string

	RegistryReplicas int
	RegistryPort     int

	RestEnabled  bool
	RestReplicas int
	RestPort     int

	// Scheme the registry serves on: https when spec.server_tls is set,
	// http otherwise. Drives the endpoint output, the probe scheme, the
	// advertised protocol (leader forwarding), and the REST proxy's
	// registry_scheme wiring.
	Scheme string

	// Registry behavior with spec defaults applied.
	TopicName              string
	Compatibility          string
	GroupId                string
	MasterElectionStrategy string

	LogLevel         string
	SecurityProtocol string

	// SASL password wiring. The password NEVER lands in the pod spec as
	// a plaintext env value: when the spec declares a literal password,
	// the module materializes it into the Secret named SaslSecretName
	// (key "password") and the env var references that Secret — pod
	// specs are world-readable to anyone with get-pod RBAC, Secrets have
	// their own ACL. When the spec references an existing Secret
	// (password_secret), the env var references it directly.
	CreateSaslSecret       bool
	SaslSecretName         string
	SaslPasswordSecretName string
	SaslPasswordSecretKey  string

	// Output endpoints
	// (`scheme://<service>.<namespace>.svc.cluster.local:<port>`).
	Endpoint          string
	RestProxyEndpoint string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskarapacev1alpha1.KubernetesKarapaceIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKarapace.String(),
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

	namespace := spec.Namespace.GetValue()
	registryName := target.Metadata.Name
	restName := fmt.Sprintf("%s-rest", registryName)

	image := spec.Image
	if image == "" {
		image = vars.KarapaceImage
	}

	registryReplicas := 1
	if spec.Replicas != nil && *spec.Replicas > 0 {
		registryReplicas = int(*spec.Replicas)
	}

	registryPort := 8081
	if spec.GetPort() != 0 {
		registryPort = int(spec.GetPort())
	}

	restEnabled := spec.GetRestProxy().GetEnabled()
	restReplicas := 1
	if spec.GetRestProxy().GetReplicas() > 0 {
		restReplicas = int(spec.GetRestProxy().GetReplicas())
	}
	restPort := 8082
	if spec.GetRestProxy().GetPort() != 0 {
		restPort = int(spec.GetRestProxy().GetPort())
	}

	scheme := "http"
	if spec.ServerTls != nil {
		scheme = "https"
	}

	topicName := spec.GetRegistry().GetTopicName()
	if topicName == "" {
		topicName = "_schemas"
	}

	compatibility := spec.GetRegistry().GetCompatibility()
	if compatibility == "" {
		compatibility = "BACKWARD"
	}

	// LEADER-ELECTION CONTRACT: replicas of ONE installation coordinate
	// leadership through this Kafka consumer group, so the group id must
	// be unique per installation sharing a Kafka cluster — two
	// installations sharing a group id fight over leadership and corrupt
	// each other's view of the schemas topic. metadata.name is the
	// natural per-installation default; spec.registry.group_id overrides
	// it for deliberate cross-installation pairing (e.g. blue/green).
	groupId := spec.GetRegistry().GetGroupId()
	if groupId == "" {
		groupId = registryName
	}

	masterElectionStrategy := spec.GetRegistry().GetMasterElectionStrategy()
	if masterElectionStrategy == "" {
		masterElectionStrategy = "lowest"
	}

	logLevel := spec.GetLogLevel()
	if logLevel == "" {
		logLevel = "INFO"
	}

	securityProtocol := spec.GetKafka().GetSecurityProtocol()
	if securityProtocol == "" {
		securityProtocol = "PLAINTEXT"
	}

	// SASL password source resolution (see the Locals field comment for
	// the never-plaintext rationale).
	saslSecretName := fmt.Sprintf("%s-sasl", registryName)
	createSaslSecret := false
	saslPasswordSecretName := ""
	saslPasswordSecretKey := ""
	if sasl := spec.GetKafka().GetSasl(); sasl != nil {
		if sasl.PasswordSecret != nil {
			saslPasswordSecretName = sasl.PasswordSecret.SecretName.GetValue()
			saslPasswordSecretKey = sasl.PasswordSecret.GetKey()
			if saslPasswordSecretKey == "" {
				saslPasswordSecretKey = "password"
			}
		} else if sasl.Password != "" {
			createSaslSecret = true
			saslPasswordSecretName = saslSecretName
			saslPasswordSecretKey = "password"
		}
	}

	endpoint := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", scheme, registryName, namespace, registryPort)

	// The REST proxy always serves plain HTTP (spec.server_tls covers the
	// registry API only); empty when the role is not deployed — an honest
	// signal for composition.
	restProxyEndpoint := ""
	if restEnabled {
		restProxyEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", restName, namespace, restPort)
	}

	return &Locals{
		Target: target,
		Spec:   spec,
		Labels: labels,
		RegistrySelectorLabels: map[string]string{
			"app":                            registryName,
			kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		},
		RestSelectorLabels: map[string]string{
			"app":                            restName,
			kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		},
		Namespace:              namespace,
		RegistryName:           registryName,
		RestName:               restName,
		Image:                  image,
		RegistryReplicas:       registryReplicas,
		RegistryPort:           registryPort,
		RestEnabled:            restEnabled,
		RestReplicas:           restReplicas,
		RestPort:               restPort,
		Scheme:                 scheme,
		TopicName:              topicName,
		Compatibility:          compatibility,
		GroupId:                groupId,
		MasterElectionStrategy: masterElectionStrategy,
		LogLevel:               logLevel,
		SecurityProtocol:       securityProtocol,
		CreateSaslSecret:       createSaslSecret,
		SaslSecretName:         saslSecretName,
		SaslPasswordSecretName: saslPasswordSecretName,
		SaslPasswordSecretKey:  saslPasswordSecretKey,
		Endpoint:               endpoint,
		RestProxyEndpoint:      restProxyEndpoint,
	}
}

// mergedLabels returns the identity label set with the role-specific
// selector labels layered on top (the "app" per-role value).
func mergedLabels(locals *Locals, selectorLabels map[string]string) map[string]string {
	merged := make(map[string]string, len(locals.Labels)+len(selectorLabels))
	for k, v := range locals.Labels {
		merged[k] = v
	}
	for k, v := range selectorLabels {
		merged[k] = v
	}
	return merged
}
