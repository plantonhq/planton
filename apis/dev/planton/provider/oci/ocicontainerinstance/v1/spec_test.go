package ocicontainerinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciContainerInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciContainerInstanceSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidContainerInstance() *OciContainerInstance {
	return &OciContainerInstance{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciContainerInstance",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-container-instance",
		},
		Spec: &OciContainerInstanceSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Uocm:PHX-AD-1",
			Shape:              "CI.Standard.E4.Flex",
			ShapeConfig: &OciContainerInstanceSpec_ShapeConfig{
				Ocpus: 1.0,
			},
			Containers: []*OciContainerInstanceSpec_Container{
				{
					ImageUrl: "docker.io/library/nginx:latest",
				},
			},
			Vnics: []*OciContainerInstanceSpec_Vnic{
				{
					SubnetId: newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciContainerInstanceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_container_instance", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidContainerInstance()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name", func() {
				input := minimalValidContainerInstance()
				input.Spec.DisplayName = "My Web Server"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with shape config memory", func() {
				input := minimalValidContainerInstance()
				input.Spec.ShapeConfig.MemoryInGbs = 4.0
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple containers", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers = append(input.Spec.Containers,
					&OciContainerInstanceSpec_Container{
						ImageUrl:    "docker.io/library/redis:7",
						DisplayName: "sidecar-cache",
					},
				)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with container command and arguments", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].Command = []string{"/bin/sh", "-c"}
				input.Spec.Containers[0].Arguments = []string{"echo hello && sleep infinity"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with environment variables", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].EnvironmentVariables = map[string]string{
					"DATABASE_URL": "postgres://db:5432/app",
					"LOG_LEVEL":    "info",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with container resource config", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].ResourceConfig = &OciContainerInstanceSpec_ContainerResourceConfig{
					MemoryLimitInGbs: 2.0,
					VcpusLimit:       1.0,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP health check", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].HealthChecks = []*OciContainerInstanceSpec_HealthCheck{
					{
						HealthCheckType:       OciContainerInstanceSpec_http,
						Port:                  8080,
						Path:                  "/healthz",
						Name:                  "readiness",
						InitialDelayInSeconds: 10,
						IntervalInSeconds:     30,
						TimeoutInSeconds:      5,
						FailureThreshold:      3,
						SuccessThreshold:      1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with TCP health check", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].HealthChecks = []*OciContainerInstanceSpec_HealthCheck{
					{
						HealthCheckType: OciContainerInstanceSpec_tcp,
						Port:            3306,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP health check headers", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].HealthChecks = []*OciContainerInstanceSpec_HealthCheck{
					{
						HealthCheckType: OciContainerInstanceSpec_http,
						Port:            8080,
						Path:            "/health",
						Headers: []*OciContainerInstanceSpec_HealthCheckHeader{
							{Name: "Authorization", Value: "Bearer test-token"},
							{Name: "Accept", Value: "application/json"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with security context", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].SecurityContext = &OciContainerInstanceSpec_SecurityContext{
					IsNonRootUserCheckEnabled: true,
					IsRootFileSystemReadonly:  true,
					RunAsUser:                 1000,
					RunAsGroup:                1000,
					Capabilities: &OciContainerInstanceSpec_Capabilities{
						DropCapabilities: []string{"ALL"},
						AddCapabilities:  []string{"NET_BIND_SERVICE"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with volume mounts", func() {
				input := minimalValidContainerInstance()
				input.Spec.Volumes = []*OciContainerInstanceSpec_Volume{
					{
						Name:       "config",
						VolumeType: OciContainerInstanceSpec_configfile,
						Configs: []*OciContainerInstanceSpec_VolumeConfig{
							{Data: "c2VydmVyIHt9", FileName: "nginx.conf"},
						},
					},
				}
				input.Spec.Containers[0].VolumeMounts = []*OciContainerInstanceSpec_VolumeMount{
					{MountPath: "/etc/nginx/conf.d", VolumeName: "config", IsReadOnly: true},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with emptydir volume", func() {
				input := minimalValidContainerInstance()
				input.Spec.Volumes = []*OciContainerInstanceSpec_Volume{
					{
						Name:         "tmpdata",
						VolumeType:   OciContainerInstanceSpec_emptydir,
						BackingStore: "EPHEMERAL_STORAGE",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with container restart policy", func() {
				input := minimalValidContainerInstance()
				input.Spec.ContainerRestartPolicy = OciContainerInstanceSpec_on_failure
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with fault domain", func() {
				input := minimalValidContainerInstance()
				input.Spec.FaultDomain = "FAULT-DOMAIN-2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with graceful shutdown timeout", func() {
				input := minimalValidContainerInstance()
				input.Spec.GracefulShutdownTimeoutInSeconds = 30
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNS config", func() {
				input := minimalValidContainerInstance()
				input.Spec.DnsConfig = &OciContainerInstanceSpec_DnsConfig{
					Nameservers: []string{"8.8.8.8", "8.8.4.4"},
					Searches:    []string{"example.com"},
					Options:     []string{"ndots:5"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with basic image pull secret", func() {
				input := minimalValidContainerInstance()
				input.Spec.ImagePullSecrets = []*OciContainerInstanceSpec_ImagePullSecret{
					{
						RegistryEndpoint: "ghcr.io",
						SecretType:       OciContainerInstanceSpec_basic,
						Username:         "dXNlcg==",
						Password:         "cGFzcw==",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vault image pull secret", func() {
				input := minimalValidContainerInstance()
				input.Spec.ImagePullSecrets = []*OciContainerInstanceSpec_ImagePullSecret{
					{
						RegistryEndpoint: "us-ashburn-1.ocir.io",
						SecretType:       OciContainerInstanceSpec_vault,
						SecretId:         newStringValueOrRef("ocid1.vaultsecret.oc1..example"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple vnics", func() {
				input := minimalValidContainerInstance()
				input.Spec.Vnics = append(input.Spec.Vnics,
					&OciContainerInstanceSpec_Vnic{
						SubnetId:    newStringValueOrRef("ocid1.subnet.oc1.iad.example2"),
						DisplayName: "secondary-vnic",
					},
				)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with vnic NSGs and public IP", func() {
				input := minimalValidContainerInstance()
				isPublic := true
				input.Spec.Vnics[0].IsPublicIpAssigned = &isPublic
				input.Spec.Vnics[0].NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example"),
				}
				input.Spec.Vnics[0].HostnameLabel = "web-server"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full spec", func() {
				isPublic := false
				input := &OciContainerInstance{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciContainerInstance",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-spec-test",
					},
					Spec: &OciContainerInstanceSpec{
						CompartmentId:                    newStringValueOrRef("ocid1.compartment.oc1..example"),
						AvailabilityDomain:               "Uocm:PHX-AD-1",
						DisplayName:                      "Full Feature Instance",
						Shape:                            "CI.Standard.E4.Flex",
						ShapeConfig:                      &OciContainerInstanceSpec_ShapeConfig{Ocpus: 4.0, MemoryInGbs: 16.0},
						ContainerRestartPolicy:           OciContainerInstanceSpec_on_failure,
						FaultDomain:                      "FAULT-DOMAIN-1",
						GracefulShutdownTimeoutInSeconds: 60,
						DnsConfig: &OciContainerInstanceSpec_DnsConfig{
							Nameservers: []string{"10.0.0.2"},
						},
						ImagePullSecrets: []*OciContainerInstanceSpec_ImagePullSecret{
							{RegistryEndpoint: "ghcr.io", SecretType: OciContainerInstanceSpec_basic, Username: "dXNlcg==", Password: "cGFzcw=="},
						},
						Volumes: []*OciContainerInstanceSpec_Volume{
							{Name: "app-config", VolumeType: OciContainerInstanceSpec_configfile, Configs: []*OciContainerInstanceSpec_VolumeConfig{{Data: "dGVzdA==", FileName: "app.conf"}}},
							{Name: "scratch", VolumeType: OciContainerInstanceSpec_emptydir, BackingStore: "MEMORY"},
						},
						Containers: []*OciContainerInstanceSpec_Container{
							{
								ImageUrl:             "docker.io/library/nginx:1.25",
								DisplayName:          "web",
								Command:              []string{"nginx", "-g", "daemon off;"},
								EnvironmentVariables: map[string]string{"PORT": "8080"},
								ResourceConfig:       &OciContainerInstanceSpec_ContainerResourceConfig{MemoryLimitInGbs: 8.0, VcpusLimit: 2.0},
								HealthChecks: []*OciContainerInstanceSpec_HealthCheck{
									{HealthCheckType: OciContainerInstanceSpec_http, Port: 8080, Path: "/", IntervalInSeconds: 10},
								},
								SecurityContext: &OciContainerInstanceSpec_SecurityContext{
									IsNonRootUserCheckEnabled: true,
									Capabilities:              &OciContainerInstanceSpec_Capabilities{DropCapabilities: []string{"ALL"}},
								},
								VolumeMounts: []*OciContainerInstanceSpec_VolumeMount{
									{MountPath: "/etc/nginx/conf.d", VolumeName: "app-config", IsReadOnly: true},
									{MountPath: "/tmp", VolumeName: "scratch"},
								},
							},
						},
						Vnics: []*OciContainerInstanceSpec_Vnic{
							{
								SubnetId:           newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
								DisplayName:        "primary",
								IsPublicIpAssigned: &isPublic,
								NsgIds:             []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.nsg.oc1.iad.example")},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_container_instance", func() {

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidContainerInstance()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Shape = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape_config is missing", func() {
				input := minimalValidContainerInstance()
				input.Spec.ShapeConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when shape_config ocpus is zero", func() {
				input := minimalValidContainerInstance()
				input.Spec.ShapeConfig.Ocpus = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when containers is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container image_url is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].ImageUrl = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vnics is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Vnics = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vnic subnet_id is missing", func() {
				input := minimalValidContainerInstance()
				input.Spec.Vnics[0].SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health check type is unspecified", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].HealthChecks = []*OciContainerInstanceSpec_HealthCheck{
					{
						HealthCheckType: OciContainerInstanceSpec_health_check_type_unspecified,
						Port:            8080,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health check port is zero", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].HealthChecks = []*OciContainerInstanceSpec_HealthCheck{
					{
						HealthCheckType: OciContainerInstanceSpec_http,
						Port:            0,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume name is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Volumes = []*OciContainerInstanceSpec_Volume{
					{Name: "", VolumeType: OciContainerInstanceSpec_emptydir},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume type is unspecified", func() {
				input := minimalValidContainerInstance()
				input.Spec.Volumes = []*OciContainerInstanceSpec_Volume{
					{Name: "data", VolumeType: OciContainerInstanceSpec_volume_type_unspecified},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image pull secret endpoint is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.ImagePullSecrets = []*OciContainerInstanceSpec_ImagePullSecret{
					{RegistryEndpoint: "", SecretType: OciContainerInstanceSpec_basic},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image pull secret type is unspecified", func() {
				input := minimalValidContainerInstance()
				input.Spec.ImagePullSecrets = []*OciContainerInstanceSpec_ImagePullSecret{
					{RegistryEndpoint: "ghcr.io", SecretType: OciContainerInstanceSpec_secret_type_unspecified},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume mount path is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].VolumeMounts = []*OciContainerInstanceSpec_VolumeMount{
					{MountPath: "", VolumeName: "data"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when volume mount name is empty", func() {
				input := minimalValidContainerInstance()
				input.Spec.Containers[0].VolumeMounts = []*OciContainerInstanceSpec_VolumeMount{
					{MountPath: "/data", VolumeName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidContainerInstance()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := minimalValidContainerInstance()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidContainerInstance()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidContainerInstance()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
