package azurecontainerappv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureContainerAppSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerAppSpec Validation Tests")
}

// helper to create a minimal valid spec
func minimalSpec() *AzureContainerApp {
	return &AzureContainerApp{
		ApiVersion: "azure.planton.dev/v1",
		Kind:       "AzureContainerApp",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-app",
		},
		Spec: &AzureContainerAppSpec{
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			Name: "my-container-app",
			ContainerAppEnvironmentId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.App/managedEnvironments/env",
				},
			},
			Containers: []*AzureContainerAppContainer{
				{
					Name:   "main",
					Image:  "mcr.microsoft.com/k8se/quickstart:latest",
					Cpu:    0.25,
					Memory: "0.5Gi",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureContainerAppSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_app", func() {

			ginkgo.It("should not return a validation error for a minimal spec", func() {
				input := minimalSpec()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ingress", func() {
				latestRev := true
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with secrets (plain value)", func() {
				input := minimalSpec()
				input.Spec.Secrets = []*AzureContainerAppSecret{
					{
						Name:  "my-secret",
						Value: "s3cr3t-v4lue",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with registries", func() {
				input := minimalSpec()
				input.Spec.Secrets = []*AzureContainerAppSecret{
					{
						Name:  "acr-password",
						Value: "p@ssword",
					},
				}
				input.Spec.Registries = []*AzureContainerAppRegistry{
					{
						Server:             "myregistry.azurecr.io",
						Username:           "myuser",
						PasswordSecretName: "acr-password",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with health probes (HTTP liveness probe)", func() {
				input := minimalSpec()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport: "HTTP",
					Port:      8080,
					Path:      "/healthz",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple containers", func() {
				input := minimalSpec()
				input.Spec.Containers = append(input.Spec.Containers, &AzureContainerAppContainer{
					Name:   "sidecar",
					Image:  "nginx:latest",
					Cpu:    0.25,
					Memory: "0.5Gi",
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with init containers (no cpu/memory)", func() {
				input := minimalSpec()
				input.Spec.InitContainers = []*AzureContainerAppInitContainer{
					{
						Name:    "db-migrate",
						Image:   "myregistry.azurecr.io/migrator:v1",
						Command: []string{"./migrate", "up"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP scale rules", func() {
				input := minimalSpec()
				input.Spec.HttpScaleRules = []*AzureContainerAppHttpScaleRule{
					{
						Name:               "http-rule",
						ConcurrentRequests: "100",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom scale rules (kafka)", func() {
				input := minimalSpec()
				input.Spec.CustomScaleRules = []*AzureContainerAppCustomScaleRule{
					{
						Name:           "kafka-scaler",
						CustomRuleType: "kafka",
						Metadata: map[string]string{
							"bootstrapServers": "kafka:9092",
							"consumerGroup":    "my-group",
							"topic":            "my-topic",
							"lagThreshold":     "100",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with dapr configuration", func() {
				input := minimalSpec()
				input.Spec.Dapr = &AzureContainerAppDapr{
					AppId: "my-dapr-app",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with volumes and volume mounts", func() {
				input := minimalSpec()
				input.Spec.Volumes = []*AzureContainerAppVolume{
					{
						Name: "scratch",
					},
				}
				input.Spec.Containers[0].VolumeMounts = []*AzureContainerAppVolumeMount{
					{
						Name: "scratch",
						Path: "/tmp/scratch",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with identity (SystemAssigned)", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureContainerAppIdentity{
					Type: "SystemAssigned",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with identity (UserAssigned with identity_ids)", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureContainerAppIdentity{
					Type: "UserAssigned",
					IdentityIds: []*foreignkeyv1.StringValueOrRef{
						{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/my-identity",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom for resource_group", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom for container_app_environment_id", func() {
				input := minimalSpec()
				input.Spec.ContainerAppEnvironmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AzureContainerAppEnvironment,
							Name:      "shared-env",
							FieldPath: "status.outputs.environment_id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields set", func() {
				latestRev := true
				revMode := "Single"
				minReplicas := int32(1)
				maxReplicas := int32(10)
				maxInactive := int32(50)
				cooldown := int32(60)
				polling := int32(15)
				termGrace := int32(30)
				daprPort := int32(3000)
				daprProto := "grpc"
				input := minimalSpec()
				input.Metadata.Org = "mycompany"
				input.Metadata.Env = "production"
				input.Spec.RevisionMode = &revMode
				input.Spec.WorkloadProfileName = "dedicated"
				input.Spec.MaxInactiveRevisions = &maxInactive
				input.Spec.MinReplicas = &minReplicas
				input.Spec.MaxReplicas = &maxReplicas
				input.Spec.CooldownPeriodInSeconds = &cooldown
				input.Spec.PollingIntervalInSeconds = &polling
				input.Spec.RevisionSuffix = "v1"
				input.Spec.TerminationGracePeriodSeconds = &termGrace
				input.Spec.Secrets = []*AzureContainerAppSecret{
					{
						Name:  "db-password",
						Value: "s3cr3t",
					},
				}
				input.Spec.Registries = []*AzureContainerAppRegistry{
					{
						Server:   "myregistry.azurecr.io",
						Identity: "System",
					},
				}
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
				}
				input.Spec.Dapr = &AzureContainerAppDapr{
					AppId:       "my-app",
					AppPort:     &daprPort,
					AppProtocol: &daprProto,
				}
				input.Spec.Identity = &AzureContainerAppIdentity{
					Type: "SystemAssigned",
				}
				input.Spec.Volumes = []*AzureContainerAppVolume{
					{
						Name: "data",
					},
				}
				input.Spec.InitContainers = []*AzureContainerAppInitContainer{
					{
						Name:  "init",
						Image: "busybox:latest",
					},
				}
				input.Spec.HttpScaleRules = []*AzureContainerAppHttpScaleRule{
					{
						Name:               "http-rule",
						ConcurrentRequests: "50",
					},
				}
				input.Spec.Containers[0].Env = []*AzureContainerAppEnvVar{
					{
						Name:  "DB_HOST",
						Value: "mydb.postgres.database.azure.com",
					},
					{
						Name:       "DB_PASSWORD",
						SecretName: "db-password",
					},
				}
				input.Spec.Containers[0].VolumeMounts = []*AzureContainerAppVolumeMount{
					{
						Name: "data",
						Path: "/data",
					},
				}
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport: "HTTP",
					Port:      8080,
					Path:      "/healthz",
				}
				input.Spec.Containers[0].ReadinessProbe = &AzureContainerAppProbe{
					Transport: "HTTP",
					Port:      8080,
					Path:      "/ready",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with revision_mode Multiple and traffic weight split", func() {
				revMode := "Multiple"
				latestRev := true
				latestRevFalse := false
				input := minimalSpec()
				input.Spec.RevisionMode = &revMode
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     80,
							Label:          "latest",
						},
						{
							LatestRevision: &latestRevFalse,
							RevisionSuffix: "canary-v1",
							Percentage:     20,
							Label:          "canary",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IP security restrictions on ingress", func() {
				latestRev := true
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
					IpSecurityRestrictions: []*AzureContainerAppIpSecurityRestriction{
						{
							Name:           "allow-office",
							Action:         "Allow",
							IpAddressRange: "203.0.113.0/24",
							Description:    "Office network",
						},
						{
							Name:           "deny-all",
							Action:         "Deny",
							IpAddressRange: "0.0.0.0/0",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with CORS policy on ingress", func() {
				latestRev := true
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
					CorsPolicy: &AzureContainerAppCorsPolicy{
						AllowedOrigins:          []string{"https://example.com", "https://*.contoso.com"},
						AllowedHeaders:          []string{"Content-Type", "Authorization"},
						AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE"},
						ExposedHeaders:          []string{"X-Custom-Header"},
						AllowCredentialsEnabled: boolPtr(true),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with init container with explicit cpu/memory", func() {
				initCpu := 0.25
				initMem := "0.5Gi"
				input := minimalSpec()
				input.Spec.InitContainers = []*AzureContainerAppInitContainer{
					{
						Name:   "db-migrate",
						Image:  "myregistry.azurecr.io/migrator:v1",
						Cpu:    &initCpu,
						Memory: &initMem,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_app", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := minimalSpec()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := minimalSpec()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short (1 char)", func() {
				input := minimalSpec()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too long (33+ chars)", func() {
				tooLong := "a"
				for len(tooLong) < 33 {
					tooLong += "b"
				}
				input := minimalSpec()
				input.Spec.Name = tooLong
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a number", func() {
				input := minimalSpec()
				input.Spec.Name = "1bad-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name starts with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "-bad-name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name ends with a hyphen", func() {
				input := minimalSpec()
				input.Spec.Name = "bad-name-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains uppercase letters", func() {
				input := minimalSpec()
				input.Spec.Name = "Bad-Name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name contains consecutive hyphens", func() {
				input := minimalSpec()
				input.Spec.Name = "bad--name"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container_app_environment_id is missing", func() {
				input := minimalSpec()
				input.Spec.ContainerAppEnvironmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when containers is empty", func() {
				input := minimalSpec()
				input.Spec.Containers = []*AzureContainerAppContainer{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container name is missing", func() {
				input := minimalSpec()
				input.Spec.Containers[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container image is missing", func() {
				input := minimalSpec()
				input.Spec.Containers[0].Image = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container cpu is below 0.1", func() {
				input := minimalSpec()
				input.Spec.Containers[0].Cpu = 0.05
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container memory is missing", func() {
				input := minimalSpec()
				input.Spec.Containers[0].Memory = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when revision_mode is invalid", func() {
				revMode := "Invalid"
				input := minimalSpec()
				input.Spec.RevisionMode = &revMode
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe transport is invalid", func() {
				input := minimalSpec()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport: "GRPC",
					Port:      8080,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe port is 0", func() {
				input := minimalSpec()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport: "TCP",
					Port:      0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when probe port is 65536", func() {
				input := minimalSpec()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport: "TCP",
					Port:      65536,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when ingress transport is invalid", func() {
				latestRev := true
				transport := "websocket"
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					Transport:       &transport,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalSpec()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureContainerApp{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureContainerApp",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := minimalSpec()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := minimalSpec()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when traffic weight percentage is out of range (101)", func() {
				latestRev := true
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     101,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when secret name contains uppercase", func() {
				input := minimalSpec()
				input.Spec.Secrets = []*AzureContainerAppSecret{
					{
						Name:  "My-Secret",
						Value: "value",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when IP restriction action is invalid", func() {
				latestRev := true
				input := minimalSpec()
				input.Spec.Ingress = &AzureContainerAppIngress{
					ExternalEnabled: boolPtr(true),
					TargetPort:      8080,
					TrafficWeight: []*AzureContainerAppTrafficWeight{
						{
							LatestRevision: &latestRev,
							Percentage:     100,
						},
					},
					IpSecurityRestrictions: []*AzureContainerAppIpSecurityRestriction{
						{
							Name:           "bad-rule",
							Action:         "Block",
							IpAddressRange: "10.0.0.0/8",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when identity type is invalid", func() {
				input := minimalSpec()
				input.Spec.Identity = &AzureContainerAppIdentity{
					Type: "ManagedIdentity",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when dapr app_protocol is invalid", func() {
				daprProto := "websocket"
				input := minimalSpec()
				input.Spec.Dapr = &AzureContainerAppDapr{
					AppId:       "my-app",
					AppProtocol: &daprProto,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume storage_type is invalid", func() {
				storageType := "NFS"
				input := minimalSpec()
				input.Spec.Volumes = []*AzureContainerAppVolume{
					{
						Name:        "bad-vol",
						StorageType: &storageType,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when min_replicas is negative (-1)", func() {
				minReplicas := int32(-1)
				input := minimalSpec()
				input.Spec.MinReplicas = &minReplicas
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_replicas is zero (0)", func() {
				maxReplicas := int32(0)
				input := minimalSpec()
				input.Spec.MaxReplicas = &maxReplicas
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when failure_count_threshold is out of range (31)", func() {
				failCount := int32(31)
				input := minimalSpec()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerAppProbe{
					Transport:             "TCP",
					Port:                  8080,
					FailureCountThreshold: &failCount,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// boolPtr is a helper to create a pointer to a bool value.
func boolPtr(v bool) *bool {
	return &v
}
