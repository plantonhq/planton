package azurecontainerappv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func boolPtr(b bool) *bool { return &b }

func int32Ptr(i int32) *int32 { return &i }

// minimalSpec returns the smallest valid app: one container inside an
// environment.
func minimalSpec() *AzureContainerApp {
	return &AzureContainerApp{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerApp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-app",
		},
		Spec: &AzureContainerAppSpec{
			ResourceGroup:             literal("my-rg"),
			ContainerAppName:          "my-app",
			ContainerAppEnvironmentId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env"),
			Containers: []*AzureContainerAppContainer{{
				Name:   "web",
				Image:  "mcr.microsoft.com/k8se/quickstart:latest",
				Cpu:    0.25,
				Memory: "0.5Gi",
			}},
		},
	}
}

func validIngress() *AzureContainerAppIngress {
	return &AzureContainerAppIngress{
		ExternalEnabled: boolPtr(true),
		TargetPort:      8080,
		TrafficWeight: []*AzureContainerAppTrafficWeight{{
			LatestRevision: boolPtr(true),
			Percentage:     100,
		}},
	}
}

var _ = ginkgo.Describe("AzureContainerAppSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal app", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("accepts ingress with a latest-revision traffic weight", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts MULTIPLE revision mode with suffix-targeted weights", func() {
			input := minimalSpec()
			input.Spec.RevisionMode = AzureContainerAppRevisionMode_MULTIPLE
			input.Spec.RevisionSuffix = "v2"
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.TrafficWeight = []*AzureContainerAppTrafficWeight{
				{RevisionSuffix: "v1", Percentage: 80},
				{RevisionSuffix: "v2", Percentage: 20, Label: "canary"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts TCP transport with an exposed port", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.Transport = AzureContainerAppIngressTransport_TCP
			input.Spec.Ingress.ExposedPort = int32Ptr(6379)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts CORS and client certificate mode on ingress", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.ClientCertificateMode = AzureContainerAppIngressClientCertificateMode_REQUIRE
			input.Spec.Ingress.Cors = &AzureContainerAppCors{
				AllowedOrigins:          []string{"https://example.com"},
				AllowedMethods:          []string{"GET", "POST"},
				AllowCredentialsEnabled: boolPtr(true),
				MaxAgeInSeconds:         int32Ptr(3600),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts IP security restrictions", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.IpSecurityRestrictions = []*AzureContainerAppIpSecurityRestriction{{
				Name:           "office",
				Action:         AzureContainerAppIpRestrictionAction_ALLOW,
				IpAddressRange: "203.0.113.0/24",
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts all three probe types with type-legal thresholds", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_HTTP_GET,
				Port:                  8080,
				Path:                  "/healthz",
				FailureCountThreshold: int32Ptr(30),
			}
			input.Spec.Containers[0].ReadinessProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_TCP_SOCKET,
				Port:                  8080,
				FailureCountThreshold: int32Ptr(48),
				SuccessCountThreshold: int32Ptr(3),
			}
			input.Spec.Containers[0].StartupProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_HTTPS_GET,
				Port:                  8443,
				Path:                  "/started",
				FailureCountThreshold: int32Ptr(240),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts init containers without resources", func() {
			input := minimalSpec()
			input.Spec.InitContainers = []*AzureContainerAppInitContainer{{
				Name:    "migrate",
				Image:   "myregistry.azurecr.io/migrate:v1",
				Command: []string{"/bin/migrate"},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every volume storage type with its storage_name contract", func() {
			input := minimalSpec()
			input.Spec.Volumes = []*AzureContainerAppVolume{
				{Name: "scratch"},
				{Name: "cache", StorageType: AzureContainerAppVolumeStorageType_EMPTY_DIR},
				{Name: "shared-smb", StorageType: AzureContainerAppVolumeStorageType_AZURE_FILE, StorageName: literal("envstorage"), MountOptions: "uid=1000"},
				{Name: "shared-nfs", StorageType: AzureContainerAppVolumeStorageType_NFS_AZURE_FILE, StorageName: literal("envstorage-nfs")},
				{Name: "certs", StorageType: AzureContainerAppVolumeStorageType_SECRET},
			}
			input.Spec.Containers[0].VolumeMounts = []*AzureContainerAppVolumeMount{{
				Name: "shared-smb",
				Path: "/data",
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a plain-text secret and a Key Vault secret with identity", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{
				{Name: "api-key", Value: "s3cret"},
				{
					Name:             "db-password",
					KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password",
					Identity:         "System",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts registry auth via identity and via username/password", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{Name: "reg-password", Value: "s3cret"}}
			input.Spec.Registries = []*AzureContainerAppRegistry{
				{Server: literal("myregistry.azurecr.io"), Identity: "System"},
				{Server: literal("docker.io"), Username: "bot", PasswordSecretName: "reg-password"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts env vars as literals and secret references", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{Name: "api-key", Value: "s3cret"}}
			input.Spec.Containers[0].Env = []*AzureContainerAppEnvVar{
				{Name: "MODE", Value: "production"},
				{Name: "API_KEY", SecretName: "api-key"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts all four scale rule families", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{Name: "queue-conn", Value: "conn-string"}}
			input.Spec.HttpScaleRules = []*AzureContainerAppHttpScaleRule{{Name: "http", ConcurrentRequests: "100"}}
			input.Spec.TcpScaleRules = []*AzureContainerAppTcpScaleRule{{Name: "tcp", ConcurrentRequests: "50"}}
			input.Spec.AzureQueueScaleRules = []*AzureContainerAppAzureQueueScaleRule{{
				Name:        "queue",
				QueueName:   "work",
				QueueLength: 5,
				Authentication: []*AzureContainerAppScaleRuleAuth{{
					SecretName:       "queue-conn",
					TriggerParameter: "connection",
				}},
			}}
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "cron-scale",
				CustomRuleType: "cron",
				Metadata: map[string]string{
					"timezone":        "UTC",
					"start":           "0 8 * * 1-5",
					"end":             "0 18 * * 1-5",
					"desiredReplicas": "5",
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom scale rule executing under a managed identity", func() {
			input := minimalSpec()
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "sb-scale",
				CustomRuleType: "azure-servicebus",
				Metadata:       map[string]string{"queueName": "work", "messageCount": "10"},
				IdentityId:     literal("System"),
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom scale rule identity referenced from a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "sb-scale",
				CustomRuleType: "azure-servicebus",
				Metadata:       map[string]string{"queueName": "work", "messageCount": "10"},
				IdentityId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{Name: "scaler-identity"},
					},
				},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts dapr, identity, tags, and scaling knobs", func() {
			input := minimalSpec()
			input.Spec.Dapr = &AzureContainerAppDapr{
				AppId:       "orders",
				AppPort:     int32Ptr(8080),
				AppProtocol: AzureContainerAppDaprProtocol_DAPR_GRPC,
			}
			input.Spec.Identity = &AzureContainerAppIdentity{
				Type: AzureContainerAppIdentityType_SYSTEM_AND_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				},
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			input.Spec.MinReplicas = int32Ptr(1)
			input.Spec.MaxReplicas = int32Ptr(30)
			input.Spec.CooldownPeriodInSeconds = int32Ptr(120)
			input.Spec.PollingIntervalInSeconds = int32Ptr(15)
			input.Spec.TerminationGracePeriodSeconds = int32Ptr(30)
			input.Spec.MaxInactiveRevisions = int32Ptr(10)
			input.Spec.WorkloadProfileName = "dedicated-d4"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects an app name with uppercase characters", func() {
			input := minimalSpec()
			input.Spec.ContainerAppName = "MyApp"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app name with consecutive hyphens", func() {
			input := minimalSpec()
			input.Spec.ContainerAppName = "my--app"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app name longer than 32 characters", func() {
			input := minimalSpec()
			input.Spec.ContainerAppName = "a12345678901234567890123456789012"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing environment reference", func() {
			input := minimalSpec()
			input.Spec.ContainerAppEnvironmentId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an app without containers", func() {
			input := minimalSpec()
			input.Spec.Containers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a container with cpu below 0.1", func() {
			input := minimalSpec()
			input.Spec.Containers[0].Cpu = 0.05
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a container without memory", func() {
			input := minimalSpec()
			input.Spec.Containers[0].Memory = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a probe with an unspecified transport", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{Port: 8080}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a liveness probe with success_count_threshold", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_TCP_SOCKET,
				Port:                  8080,
				SuccessCountThreshold: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a liveness probe with failure_count_threshold above 30", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_TCP_SOCKET,
				Port:                  8080,
				FailureCountThreshold: int32Ptr(31),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a readiness probe with failure_count_threshold above 48", func() {
			input := minimalSpec()
			input.Spec.Containers[0].ReadinessProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_TCP_SOCKET,
				Port:                  8080,
				FailureCountThreshold: int32Ptr(49),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a startup probe with success_count_threshold", func() {
			input := minimalSpec()
			input.Spec.Containers[0].StartupProbe = &AzureContainerAppProbe{
				Transport:             AzureContainerAppProbeTransport_TCP_SOCKET,
				Port:                  8080,
				SuccessCountThreshold: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an env var carrying both a value and a secret reference", func() {
			input := minimalSpec()
			input.Spec.Containers[0].Env = []*AzureContainerAppEnvVar{{
				Name:       "API_KEY",
				Value:      "literal",
				SecretName: "api-key",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an AZURE_FILE volume without storage_name", func() {
			input := minimalSpec()
			input.Spec.Volumes = []*AzureContainerAppVolume{{
				Name:        "shared",
				StorageType: AzureContainerAppVolumeStorageType_AZURE_FILE,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an EMPTY_DIR volume with storage_name", func() {
			input := minimalSpec()
			input.Spec.Volumes = []*AzureContainerAppVolume{{
				Name:        "scratch",
				StorageType: AzureContainerAppVolumeStorageType_EMPTY_DIR,
				StorageName: literal("envstorage"),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a secret carrying both a value and a Key Vault reference", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{
				Name:             "db-password",
				Value:            "literal",
				KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password",
				Identity:         "System",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a Key Vault secret without an identity", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{
				Name:             "db-password",
				KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a secret identity without a Key Vault reference", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppSecret{{
				Name:     "db-password",
				Value:    "literal",
				Identity: "System",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a registry mixing identity and username/password auth", func() {
			input := minimalSpec()
			input.Spec.Registries = []*AzureContainerAppRegistry{{
				Server:             literal("myregistry.azurecr.io"),
				Identity:           "System",
				Username:           "bot",
				PasswordSecretName: "reg-password",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a registry username without a password secret", func() {
			input := minimalSpec()
			input.Spec.Registries = []*AzureContainerAppRegistry{{
				Server:   literal("docker.io"),
				Username: "bot",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a registry with no authentication at all", func() {
			input := minimalSpec()
			input.Spec.Registries = []*AzureContainerAppRegistry{{
				Server: literal("docker.io"),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects ingress without traffic weights", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.TrafficWeight = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a traffic weight targeting both latest and a suffix", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.TrafficWeight = []*AzureContainerAppTrafficWeight{{
				LatestRevision: boolPtr(true),
				RevisionSuffix: "v1",
				Percentage:     100,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a traffic weight targeting neither latest nor a suffix", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.TrafficWeight = []*AzureContainerAppTrafficWeight{{
				Percentage: 100,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an exposed port without TCP transport", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.ExposedPort = int32Ptr(6379)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an IP restriction with an unspecified action", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.IpSecurityRestrictions = []*AzureContainerAppIpSecurityRestriction{{
				Name:           "office",
				IpAddressRange: "203.0.113.0/24",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects CORS without allowed origins", func() {
			input := minimalSpec()
			input.Spec.Ingress = validIngress()
			input.Spec.Ingress.Cors = &AzureContainerAppCors{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an HTTP scale rule with a non-numeric threshold", func() {
			input := minimalSpec()
			input.Spec.HttpScaleRules = []*AzureContainerAppHttpScaleRule{{
				Name:               "http",
				ConcurrentRequests: "lots",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an HTTP scale rule with a zero threshold", func() {
			input := minimalSpec()
			input.Spec.HttpScaleRules = []*AzureContainerAppHttpScaleRule{{
				Name:               "http",
				ConcurrentRequests: "0",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an Azure Queue scale rule without authentication", func() {
			input := minimalSpec()
			input.Spec.AzureQueueScaleRules = []*AzureContainerAppAzureQueueScaleRule{{
				Name:        "queue",
				QueueName:   "work",
				QueueLength: 5,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an Azure Queue scale rule auth entry without trigger_parameter", func() {
			input := minimalSpec()
			input.Spec.AzureQueueScaleRules = []*AzureContainerAppAzureQueueScaleRule{{
				Name:        "queue",
				QueueName:   "work",
				QueueLength: 5,
				Authentication: []*AzureContainerAppScaleRuleAuth{{
					SecretName: "queue-conn",
				}},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom scale rule with an unknown scaler type", func() {
			input := minimalSpec()
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "bad",
				CustomRuleType: "not-a-scaler",
				Metadata:       map[string]string{"key": "value"},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom scale rule without metadata", func() {
			input := minimalSpec()
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "cron-scale",
				CustomRuleType: "cron",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom scale rule auth entry without trigger_parameter", func() {
			input := minimalSpec()
			input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{{
				Name:           "kafka-scale",
				CustomRuleType: "kafka",
				Metadata:       map[string]string{"topic": "orders"},
				Authentication: []*AzureContainerAppScaleRuleAuth{{
					SecretName: "sasl-secret",
				}},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects USER_ASSIGNED identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppIdentity{
				Type: AzureContainerAppIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects min_replicas above 300 and max_replicas of 0", func() {
			input := minimalSpec()
			input.Spec.MinReplicas = int32Ptr(301)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input2 := minimalSpec()
			input2.Spec.MaxReplicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input2)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a dapr block without an app_id", func() {
			input := minimalSpec()
			input.Spec.Dapr = &AzureContainerAppDapr{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
