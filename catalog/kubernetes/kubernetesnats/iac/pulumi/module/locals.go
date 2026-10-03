package module

import (
	"fmt"
	"strconv"

	kubernetesnatsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnats/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// authUser pairs a declared username with its deterministic password
// env-var name — ONE ordering shared by the auth Secret, the container
// env and the rendered config (and byte-identical in the Terraform twin):
// flat users indexed in spec order; account users indexed
// `<account-index>_<user-index>`.
type authUser struct {
	Username string
	EnvVar   string
}

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesnatsv1alpha1.KubernetesNatsSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace and the auth Secret — never injected into the
	// chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace NATS installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name:
	// several NATS systems can coexist in one cluster.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// Whether auth is declared (users or accounts) — drives the auth
	// Secret and the config's authorization/accounts rendering.
	AuthEnabled bool

	// Name of the module-generated credential Secret
	// (`<metadata.name>-auth`, one key per username); "" without auth.
	AuthSecretName string

	// Every declared user (flat, then per account in spec order) with
	// its deterministic password env-var name.
	FlatUsers    []authUser
	AccountUsers [][]authUser

	// Whether JetStream is on (spec default true — the kind's
	// persistent-messaging posture).
	JetStreamEnabled bool

	// Name of the client Service (equals metadata.name via
	// fullnameOverride) and the headless sibling.
	ServiceName         string
	HeadlessServiceName string

	// In-cluster client endpoint (nats:// URL).
	ClientEndpoint string

	// In-cluster WebSocket endpoint ("" when the listener is off).
	WebsocketEndpoint string

	// kubectl one-liner for reaching the client port from a
	// workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesnatsv1alpha1.KubernetesNatsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesNats.String(),
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

	auth := spec.GetAuth()
	authEnabled := auth != nil && (len(auth.GetUsers()) > 0 || len(auth.GetAccounts()) > 0)
	authSecretName := ""
	var flatUsers []authUser
	var accountUsers [][]authUser
	if authEnabled {
		authSecretName = fmt.Sprintf("%s-auth", releaseName)
		for i, u := range auth.GetUsers() {
			flatUsers = append(flatUsers, authUser{
				Username: u.GetUsername(),
				EnvVar:   fmt.Sprintf("%s%d", vars.PasswordEnvPrefix, i),
			})
		}
		for ai, account := range auth.GetAccounts() {
			var users []authUser
			for ui, u := range account.GetUsers() {
				users = append(users, authUser{
					Username: u.GetUsername(),
					EnvVar:   fmt.Sprintf("%s%d_%d", vars.PasswordEnvPrefix, ai, ui),
				})
			}
			accountUsers = append(accountUsers, users)
		}
	}

	jetStreamEnabled := true
	if spec.GetJetStream() != nil && spec.GetJetStream().Enabled != nil {
		jetStreamEnabled = spec.GetJetStream().GetEnabled()
	}

	// fullnameOverride pins the fullname to metadata.name (values.go),
	// so the client Service is exactly `<name>` and the headless
	// sibling `<name>-headless`.
	serviceName := releaseName
	headlessServiceName := fmt.Sprintf("%s-headless", releaseName)
	clientEndpoint := fmt.Sprintf("nats://%s.%s.svc.cluster.local:%d",
		serviceName, namespace, vars.ClientPort)

	websocketEndpoint := ""
	if ws := spec.GetWebsocket(); ws != nil && ws.GetEnabled() {
		port := int32(8080)
		if ws.Port != nil {
			port = ws.GetPort()
		}
		websocketEndpoint = fmt.Sprintf("ws://%s.%s.svc.cluster.local:%d",
			serviceName, namespace, port)
	}

	portForwardCommand := fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		serviceName, namespace, vars.ClientPort, vars.ClientPort)

	return &Locals{
		Spec:                spec,
		Labels:              labels,
		Namespace:           namespace,
		ReleaseName:         releaseName,
		ChartVersion:        chartVersion,
		AuthEnabled:         authEnabled,
		AuthSecretName:      authSecretName,
		FlatUsers:           flatUsers,
		AccountUsers:        accountUsers,
		JetStreamEnabled:    jetStreamEnabled,
		ServiceName:         serviceName,
		HeadlessServiceName: headlessServiceName,
		ClientEndpoint:      clientEndpoint,
		WebsocketEndpoint:   websocketEndpoint,
		PortForwardCommand:  portForwardCommand,
	}
}
