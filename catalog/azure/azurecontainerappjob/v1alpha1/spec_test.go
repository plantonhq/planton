package azurecontainerappjobv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppJobSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppJobSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(i int32) *int32 { return &i }

// minimalSpec returns the smallest valid job: one container, manual trigger.
func minimalSpec() *AzureContainerAppJob {
	return &AzureContainerAppJob{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerAppJob",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-job",
		},
		Spec: &AzureContainerAppJobSpec{
			Region:                    "eastus",
			ResourceGroup:             literal("my-rg"),
			JobName:                   "my-job",
			ContainerAppEnvironmentId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env"),
			ReplicaTimeoutInSeconds:   300,
			Containers: []*AzureContainerAppJobContainer{{
				Name:   "worker",
				Image:  "mcr.microsoft.com/k8se/quickstart-jobs:latest",
				Cpu:    0.25,
				Memory: "0.5Gi",
			}},
			ManualTrigger: &AzureContainerAppJobManualTrigger{},
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppJobSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal manually triggered job", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a scheduled job with parallelism", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.ScheduleTrigger = &AzureContainerAppJobScheduleTrigger{
				CronExpression:         "0 2 * * *",
				Parallelism:            int32Ptr(3),
				ReplicaCompletionCount: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an event-triggered job with a scale rule", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.Secrets = []*AzureContainerAppJobSecret{{Name: "queue-conn", Value: "conn-string"}}
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					MaxExecutions:            int32Ptr(20),
					MinExecutions:            int32Ptr(0),
					PollingIntervalInSeconds: int32Ptr(60),
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "queue-depth",
						CustomRuleType: "azure-queue",
						Metadata:       map[string]string{"queueName": "work", "queueLength": "5"},
						Authentication: []*AzureContainerAppJobScaleRuleAuth{{
							SecretName:       "queue-conn",
							TriggerParameter: "connection",
						}},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an event scale rule executing under a managed identity", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "sb-depth",
						CustomRuleType: "azure-servicebus",
						Metadata:       map[string]string{"queueName": "work", "messageCount": "10"},
						IdentityId:     literal("System"),
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an event scale rule identity referenced from a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "sb-depth",
						CustomRuleType: "azure-servicebus",
						Metadata:       map[string]string{"queueName": "work", "messageCount": "10"},
						IdentityId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{Name: "scaler-identity"},
							},
						},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts retry limit, workload profile, volumes, probes, and tags", func() {
			input := minimalSpec()
			input.Spec.ReplicaRetryLimit = int32Ptr(2)
			input.Spec.WorkloadProfileName = "dedicated-d4"
			input.Spec.Volumes = []*AzureContainerAppJobVolume{{
				Name:        "work",
				StorageType: AzureContainerAppJobVolumeStorageType_AZURE_FILE,
				StorageName: literal("envstorage"),
			}}
			input.Spec.Containers[0].VolumeMounts = []*AzureContainerAppJobVolumeMount{{
				Name: "work",
				Path: "/work",
			}}
			input.Spec.Containers[0].StartupProbe = &AzureContainerAppJobProbe{
				Transport:             AzureContainerAppJobProbeTransport_TCP_SOCKET,
				Port:                  8080,
				FailureCountThreshold: int32Ptr(240),
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts secrets, registries, and identity", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppJobSecret{
				{Name: "api-key", Value: "s3cret"},
				{Name: "db-password", KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password", Identity: "System"},
			}
			input.Spec.Registries = []*AzureContainerAppJobRegistry{
				{Server: literal("myregistry.azurecr.io"), Identity: "System"},
			}
			input.Spec.Identity = &AzureContainerAppJobIdentity{
				Type: AzureContainerAppJobIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a job with no trigger", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a job with two triggers", func() {
			input := minimalSpec()
			input.Spec.ScheduleTrigger = &AzureContainerAppJobScheduleTrigger{CronExpression: "0 2 * * *"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a schedule trigger without a cron expression", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.ScheduleTrigger = &AzureContainerAppJobScheduleTrigger{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing replica timeout", func() {
			input := minimalSpec()
			input.Spec.ReplicaTimeoutInSeconds = 0
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a job name with dots", func() {
			input := minimalSpec()
			input.Spec.JobName = "my.job"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a job name with consecutive hyphens", func() {
			input := minimalSpec()
			input.Spec.JobName = "my--job"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a job name longer than 32 characters", func() {
			input := minimalSpec()
			input.Spec.JobName = "a12345678901234567890123456789012"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a job without containers", func() {
			input := minimalSpec()
			input.Spec.Containers = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an event scale rule without metadata", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "queue-depth",
						CustomRuleType: "azure-queue",
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an event scale rule with a scaler type outside the KEDA vocabulary", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "queue-depth",
						CustomRuleType: "not-a-keda-scaler",
						Metadata:       map[string]string{"queueName": "work"},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an event scale rule auth entry without trigger_parameter", func() {
			input := minimalSpec()
			input.Spec.ManualTrigger = nil
			input.Spec.EventTrigger = &AzureContainerAppJobEventTrigger{
				Scale: &AzureContainerAppJobEventScale{
					Rules: []*AzureContainerAppJobEventScaleRule{{
						Name:           "queue-depth",
						CustomRuleType: "azure-queue",
						Metadata:       map[string]string{"queueName": "work"},
						Authentication: []*AzureContainerAppJobScaleRuleAuth{{
							SecretName: "queue-conn",
						}},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a liveness probe with success_count_threshold", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppJobProbe{
				Transport:             AzureContainerAppJobProbeTransport_TCP_SOCKET,
				Port:                  8080,
				SuccessCountThreshold: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a liveness probe with failure_count_threshold above 30", func() {
			input := minimalSpec()
			input.Spec.Containers[0].LivenessProbe = &AzureContainerAppJobProbe{
				Transport:             AzureContainerAppJobProbeTransport_TCP_SOCKET,
				Port:                  8080,
				FailureCountThreshold: int32Ptr(31),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an AZURE_FILE volume without storage_name", func() {
			input := minimalSpec()
			input.Spec.Volumes = []*AzureContainerAppJobVolume{{
				Name:        "work",
				StorageType: AzureContainerAppJobVolumeStorageType_AZURE_FILE,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a secret carrying both a value and a Key Vault reference", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppJobSecret{{
				Name:             "db-password",
				Value:            "literal",
				KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password",
				Identity:         "System",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a Key Vault secret without an identity", func() {
			input := minimalSpec()
			input.Spec.Secrets = []*AzureContainerAppJobSecret{{
				Name:             "db-password",
				KeyVaultSecretId: "https://vault.vault.azure.net/secrets/db-password",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a registry mixing identity and username/password auth", func() {
			input := minimalSpec()
			input.Spec.Registries = []*AzureContainerAppJobRegistry{{
				Server:             literal("myregistry.azurecr.io"),
				Identity:           "System",
				Username:           "bot",
				PasswordSecretName: "reg-password",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an env var carrying both a value and a secret reference", func() {
			input := minimalSpec()
			input.Spec.Containers[0].Env = []*AzureContainerAppJobEnvVar{{
				Name:       "API_KEY",
				Value:      "literal",
				SecretName: "api-key",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects USER_ASSIGNED identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureContainerAppJobIdentity{
				Type: AzureContainerAppJobIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
