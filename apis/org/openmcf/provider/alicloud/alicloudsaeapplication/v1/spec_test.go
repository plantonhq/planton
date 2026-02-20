package alicloudsaeapplicationv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudSaeApplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudSaeApplicationSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudSaeApplicationSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal Image deployment", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata: &shared.CloudResourceMetadata{
					Name: "my-app",
				},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "hello-sae",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/my-ns/my-app:v1",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with FatJar deployment and JDK options", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata: &shared.CloudResourceMetadata{
					Name: "java-app",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudSaeApplicationSpec{
					Region:              "cn-shanghai",
					AppName:             "order-service",
					PackageType:         "FatJar",
					Replicas:            3,
					Cpu:                 4000,
					Memory:              8192,
					PackageUrl:          "https://my-bucket.oss-cn-shanghai.aliyuncs.com/app.jar",
					PackageVersion:      "1.0.0",
					Jdk:                 "Open JDK 11",
					JarStartOptions:     "-Xms512m -Xmx4096m",
					ProgrammingLanguage: proto.String("java"),
					Timezone:            "Asia/Shanghai",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with War deployment", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "war-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "legacy-web",
					PackageType: "War",
					Replicas:    2,
					Cpu:         2000,
					Memory:      4096,
					PackageUrl:  "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/app.war",
					Jdk:         "Open JDK 8",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PythonZip deployment", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "py-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:              "cn-hangzhou",
					AppName:             "flask-api",
					PackageType:         "PythonZip",
					Replicas:            1,
					Cpu:                 500,
					Memory:              1024,
					PackageUrl:          "https://my-bucket.oss-cn-hangzhou.aliyuncs.com/app.zip",
					ProgrammingLanguage: proto.String("other"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with VPC configuration", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "vpc-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "vpc-service",
					PackageType: "Image",
					Replicas:    2,
					Cpu:         2000,
					Memory:      4096,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					VpcId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-abc123"},
					},
					VswitchId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					SecurityGroupId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-def456"},
					},
					NamespaceId: "cn-hangzhou:my-ns",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with health checks", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "healthy-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "healthy-service",
					PackageType: "Image",
					Replicas:    2,
					Cpu:         2000,
					Memory:      4096,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					Liveness: &AlicloudSaeApplicationHealthCheck{
						HttpGet: &AlicloudSaeApplicationHttpGetAction{
							Path: "/healthz",
							Port: 8080,
						},
						InitialDelaySeconds: proto.Int32(10),
						PeriodSeconds:       proto.Int32(30),
						TimeoutSeconds:      proto.Int32(5),
						FailureThreshold:    proto.Int32(3),
					},
					Readiness: &AlicloudSaeApplicationHealthCheck{
						TcpSocket: &AlicloudSaeApplicationTcpSocketAction{
							Port: 8080,
						},
						InitialDelaySeconds: proto.Int32(5),
						PeriodSeconds:       proto.Int32(10),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with exec health check", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "exec-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "exec-check",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					Liveness: &AlicloudSaeApplicationHealthCheck{
						Exec: &AlicloudSaeApplicationExecAction{
							Command: "cat /tmp/healthy",
						},
						InitialDelaySeconds: proto.Int32(5),
						PeriodSeconds:       proto.Int32(10),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with environment variables and command args", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "env-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "env-service",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					Command:     "/app/server",
					CommandArgs: []string{"--port", "8080", "--log-level", "info"},
					Envs: map[string]string{
						"DB_HOST":     "rds.internal",
						"DB_PORT":     "3306",
						"ENVIRONMENT": "production",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with custom host aliases", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "alias-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "alias-service",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         500,
					Memory:      1024,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					CustomHostAliases: []*AlicloudSaeApplicationCustomHostAlias{
						{HostName: "db.internal", Ip: "10.0.1.100"},
						{HostName: "cache.internal", Ip: "10.0.1.200"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with update strategy", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "strategy-app"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "strategy-service",
					PackageType: "Image",
					Replicas:    4,
					Cpu:         2000,
					Memory:      4096,
					ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					UpdateStrategy: &AlicloudSaeApplicationUpdateStrategy{
						Type: proto.String("BatchUpdate"),
						BatchUpdate: &AlicloudSaeApplicationBatchUpdate{
							Batch:         proto.Int32(2),
							BatchWaitTime: proto.Int32(10),
							ReleaseType:   proto.String("auto"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all CPU tiers", func() {
			cpuValues := []int32{500, 1000, 2000, 4000, 8000, 16000, 32000}
			for _, cpu := range cpuValues {
				input := &AlicloudSaeApplication{
					ApiVersion: "alicloud.openmcf.org/v1",
					Kind:       "AlicloudSaeApplication",
					Metadata:   &shared.CloudResourceMetadata{Name: "cpu-test"},
					Spec: &AlicloudSaeApplicationSpec{
						Region:      "cn-hangzhou",
						AppName:     "cpu-test",
						PackageType: "Image",
						Replicas:    1,
						Cpu:         cpu,
						Memory:      2048,
						ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with all memory tiers", func() {
			memValues := []int32{1024, 2048, 4096, 8192, 12288, 16384, 24576, 32768, 65536, 131072}
			for _, mem := range memValues {
				input := &AlicloudSaeApplication{
					ApiVersion: "alicloud.openmcf.org/v1",
					Kind:       "AlicloudSaeApplication",
					Metadata:   &shared.CloudResourceMetadata{Name: "mem-test"},
					Spec: &AlicloudSaeApplicationSpec{
						Region:      "cn-hangzhou",
						AppName:     "mem-test",
						PackageType: "Image",
						Replicas:    1,
						Cpu:         1000,
						Memory:      mem,
						ImageUrl:    "registry.cn-hangzhou.aliyuncs.com/ns/app:v1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with full production configuration", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-app",
					Org:  "acme-corp",
					Env:  "production",
					Id:   "acsae-prod-app",
				},
				Spec: &AlicloudSaeApplicationSpec{
					Region:         "cn-hangzhou",
					AppName:        "production-api",
					AppDescription: "Production REST API service",
					PackageType:    "Image",
					Replicas:       4,
					Cpu:            4000,
					Memory:         8192,
					ImageUrl:       "registry.cn-hangzhou.aliyuncs.com/prod/api:v2.1.0",
					VpcId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-prod"},
					},
					VswitchId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vsw-prod-a"},
					},
					SecurityGroupId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-prod"},
					},
					NamespaceId: "cn-hangzhou:production",
					Command:     "/app/server",
					CommandArgs: []string{"--config", "/etc/app/config.yaml"},
					Envs: map[string]string{
						"ENV":       "production",
						"LOG_LEVEL": "warn",
					},
					Timezone:                      "Asia/Shanghai",
					TerminationGracePeriodSeconds: proto.Int32(30),
					MinReadyInstances:             proto.Int32(2),
					AcrInstanceId:                 "cri-abc123",
					ProgrammingLanguage:           proto.String("other"),
					Liveness: &AlicloudSaeApplicationHealthCheck{
						HttpGet: &AlicloudSaeApplicationHttpGetAction{
							Path: "/healthz",
							Port: 8080,
						},
						InitialDelaySeconds: proto.Int32(15),
						PeriodSeconds:       proto.Int32(30),
						TimeoutSeconds:      proto.Int32(5),
						FailureThreshold:    proto.Int32(3),
						SuccessThreshold:    proto.Int32(1),
					},
					Readiness: &AlicloudSaeApplicationHealthCheck{
						HttpGet: &AlicloudSaeApplicationHttpGetAction{
							Path: "/ready",
							Port: 8080,
						},
						InitialDelaySeconds: proto.Int32(5),
						PeriodSeconds:       proto.Int32(10),
						TimeoutSeconds:      proto.Int32(3),
					},
					CustomHostAliases: []*AlicloudSaeApplicationCustomHostAlias{
						{HostName: "db.internal", Ip: "10.0.1.100"},
					},
					UpdateStrategy: &AlicloudSaeApplicationUpdateStrategy{
						Type: proto.String("BatchUpdate"),
						BatchUpdate: &AlicloudSaeApplicationBatchUpdate{
							Batch:         proto.Int32(2),
							BatchWaitTime: proto.Int32(10),
							ReleaseType:   proto.String("auto"),
						},
					},
					Tags: map[string]string{
						"team":        "platform",
						"cost_center": "eng-123",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when app_name is missing", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when app_name exceeds 36 characters", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "this-app-name-is-way-too-long-for-sae-limits",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when package_type is invalid", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Docker",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when replicas is zero", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    0,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cpu is not a valid tier", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         750,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when memory is not a valid tier", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      3072,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when termination_grace_period_seconds exceeds 60", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:                        "cn-hangzhou",
					AppName:                       "test",
					PackageType:                   "Image",
					Replicas:                      1,
					Cpu:                           1000,
					Memory:                        2048,
					TerminationGracePeriodSeconds: proto.Int32(120),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when termination_grace_period_seconds is zero", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:                        "cn-hangzhou",
					AppName:                       "test",
					PackageType:                   "Image",
					Replicas:                      1,
					Cpu:                           1000,
					Memory:                        2048,
					TerminationGracePeriodSeconds: proto.Int32(0),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when programming_language is invalid", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:              "cn-hangzhou",
					AppName:             "test",
					PackageType:         "Image",
					Replicas:            1,
					Cpu:                 1000,
					Memory:              2048,
					ProgrammingLanguage: proto.String("python"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when update_strategy type is invalid", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
					UpdateStrategy: &AlicloudSaeApplicationUpdateStrategy{
						Type: proto.String("RollingUpdate"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when batch_update release_type is invalid", func() {
			input := &AlicloudSaeApplication{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudSaeApplication",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudSaeApplicationSpec{
					Region:      "cn-hangzhou",
					AppName:     "test",
					PackageType: "Image",
					Replicas:    1,
					Cpu:         1000,
					Memory:      2048,
					UpdateStrategy: &AlicloudSaeApplicationUpdateStrategy{
						Type: proto.String("BatchUpdate"),
						BatchUpdate: &AlicloudSaeApplicationBatchUpdate{
							ReleaseType: proto.String("canary"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
