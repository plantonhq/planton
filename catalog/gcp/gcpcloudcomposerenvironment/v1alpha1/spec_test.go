package gcpcloudcomposerenvironmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudComposerEnvironmentSpec Suite")
}

var _ = ginkgo.Describe("GcpCloudComposerEnvironmentSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpCloudComposerEnvironment.
	minimal := func() *GcpCloudComposerEnvironment {
		return &GcpCloudComposerEnvironment{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudComposerEnvironment",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-composer-env",
			},
			Spec: &GcpCloudComposerEnvironmentSpec{
				ProjectId: svr("my-gcp-project"),
				Region:    "us-central1",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an omitted project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a multi-digit region", func() {
		msg := minimal()
		msg.Spec.Region = "europe-west12"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept environment_name with valid format", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "my-airflow-prod-2026"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty environment_name (uses metadata.name)", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node_config with VPC networking", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			Network:        svr("projects/my-project/global/networks/my-vpc"),
			Subnetwork:     svr("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
			ServiceAccount: svr("composer-sa@my-project.iam.gserviceaccount.com"),
			Tags:           []string{"composer", "airflow"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node_config with Composer 3 PSC networking", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			ComposerNetworkAttachment:     "projects/my-project/regions/us-central1/networkAttachments/my-attachment",
			ComposerInternalIpv4CidrBlock: "10.0.0.0/20",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node_config with enable_ip_masq_agent", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			Network:           svr("projects/my-project/global/networks/my-vpc"),
			Subnetwork:        svr("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
			EnableIpMasqAgent: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ip_allocation_policy with named secondary ranges", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			IpAllocationPolicy: &GcpCloudComposerIpAllocationPolicy{
				ClusterSecondaryRangeName:  "composer-pods",
				ServicesSecondaryRangeName: "composer-services",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ip_allocation_policy with CIDR blocks", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			IpAllocationPolicy: &GcpCloudComposerIpAllocationPolicy{
				ClusterIpv4CidrBlock:  "10.4.0.0/14",
				ServicesIpv4CidrBlock: "10.8.0.0/20",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ip_allocation_policy mixing a named pod range with a services CIDR", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			IpAllocationPolicy: &GcpCloudComposerIpAllocationPolicy{
				ClusterSecondaryRangeName: "composer-pods",
				ServicesIpv4CidrBlock:     "10.8.0.0/20",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept software_config with image version and packages", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			ImageVersion: "composer-2.9.7-airflow-2.9.3",
			PypiPackages: map[string]string{
				"numpy":    ">=1.21",
				"requests": "",
			},
			AirflowConfigOverrides: map[string]string{
				"core-dags_are_paused_at_creation": "True",
			},
			EnvVariables: map[string]string{
				"MY_CUSTOM_VAR": "custom-value",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept web_server_plugins_mode ENABLED", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			WebServerPluginsMode: "ENABLED",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept web_server_plugins_mode DISABLED", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			WebServerPluginsMode: "DISABLED",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cloud_data_lineage_integration", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			CloudDataLineageIntegration: &GcpCloudComposerCloudDataLineageIntegration{
				Enabled: true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept private_environment_config with VPC_PEERING", func() {
		msg := minimal()
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			EnablePrivateEndpoint:             true,
			ConnectionType:                    "VPC_PEERING",
			MasterIpv4CidrBlock:               "172.16.0.0/28",
			CloudSqlIpv4CidrBlock:             "10.0.32.0/20",
			CloudComposerNetworkIpv4CidrBlock: "10.0.48.0/20",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept private_environment_config with PRIVATE_SERVICE_CONNECT", func() {
		msg := minimal()
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			ConnectionType:                    "PRIVATE_SERVICE_CONNECT",
			CloudComposerConnectionSubnetwork: "projects/my-project/regions/us-central1/subnetworks/psc-subnet",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept environment_size ENVIRONMENT_SIZE_SMALL", func() {
		msg := minimal()
		msg.Spec.EnvironmentSize = "ENVIRONMENT_SIZE_SMALL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept environment_size ENVIRONMENT_SIZE_MEDIUM", func() {
		msg := minimal()
		msg.Spec.EnvironmentSize = "ENVIRONMENT_SIZE_MEDIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept environment_size ENVIRONMENT_SIZE_LARGE", func() {
		msg := minimal()
		msg.Spec.EnvironmentSize = "ENVIRONMENT_SIZE_LARGE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept environment_size ENVIRONMENT_SIZE_EXTRA_LARGE", func() {
		msg := minimal()
		msg.Spec.EnvironmentSize = "ENVIRONMENT_SIZE_EXTRA_LARGE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept resilience_mode STANDARD_RESILIENCE", func() {
		msg := minimal()
		msg.Spec.ResilienceMode = "STANDARD_RESILIENCE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept resilience_mode HIGH_RESILIENCE", func() {
		msg := minimal()
		msg.Spec.ResilienceMode = "HIGH_RESILIENCE"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CMEK encryption via kms_key_name", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = svr("projects/my-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept workloads_config with all components", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Scheduler: &GcpCloudComposerWorkloadResource{
				Cpu:       2.0,
				MemoryGb:  7.5,
				StorageGb: 5.0,
				Count:     2,
			},
			WebServer: &GcpCloudComposerWebServerResource{
				Cpu:       2.0,
				MemoryGb:  7.5,
				StorageGb: 5.0,
			},
			Worker: &GcpCloudComposerWorkerResource{
				Cpu:       2.0,
				MemoryGb:  7.5,
				StorageGb: 5.0,
				MinCount:  2,
				MaxCount:  6,
			},
			Triggerer: &GcpCloudComposerTriggererResource{
				Cpu:      1.0,
				MemoryGb: 1.0,
				Count:    2,
			},
			DagProcessor: &GcpCloudComposerWorkloadResource{
				Cpu:       1.0,
				MemoryGb:  2.0,
				StorageGb: 1.0,
				Count:     2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept worker with equal min_count and max_count", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Worker: &GcpCloudComposerWorkerResource{
				MinCount: 3,
				MaxCount: 3,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept worker with zero max_count (unset)", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Worker: &GcpCloudComposerWorkerResource{
				MinCount: 2,
				MaxCount: 0,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept maintenance_window configuration", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpCloudComposerMaintenanceWindow{
			StartTime:  "2026-01-01T00:00:00Z",
			EndTime:    "2026-01-01T12:00:00Z",
			Recurrence: "FREQ=WEEKLY;BYDAY=TU,WE,TH",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept recovery_config with scheduled snapshots", func() {
		msg := minimal()
		msg.Spec.RecoveryConfig = &GcpCloudComposerRecoveryConfig{
			Enabled:                  true,
			SnapshotLocation:         "gs://my-bucket/composer-snapshots",
			SnapshotCreationSchedule: "0 4 * * *",
			TimeZone:                 "America/Los_Angeles",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept web_server_network_access_control", func() {
		msg := minimal()
		msg.Spec.WebServerNetworkAccessControl = &GcpCloudComposerWebServerAccessControl{
			AllowedIpRanges: []*GcpCloudComposerAllowedIpRange{
				{Value: "10.0.0.0/8", Description: "Internal network"},
				{Value: "203.0.113.0/24", Description: "Office network"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept master_authorized_networks_config with cidr_blocks", func() {
		msg := minimal()
		msg.Spec.MasterAuthorizedNetworksConfig = &GcpCloudComposerMasterAuthorizedNetworksConfig{
			Enabled: true,
			CidrBlocks: []*GcpCloudComposerCidrBlock{
				{CidrBlock: "10.0.0.0/8", DisplayName: "corp-network"},
				{CidrBlock: "203.0.113.0/24"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept data_retention_config with valid modes", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			TaskLogsStorageMode:          "CLOUD_LOGGING_ONLY",
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 90,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept airflow_metadata_retention_days at the 30-day floor", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 30,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept airflow_metadata_retention_days at the 730-day ceiling", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 730,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an unset airflow_metadata_retention_days", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			TaskLogsStorageMode: "CLOUD_LOGGING_AND_CLOUD_STORAGE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a storage_bucket literal", func() {
		msg := minimal()
		msg.Spec.StorageBucket = svr("my-composer-dags-bucket")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a reference-shaped storage_bucket", func() {
		msg := minimal()
		msg.Spec.StorageBucket = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "composer-dags"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{
			"team":        "data-platform",
			"cost-center": "cc-1234",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept Composer 3 private environment flags", func() {
		msg := minimal()
		msg.Spec.EnablePrivateEnvironment = true
		msg.Spec.EnablePrivateBuildsOnly = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-featured spec", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "prod-airflow"
		msg.Spec.EnvironmentSize = "ENVIRONMENT_SIZE_MEDIUM"
		msg.Spec.ResilienceMode = "HIGH_RESILIENCE"
		msg.Spec.KmsKeyName = svr("projects/my-project/locations/us-central1/keyRings/ring/cryptoKeys/key")
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			Network:        svr("projects/my-project/global/networks/my-vpc"),
			Subnetwork:     svr("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
			ServiceAccount: svr("composer@my-project.iam.gserviceaccount.com"),
			Tags:           []string{"composer"},
		}
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			ImageVersion: "composer-2.9.7-airflow-2.9.3",
			PypiPackages: map[string]string{"apache-airflow-providers-google": ""},
			AirflowConfigOverrides: map[string]string{
				"webserver-dag_default_view": "grid",
			},
		}
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			EnablePrivateEndpoint:             true,
			ConnectionType:                    "VPC_PEERING",
			MasterIpv4CidrBlock:               "172.16.0.0/28",
			CloudSqlIpv4CidrBlock:             "10.0.32.0/20",
			CloudComposerNetworkIpv4CidrBlock: "10.0.48.0/20",
		}
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Scheduler: &GcpCloudComposerWorkloadResource{
				Cpu: 2.0, MemoryGb: 7.5, StorageGb: 5.0, Count: 2,
			},
			WebServer: &GcpCloudComposerWebServerResource{
				Cpu: 2.0, MemoryGb: 7.5, StorageGb: 5.0,
			},
			Worker: &GcpCloudComposerWorkerResource{
				Cpu: 2.0, MemoryGb: 7.5, StorageGb: 5.0, MinCount: 2, MaxCount: 6,
			},
			Triggerer: &GcpCloudComposerTriggererResource{
				Cpu: 1.0, MemoryGb: 1.0, Count: 2,
			},
		}
		msg.Spec.MaintenanceWindow = &GcpCloudComposerMaintenanceWindow{
			StartTime:  "2026-01-01T00:00:00Z",
			EndTime:    "2026-01-01T12:00:00Z",
			Recurrence: "FREQ=WEEKLY;BYDAY=SA,SU",
		}
		msg.Spec.RecoveryConfig = &GcpCloudComposerRecoveryConfig{
			Enabled:                  true,
			SnapshotLocation:         "gs://my-bucket/snapshots",
			SnapshotCreationSchedule: "0 2 * * *",
			TimeZone:                 "UTC",
		}
		msg.Spec.WebServerNetworkAccessControl = &GcpCloudComposerWebServerAccessControl{
			AllowedIpRanges: []*GcpCloudComposerAllowedIpRange{
				{Value: "10.0.0.0/8", Description: "VPN"},
			},
		}
		msg.Spec.MasterAuthorizedNetworksConfig = &GcpCloudComposerMasterAuthorizedNetworksConfig{
			Enabled: true,
			CidrBlocks: []*GcpCloudComposerCidrBlock{
				{CidrBlock: "10.0.0.0/8", DisplayName: "corp-network"},
			},
		}
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			TaskLogsStorageMode:          "CLOUD_LOGGING_ONLY",
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 90,
		}
		msg.Spec.StorageBucket = svr("my-composer-dags-bucket")
		msg.Spec.Labels = map[string]string{"team": "data-platform"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing region", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a malformed region", func() {
		msg := minimal()
		msg.Spec.Region = "US-Central1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a zone-shaped region", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1-a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject environment_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "1-invalid-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject environment_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "My-Airflow-Env"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject environment_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "my-env-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject environment_name with underscores", func() {
		msg := minimal()
		msg.Spec.EnvironmentName = "my_env_name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid environment_size", func() {
		msg := minimal()
		msg.Spec.EnvironmentSize = "EXTRA_LARGE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid resilience_mode", func() {
		msg := minimal()
		msg.Spec.ResilienceMode = "ULTRA_RESILIENCE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid connection_type", func() {
		msg := minimal()
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			ConnectionType: "DIRECT_CONNECT",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid web_server_plugins_mode", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			WebServerPluginsMode: "CUSTOM",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject worker max_count less than min_count", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Worker: &GcpCloudComposerWorkerResource{
				MinCount: 5,
				MaxCount: 2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject negative worker min_count", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Worker: &GcpCloudComposerWorkerResource{
				MinCount: -1,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject negative worker max_count", func() {
		msg := minimal()
		msg.Spec.WorkloadsConfig = &GcpCloudComposerWorkloadsConfig{
			Worker: &GcpCloudComposerWorkerResource{
				MaxCount: -1,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window missing start_time", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpCloudComposerMaintenanceWindow{
			EndTime:    "2026-01-01T12:00:00Z",
			Recurrence: "FREQ=DAILY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window missing end_time", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpCloudComposerMaintenanceWindow{
			StartTime:  "2026-01-01T00:00:00Z",
			Recurrence: "FREQ=DAILY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window missing recurrence", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpCloudComposerMaintenanceWindow{
			StartTime: "2026-01-01T00:00:00Z",
			EndTime:   "2026-01-01T12:00:00Z",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject ip_allocation_policy with both cluster range name and CIDR", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			IpAllocationPolicy: &GcpCloudComposerIpAllocationPolicy{
				ClusterSecondaryRangeName: "composer-pods",
				ClusterIpv4CidrBlock:      "10.4.0.0/14",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject ip_allocation_policy with both services range name and CIDR", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			IpAllocationPolicy: &GcpCloudComposerIpAllocationPolicy{
				ServicesSecondaryRangeName: "composer-services",
				ServicesIpv4CidrBlock:      "10.8.0.0/20",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject master_authorized_networks cidr_block missing its CIDR", func() {
		msg := minimal()
		msg.Spec.MasterAuthorizedNetworksConfig = &GcpCloudComposerMasterAuthorizedNetworksConfig{
			Enabled: true,
			CidrBlocks: []*GcpCloudComposerCidrBlock{
				{DisplayName: "missing-cidr"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid task_logs_storage_mode", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			TaskLogsStorageMode: "LOCAL_DISK",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid airflow_metadata_retention_mode", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionMode: "RETENTION_FOREVER",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject airflow_metadata_retention_days below the 30-day floor", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 29,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject airflow_metadata_retention_days above the 730-day ceiling", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionMode: "RETENTION_MODE_ENABLED",
			AirflowMetadataRetentionDays: 731,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject airflow_metadata_retention_days without a retention mode", func() {
		msg := minimal()
		msg.Spec.DataRetentionConfig = &GcpCloudComposerDataRetentionConfig{
			AirflowMetadataRetentionDays: 90,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("airflow_metadata_retention_mode"))
	})

	ginkgo.It("should reject allowed_ip_range missing value", func() {
		msg := minimal()
		msg.Spec.WebServerNetworkAccessControl = &GcpCloudComposerWebServerAccessControl{
			AllowedIpRanges: []*GcpCloudComposerAllowedIpRange{
				{Description: "Missing CIDR"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept deletion_policy DELETE, PREVENT, ABANDON, and empty", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "policy %q", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Secret environment variables ────────────────

	withSecrets := func() *GcpCloudComposerEnvironment {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpCloudComposerSoftwareConfig{
			EnvVariables:       map[string]string{"MODE": "prod"},
			SecretEnvVariables: map[string]string{"WAREHOUSE_PASSWORD": "pw"},
		}
		return msg
	}

	ginkgo.It("should accept secret env variables beside literals, with or without a node service account", func() {
		msg := withSecrets()
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{ServiceAccount: svr("projects/my-gcp-project/serviceAccounts/airflow@my-gcp-project.iam.gserviceaccount.com")}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a variable in both env_variables and secret_env_variables", func() {
		msg := withSecrets()
		msg.Spec.SoftwareConfig.EnvVariables["WAREHOUSE_PASSWORD"] = "literal"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("never both"))
	})

	ginkgo.It("should reject secret env variables when the node service account is not named by email", func() {
		msg := withSecrets()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{ServiceAccount: svr("projects/my-gcp-project/serviceAccounts/104510491012345678901")}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("must name the account by email"))
		msg.Spec.SoftwareConfig.SecretEnvVariables = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})
})
