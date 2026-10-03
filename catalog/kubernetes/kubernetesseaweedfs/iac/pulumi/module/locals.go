package module

import (
	"fmt"
	"strconv"

	kubernetesseaweedfsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesseaweedfs/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kubernetesseaweedfsv1alpha1.KubernetesSeaweedFsSpec

	// Resource-identity labels stamped on the module-created satellites
	// (namespace, the admin-auth Secret — never injected into the chart's
	// own resources; Helm owns those).
	Labels map[string]string

	// Namespace SeaweedFS installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name — metadata.name, NOT a fixed chart name: several
	// SeaweedFS stores coexist in one cluster. fullnameOverride is pinned
	// to this (values.go), so every componentName-derived child
	// (`<name>-master`, `-filer`, `-s3`, `-admin`, `-volume`) is
	// deterministic.
	ReleaseName string

	// Chart version resolved to the pinned default when unset, so both
	// engines install the same chart whether or not the platform's
	// defaulting middleware ran.
	ChartVersion string

	// The S3 gateway posture, resolved from the spec's optional bools
	// (component defaults: enabled with auth). Dedicated means the
	// gateway runs as its own Deployment; embedded runs it on the filer.
	S3Enabled    bool
	S3Auth       bool
	S3Dedicated  bool
	S3ConfigName string // spec.s3.existing_config_secret, or "" when chart-generated

	// Chart-derived service names (componentName = "<fullname>-<suffix>").
	MasterServiceName string
	FilerServiceName  string
	S3ServiceName     string
	AdminServiceName  string

	// In-cluster S3 endpoint (the `-s3` Service exists for the embedded
	// and dedicated shapes alike). Empty when the gateway is disabled.
	S3Endpoint string

	// Name of the Secret carrying the S3 credentials: the
	// chart-generated `<fullname>-s3-secret` (admin + read-only pairs,
	// stable across upgrades, kept on uninstall), the referenced existing
	// config secret, or "" when auth is off.
	S3CredentialsSecretName string

	// Admin console posture.
	AdminEnabled bool
	// Whether the module materializes the console credentials Secret
	// (admin enabled without an existing secret).
	CreateAdminSecret bool
	// Name of the Secret carrying the console credentials — the
	// module-materialized `<name>-admin-auth` (keys user/password) or
	// the referenced existing secret. Empty when the console is off.
	AdminAuthSecretName string
	// In-cluster console endpoint; empty when the console is off.
	AdminEndpoint string

	// kubectl one-liner for reaching S3 from a workstation.
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kubernetesseaweedfsv1alpha1.KubernetesSeaweedFsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesSeaweedFs.String(),
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

	// S3 posture: the optional bools default to TRUE (this kind IS the
	// catalog's S3 store) — only an explicit false turns them off.
	s3 := spec.GetS3()
	s3Enabled := s3 == nil || s3.Enabled == nil || s3.GetEnabled()
	s3Auth := s3 == nil || s3.EnableAuth == nil || s3.GetEnableAuth()
	s3Dedicated := s3.GetDedicated() != nil
	s3ConfigName := s3.GetExistingConfigSecret()

	masterServiceName := releaseName + "-master"
	filerServiceName := releaseName + "-filer"
	s3ServiceName := releaseName + "-s3"
	adminServiceName := releaseName + "-admin"

	s3Endpoint := ""
	portForwardCommand := ""
	if s3Enabled {
		s3Endpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:8333", s3ServiceName, namespace)
		portForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s 8333:8333", s3ServiceName, namespace)
	}

	s3CredentialsSecretName := ""
	if s3Enabled && s3Auth {
		if s3ConfigName != "" {
			s3CredentialsSecretName = s3ConfigName
		} else {
			s3CredentialsSecretName = releaseName + "-s3-secret"
		}
	}

	adminEnabled := spec.GetAdmin().GetEnabled()
	createAdminSecret := adminEnabled && spec.GetAdmin().GetExistingAuthSecret() == ""
	adminAuthSecretName := ""
	adminEndpoint := ""
	if adminEnabled {
		if createAdminSecret {
			adminAuthSecretName = releaseName + "-admin-auth"
		} else {
			adminAuthSecretName = spec.GetAdmin().GetExistingAuthSecret()
		}
		adminEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:23646", adminServiceName, namespace)
	}

	return &Locals{
		Spec:                    spec,
		Labels:                  labels,
		Namespace:               namespace,
		ReleaseName:             releaseName,
		ChartVersion:            chartVersion,
		S3Enabled:               s3Enabled,
		S3Auth:                  s3Auth,
		S3Dedicated:             s3Dedicated,
		S3ConfigName:            s3ConfigName,
		MasterServiceName:       masterServiceName,
		FilerServiceName:        filerServiceName,
		S3ServiceName:           s3ServiceName,
		AdminServiceName:        adminServiceName,
		S3Endpoint:              s3Endpoint,
		S3CredentialsSecretName: s3CredentialsSecretName,
		AdminEnabled:            adminEnabled,
		CreateAdminSecret:       createAdminSecret,
		AdminAuthSecretName:     adminAuthSecretName,
		AdminEndpoint:           adminEndpoint,
		PortForwardCommand:      portForwardCommand,
	}
}
