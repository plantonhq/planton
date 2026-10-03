package azurecontainerinstancev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureContainerInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerInstanceSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testSubnetId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/aci-subnet"
	testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-mi"
)

// validContainer returns the smallest valid main container.
func validContainer() *AzureContainerInstanceContainer {
	return &AzureContainerInstanceContainer{
		Name:   "app",
		Image:  "mcr.microsoft.com/azuredocs/aci-helloworld:latest",
		Cpu:    0.5,
		Memory: 1.5,
	}
}

// validResource returns a minimal valid container group that
// individual cases mutate into the shape under test.
func validResource() *AzureContainerInstance {
	return &AzureContainerInstance{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureContainerInstance",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-container-group",
		},
		Spec: &AzureContainerInstanceSpec{
			ResourceGroup: literal("app-rg"),
			Name:          "app-aci",
			Region:        "eastus",
			OsType:        "Linux",
			Containers:    []*AzureContainerInstanceContainer{validContainer()},
		},
	}
}

// volumeOf returns a volume with the common fields set and no form --
// cases attach exactly the form under test.
func volumeOf() *AzureContainerInstanceVolume {
	return &AzureContainerInstanceVolume{
		Name:      "data",
		MountPath: "/data",
	}
}

// validAzureFile returns a complete Azure File volume form.
func validAzureFile() *AzureContainerInstanceVolumeAzureFile {
	return &AzureContainerInstanceVolumeAzureFile{
		ShareName:          literal("app-share"),
		StorageAccountName: literal("appstorage"),
		StorageAccountKey:  literal("c3VwZXItc2VjcmV0LWtleQ=="),
	}
}

var _ = ginkgo.Describe("AzureContainerInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_instance", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				gomega.Expect(protovalidate.Validate(validResource())).To(gomega.BeNil())
			})

			ginkgo.It("should accept every restart policy, sku, and priority token", func() {
				input := validResource()
				for _, policy := range []string{"", "Always", "Never", "OnFailure"} {
					input.Spec.RestartPolicy = policy
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected restart policy %q to be accepted", policy)
				}
				input = validResource()
				for _, sku := range []string{"", "Standard", "Dedicated", "Confidential", "NotSpecified"} {
					input.Spec.Sku = sku
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected sku %q to be accepted", sku)
				}
				input = validResource()
				input.Spec.Priority = "Regular"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Spot priority when the group has no IP", func() {
				input := validResource()
				input.Spec.Priority = "Spot"
				input.Spec.IpAddressType = "None"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a public group with a DNS label and reuse policy", func() {
				input := validResource()
				input.Spec.IpAddressType = "Public"
				input.Spec.DnsNameLabel = "acme-app"
				input.Spec.DnsNameLabelReusePolicy = "TenantReuse"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private group joined to a subnet", func() {
				input := validResource()
				input.Spec.IpAddressType = "Private"
				input.Spec.SubnetId = literal(testSubnetId)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept exposed ports that match a container's ports, including the TCP-default seam", func() {
				input := validResource()
				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: 80}}
				input.Spec.ExposedPorts = []*AzureContainerInstancePort{{Port: 80, Protocol: "TCP"}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: 53, Protocol: "UDP"}}
				input.Spec.ExposedPorts = []*AzureContainerInstancePort{{Port: 53, Protocol: "UDP"}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept each volume form alone", func() {
				azureFile := volumeOf()
				azureFile.AzureFile = validAzureFile()

				emptyDir := volumeOf()
				emptyDir.EmptyDir = true

				gitRepo := volumeOf()
				gitRepo.GitRepo = &AzureContainerInstanceVolumeGitRepo{Url: "https://github.com/acme/config", Revision: "main"}

				secret := volumeOf()
				secret.Secret = map[string]string{"config.json": "eyJrZXkiOiJ2YWx1ZSJ9"}

				for _, volume := range []*AzureContainerInstanceVolume{azureFile, emptyDir, gitRepo, secret} {
					input := validResource()
					input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{volume}
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected volume %q form to be accepted", volume.Name)
				}
			})

			ginkgo.It("should accept an init container seeding a shared scratch volume", func() {
				input := validResource()
				seeded := volumeOf()
				seeded.EmptyDir = true
				input.Spec.InitContainers = []*AzureContainerInstanceInitContainer{{
					Name:                       "seed",
					Image:                      "busybox:1.36",
					Commands:                   []string{"sh", "-c", "echo ready > /data/ready"},
					SecureEnvironmentVariables: map[string]string{"SEED_TOKEN": "hunter2"},
					Volumes:                    []*AzureContainerInstanceVolume{seeded},
				}}
				input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{seeded}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every identity flavor with matching identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureContainerInstanceIdentity{Type: AzureContainerInstanceIdentityType_SYSTEM_ASSIGNED}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

				input.Spec.Identity = &AzureContainerInstanceIdentity{
					Type:        AzureContainerInstanceIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

				input.Spec.Identity.Type = AzureContainerInstanceIdentityType_SYSTEM_AND_USER_ASSIGNED
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Log Analytics diagnostics with a log type and metadata", func() {
				input := validResource()
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceId:  literal("00000000-0000-0000-0000-000000000000"),
					WorkspaceKey: literal("d29ya3NwYWNlLWtleQ=="),
					LogType:      "ContainerInsights",
					// Keys come from ARM's closed vocabulary (pod-uuid,
					// cluster-resource-id, node-name) -- see the
					// closed-vocabulary rejection test below.
					Metadata: map[string]string{"node-name": "aci-connector"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept liveness and readiness probes in both forms", func() {
				input := validResource()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerInstanceProbe{
					Exec:             []string{"cat", "/tmp/healthy"},
					PeriodSeconds:    10,
					FailureThreshold: 3,
				}
				input.Spec.Containers[0].ReadinessProbe = &AzureContainerInstanceProbe{
					HttpGet: &AzureContainerInstanceProbeHttpGet{
						Path:        "/healthz",
						Port:        8080,
						Scheme:      "https",
						HttpHeaders: map[string]string{"X-Probe": "planton"},
					},
					InitialDelaySeconds: 5,
					TimeoutSeconds:      2,
					SuccessThreshold:    1,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept both registry credential forms", func() {
				input := validResource()
				input.Spec.ImageRegistryCredentials = []*AzureContainerInstanceImageRegistryCredential{
					{Server: literal("acme.azurecr.io"), UserAssignedIdentityId: literal(testIdentityId)},
					{Server: literal("registry.example.com"), Username: "puller", Password: "hunter2"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept custom DNS, zones, and the CMK pair", func() {
				input := validResource()
				input.Spec.DnsConfig = &AzureContainerInstanceDnsConfig{
					Nameservers:   []string{"10.0.0.10"},
					SearchDomains: []string{"internal.acme.com"},
					Options:       []string{"ndots:2"},
				}
				input.Spec.Zones = []string{"1"}
				input.Spec.KeyVaultKeyId = literal("https://vault.vault.azure.net/keys/aci-cmk/0123456789abcdef0123456789abcdef")
				input.Spec.KeyVaultUserAssignedIdentityId = literal(testIdentityId)
				input.Spec.Identity = &AzureContainerInstanceIdentity{
					Type:        AzureContainerInstanceIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept resource limits above the requests", func() {
				input := validResource()
				input.Spec.Containers[0].CpuLimit = proto.Float64(1)
				input.Spec.Containers[0].MemoryLimit = proto.Float64(2)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_instance", func() {

			ginkgo.It("should reject a missing resource group, name, region, or os type", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.OsType = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty containers list", func() {
				input := validResource()
				input.Spec.Containers = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject tokens outside the provider vocabularies", func() {
				input := validResource()
				input.Spec.OsType = "linux"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.RestartPolicy = "always"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Sku = "Basic"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Priority = "Low"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.IpAddressType = "public"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.DnsNameLabelReusePolicy = "AnyReuse"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Spot priority with a group IP (including the unset-means-Public default)", func() {
				input := validResource()
				input.Spec.Priority = "Spot"
				input.Spec.IpAddressType = "Public"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.IpAddressType = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a subnet together with a DNS label", func() {
				input := validResource()
				input.Spec.IpAddressType = "Private"
				input.Spec.SubnetId = literal(testSubnetId)
				input.Spec.DnsNameLabel = "acme-app"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a DNS label, reuse policy, or exposed ports on an IP-less group", func() {
				input := validResource()
				input.Spec.IpAddressType = "None"
				input.Spec.DnsNameLabel = "acme-app"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.IpAddressType = "None"
				input.Spec.DnsNameLabelReusePolicy = "Noreuse"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.IpAddressType = "None"
				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: 80}}
				input.Spec.ExposedPorts = []*AzureContainerInstancePort{{Port: 80}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject exposed ports no container declares, by port and by protocol", func() {
				input := validResource()
				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: 80}}
				input.Spec.ExposedPorts = []*AzureContainerInstancePort{{Port: 443}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.ExposedPorts = []*AzureContainerInstancePort{{Port: 80, Protocol: "UDP"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject ports outside 1-65535", func() {
				input := validResource()
				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: 65536}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Containers[0].Ports = []*AzureContainerInstancePort{{Port: -1}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a container without a name, image, cpu, or memory", func() {
				input := validResource()
				input.Spec.Containers[0].Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Containers[0].Image = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Containers[0].Cpu = 0
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Containers[0].Memory = 0
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a volume with no form and a volume with two forms", func() {
				input := validResource()
				input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{volumeOf()}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				twoForms := volumeOf()
				twoForms.AzureFile = validAzureFile()
				twoForms.EmptyDir = true
				input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{twoForms}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

				gitAndSecret := volumeOf()
				gitAndSecret.GitRepo = &AzureContainerInstanceVolumeGitRepo{Url: "https://github.com/acme/config"}
				gitAndSecret.Secret = map[string]string{"a": "Yg=="}
				input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{gitAndSecret}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an Azure File volume missing any of its three fields", func() {
				for _, mutate := range []func(*AzureContainerInstanceVolumeAzureFile){
					func(f *AzureContainerInstanceVolumeAzureFile) { f.ShareName = nil },
					func(f *AzureContainerInstanceVolumeAzureFile) { f.StorageAccountName = nil },
					func(f *AzureContainerInstanceVolumeAzureFile) { f.StorageAccountKey = nil },
				} {
					input := validResource()
					volume := volumeOf()
					volume.AzureFile = validAzureFile()
					mutate(volume.AzureFile)
					input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{volume}
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				}
			})

			ginkgo.It("should reject a volume missing its name or mount path", func() {
				input := validResource()
				volume := volumeOf()
				volume.EmptyDir = true
				volume.Name = ""
				input.Spec.Containers[0].Volumes = []*AzureContainerInstanceVolume{volume}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				volume.Name = "data"
				volume.MountPath = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject identity flavors with mismatched identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureContainerInstanceIdentity{Type: AzureContainerInstanceIdentityType_USER_ASSIGNED}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Identity = &AzureContainerInstanceIdentity{
					Type:        AzureContainerInstanceIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Identity = &AzureContainerInstanceIdentity{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Log Analytics metadata without a log type, and unknown log types", func() {
				input := validResource()
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceId:  literal("00000000-0000-0000-0000-000000000000"),
					WorkspaceKey: literal("d29ya3NwYWNlLWtleQ=="),
					Metadata:     map[string]string{"team": "platform"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.DiagnosticsLogAnalytics.Metadata = nil
				input.Spec.DiagnosticsLogAnalytics.LogType = "AllLogs"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the undeployable ContainerInstanceLogs log type and accept ContainerInsights with metadata", func() {
				// ARM rejects metadata for ContainerInstanceLogs
				// (LogAnalyticsMetadataNotAllowed) and the provider always
				// sends a metadata object alongside a log type, so the value
				// cannot deploy through either engine -- the CEL front-loads
				// the failure at validation.
				input := validResource()
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceId:  literal("00000000-0000-0000-0000-000000000000"),
					WorkspaceKey: literal("d29ya3NwYWNlLWtleQ=="),
					LogType:      "ContainerInstanceLogs",
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.DiagnosticsLogAnalytics.LogType = "ContainerInsights"
				input.Spec.DiagnosticsLogAnalytics.Metadata = map[string]string{"node-name": "platform"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject metadata keys outside ARM's closed vocabulary", func() {
				// ARM validates the metadata KEY set server-side
				// (InvalidLogAnalyticsMetadataKeys, live-proven): only
				// pod-uuid, cluster-resource-id, and node-name pass.
				input := validResource()
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceId:  literal("00000000-0000-0000-0000-000000000000"),
					WorkspaceKey: literal("d29ya3NwYWNlLWtleQ=="),
					LogType:      "ContainerInsights",
					Metadata:     map[string]string{"team": "platform"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.DiagnosticsLogAnalytics.Metadata = map[string]string{
					"pod-uuid":            "00000000-0000-0000-0000-000000000000",
					"cluster-resource-id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/c",
					"node-name":           "n1",
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject Log Analytics diagnostics missing the workspace id or key", func() {
				input := validResource()
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceKey: literal("d29ya3NwYWNlLWtleQ=="),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.DiagnosticsLogAnalytics = &AzureContainerInstanceLogAnalytics{
					WorkspaceId: literal("00000000-0000-0000-0000-000000000000"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject probe schemes outside the provider's lowercase vocabulary and negative timings", func() {
				input := validResource()
				input.Spec.Containers[0].LivenessProbe = &AzureContainerInstanceProbe{
					HttpGet: &AzureContainerInstanceProbeHttpGet{Port: 8080, Scheme: "HTTP"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Containers[0].LivenessProbe = &AzureContainerInstanceProbe{
					Exec:          []string{"true"},
					PeriodSeconds: -5,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject custom DNS without nameservers", func() {
				input := validResource()
				input.Spec.DnsConfig = &AzureContainerInstanceDnsConfig{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a registry credential without its server", func() {
				input := validResource()
				input.Spec.ImageRegistryCredentials = []*AzureContainerInstanceImageRegistryCredential{{Username: "puller", Password: "hunter2"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
