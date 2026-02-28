package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesopenbaov1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesopenbao/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream OpenBao Helm chart and tailors it to the spec.
func helmChart(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	target := locals.KubernetesOpenBao
	spec := target.Spec

	sealHcl := sealConfigHcl(spec)

	// Build helm values based on spec
	helmValues := pulumi.Map{
		"fullnameOverride": pulumi.String(target.Metadata.Name),
		"global": pulumi.Map{
			"enabled":    pulumi.Bool(true),
			"tlsDisable": pulumi.Bool(!spec.TlsEnabled),
		},
		"server": pulumi.Map{
			"extraLabels": convertstringmaps.ConvertGoStringMapToPulumiMap(locals.Labels),
			"dataStorage": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"size":    pulumi.String(spec.ServerContainer.DataStorageSize),
			},
		},
	}

	// Configure server resources if provided
	if spec.ServerContainer != nil && spec.ServerContainer.Resources != nil {
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["resources"] = containerresources.ConvertToPulumiMap(spec.ServerContainer.Resources)
	}

	// Configure standalone vs HA mode
	if locals.HaEnabled {
		// HA mode with Raft
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["ha"] = pulumi.Map{
			"enabled":  pulumi.Bool(true),
			"replicas": pulumi.Int(locals.HaReplicas),
			"raft": pulumi.Map{
				"enabled":   pulumi.Bool(true),
				"setNodeId": pulumi.Bool(true),
				"config": pulumi.String(`ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "raft" {
  path = "/openbao/data"
}

service_registration "kubernetes" {}
` + sealHcl),
			},
		}
		serverMap["standalone"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	} else {
		// Standalone mode
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["standalone"] = pulumi.Map{
			"enabled": pulumi.Bool(true),
			"config": pulumi.String(`ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "file" {
  path = "/openbao/data"
}
` + sealHcl),
		}
		serverMap["ha"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	}

	// Configure Workload Identity service account annotation for GCP KMS auto-unseal
	if sa := workloadIdentityServiceAccount(spec); sa != "" {
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["serviceAccount"] = pulumi.Map{
			"annotations": pulumi.Map{
				"iam.gke.io/gcp-service-account": pulumi.String(sa),
			},
		}
	}

	// Configure UI
	uiEnabled := true
	if spec.UiEnabled != nil {
		uiEnabled = *spec.UiEnabled
	}
	helmValues["ui"] = pulumi.Map{
		"enabled": pulumi.Bool(uiEnabled),
	}

	// Configure injector if enabled
	if spec.Injector != nil && spec.Injector.Enabled {
		injectorReplicas := int32(1)
		if spec.Injector.Replicas != nil {
			injectorReplicas = *spec.Injector.Replicas
		}
		helmValues["injector"] = pulumi.Map{
			"enabled":  pulumi.Bool(true),
			"replicas": pulumi.Int(injectorReplicas),
		}
	} else {
		helmValues["injector"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	}

	// Install helm chart
	chartOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	chartOpts = append(chartOpts, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesOpenBao.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(locals.HelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values:    helmValues,
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(vars.HelmChartRepoUrl),
			},
		}, chartOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create helm chart")
	}

	return nil
}

// sealConfigHcl returns the HCL seal stanza for the configured auto-unseal method.
// Returns an empty string when auto-unseal is not configured.
func sealConfigHcl(spec *kubernetesopenbaov1.KubernetesOpenBaoSpec) string {
	if spec.AutoUnseal == nil {
		return ""
	}

	switch s := spec.AutoUnseal.Seal.(type) {
	case *kubernetesopenbaov1.KubernetesOpenBaoAutoUnseal_GcpKms:
		return fmt.Sprintf(`
seal "gcpckms" {
  project    = %q
  region     = %q
  key_ring   = %q
  crypto_key = %q
}
`, s.GcpKms.Project.GetValue(), s.GcpKms.Region,
			s.GcpKms.KeyRing.GetValue(), s.GcpKms.CryptoKey.GetValue())

	case *kubernetesopenbaov1.KubernetesOpenBaoAutoUnseal_AwsKms:
		return fmt.Sprintf(`
seal "awskms" {
  region     = %q
  kms_key_id = %q
}
`, s.AwsKms.Region, s.AwsKms.KmsKeyId)

	case *kubernetesopenbaov1.KubernetesOpenBaoAutoUnseal_AzureKeyVault:
		return fmt.Sprintf(`
seal "azurekeyvault" {
  vault_name = %q
  key_name   = %q
  tenant_id  = %q
}
`, s.AzureKeyVault.VaultName, s.AzureKeyVault.KeyName, s.AzureKeyVault.TenantId)

	case *kubernetesopenbaov1.KubernetesOpenBaoAutoUnseal_Transit:
		mountPath := s.Transit.MountPath
		if mountPath == "" {
			mountPath = "transit/"
		}
		return fmt.Sprintf(`
seal "transit" {
  address    = %q
  key_name   = %q
  mount_path = %q
}
`, s.Transit.Address, s.Transit.KeyName, mountPath)

	default:
		return ""
	}
}

// workloadIdentityServiceAccount extracts the GCP service account email from
// the GCP KMS seal config for Workload Identity annotation. Returns an empty
// string when GCP KMS is not configured or the field is not set.
func workloadIdentityServiceAccount(spec *kubernetesopenbaov1.KubernetesOpenBaoSpec) string {
	if spec.AutoUnseal == nil {
		return ""
	}
	gcpKms := spec.AutoUnseal.GetGcpKms()
	if gcpKms == nil || gcpKms.WorkloadIdentityServiceAccount == nil {
		return ""
	}
	return gcpKms.WorkloadIdentityServiceAccount.GetValue()
}
