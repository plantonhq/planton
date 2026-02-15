package gcpcloudcomposerenvironmentv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
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
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpCloudComposerEnvironment",
			Metadata: &shared.CloudResourceMetadata{
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
			Network:    svr("projects/my-project/global/networks/my-vpc"),
			Subnetwork: svr("projects/my-project/regions/us-central1/subnetworks/my-subnet"),
			ServiceAccount: svr("composer-sa@my-project.iam.gserviceaccount.com"),
			Tags:       []string{"composer", "airflow"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node_config with Composer 3 PSC networking", func() {
		msg := minimal()
		msg.Spec.NodeConfig = &GcpCloudComposerNodeConfig{
			ComposerNetworkAttachment:      "projects/my-project/regions/us-central1/networkAttachments/my-attachment",
			ComposerInternalIpv4CidrBlock: "10.0.0.0/20",
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

	ginkgo.It("should accept private_environment_config with VPC_PEERING", func() {
		msg := minimal()
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			EnablePrivateEndpoint:                  true,
			ConnectionType:                         "VPC_PEERING",
			MasterIpv4CidrBlock:                    "172.16.0.0/28",
			CloudSqlIpv4CidrBlock:                  "10.0.32.0/20",
			CloudComposerNetworkIpv4CidrBlock:      "10.0.48.0/20",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept private_environment_config with PRIVATE_SERVICE_CONNECT", func() {
		msg := minimal()
		msg.Spec.PrivateEnvironmentConfig = &GcpCloudComposerPrivateEnvironmentConfig{
			ConnectionType:                         "PRIVATE_SERVICE_CONNECT",
			CloudComposerConnectionSubnetwork:      "projects/my-project/regions/us-central1/subnetworks/psc-subnet",
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
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing region", func() {
		msg := minimal()
		msg.Spec.Region = ""
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
})
