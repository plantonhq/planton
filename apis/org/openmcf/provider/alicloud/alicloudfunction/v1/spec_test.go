package alicloudfunctionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudFunctionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudFunctionSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudFunctionSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "my-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "hello-world",
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all compute sizing fields", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "compute-func",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudFunctionSpec{
					Region:              "cn-shanghai",
					FunctionName:        "data-processor",
					Handler:             "main",
					Runtime:             "go1",
					Description:         "Processes incoming data events",
					Cpu:                 proto.Float64(2.0),
					MemorySize:          proto.Int32(4096),
					Timeout:             proto.Int32(300),
					DiskSize:            proto.Int32(1024),
					InstanceConcurrency: proto.Int32(10),
					EnvironmentVariables: map[string]string{
						"ENV": "production",
					},
					Tags: map[string]string{
						"team": "platform",
					},
					ResourceGroupId: "rg-abc123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with code from OSS", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "oss-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "oss-handler",
					Handler:      "index.handler",
					Runtime:      "nodejs20",
					Code: &AlicloudFunctionCode{
						OssBucketName: "my-code-bucket",
						OssObjectName: "functions/handler.zip",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with VPC configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "vpc-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "db-reader",
					Handler:      "index.handler",
					Runtime:      "python3.10",
					VpcConfig: &AlicloudFunctionVpcConfig{
						VpcId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-123"},
						},
						VswitchIds: []*foreignkeyv1.StringValueOrRef{
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vsw-abc"}},
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vsw-def"}},
						},
						SecurityGroupId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "sg-xyz"},
						},
					},
					InternetAccess: proto.Bool(true),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with log configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "logged-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "logged-handler",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					LogConfig: &AlicloudFunctionLogConfig{
						Project: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-sls-project"},
						},
						Logstore:              "function-logs",
						LogBeginRule:          proto.String("DefaultRegex"),
						EnableInstanceMetrics: proto.Bool(true),
						EnableRequestMetrics:  proto.Bool(true),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with custom container configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "container-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "container-handler",
					Handler:      "not-used",
					Runtime:      "custom-container",
					Cpu:          proto.Float64(4.0),
					MemorySize:   proto.Int32(8192),
					CustomContainerConfig: &AlicloudFunctionCustomContainerConfig{
						Image:      "registry.cn-hangzhou.aliyuncs.com/my-ns/my-func:v1",
						Entrypoint: []string{"/app/entrypoint.sh"},
						Command:    []string{"serve"},
						Port:       proto.Int32(8080),
						HealthCheckConfig: &AlicloudFunctionHealthCheckConfig{
							HttpGetUrl:          "/healthz",
							InitialDelaySeconds: proto.Int32(5),
							PeriodSeconds:       proto.Int32(10),
							TimeoutSeconds:      proto.Int32(2),
							FailureThreshold:    proto.Int32(3),
							SuccessThreshold:    proto.Int32(1),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with custom runtime configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "custom-rt-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "custom-runtime-handler",
					Handler:      "index.handler",
					Runtime:      "custom.debian12",
					CustomRuntimeConfig: &AlicloudFunctionCustomRuntimeConfig{
						Command: []string{"./bootstrap"},
						Args:    []string{"--port", "9000"},
						Port:    proto.Int32(9000),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with lifecycle hooks", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "lifecycle-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "lifecycle-handler",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					InstanceLifecycleConfig: &AlicloudFunctionInstanceLifecycleConfig{
						Initializer: &AlicloudFunctionLifecycleHook{
							Handler: "index.initializer",
							Timeout: proto.Int32(30),
						},
						PreStop: &AlicloudFunctionLifecycleHook{
							Handler: "index.pre_stop",
							Timeout: proto.Int32(15),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with NAS configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "nas-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "nas-handler",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					NasConfig: &AlicloudFunctionNasConfig{
						UserId:  proto.Int32(0),
						GroupId: proto.Int32(0),
						MountPoints: []*AlicloudFunctionNasMountPoint{
							{
								ServerAddr: "0f2a1b2c3d-abc12.cn-hangzhou.nas.aliyuncs.com:/data",
								MountDir:   "/mnt/data",
								EnableTls:  proto.Bool(true),
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with GPU configuration", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "gpu-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "inference",
					Handler:      "index.handler",
					Runtime:      "custom.debian12",
					GpuConfig: &AlicloudFunctionGpuConfig{
						GpuMemorySize: 8192,
						GpuType:       "fc.gpu.ampere.1",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all valid runtimes", func() {
			runtimes := []string{
				"python3.12", "python3.10", "python3.9", "python3",
				"nodejs20", "nodejs18", "nodejs16", "nodejs14",
				"java11", "java8", "go1", "php7.2", "dotnetcore3.1",
				"custom", "custom.debian10", "custom.debian11", "custom.debian12",
				"custom-container",
			}
			for _, rt := range runtimes {
				input := &AlicloudFunction{
					ApiVersion: "alicloud.openmcf.org/v1",
					Kind:       "AlicloudFunction",
					Metadata:   &shared.CloudResourceMetadata{Name: "rt-test"},
					Spec: &AlicloudFunctionSpec{
						Region:       "cn-hangzhou",
						FunctionName: "test-" + rt,
						Handler:      "index.handler",
						Runtime:      rt,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			}
		})

		ginkgo.It("should pass with layers and role reference", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata: &shared.CloudResourceMetadata{
					Name: "layered-func",
				},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "layered-handler",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					Layers: []string{
						"acs:fc:cn-hangzhou:123456:layers/common-libs/versions/3",
					},
					Role: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "acs:ram::123456:role/fc-execution-role",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with compute sizing at boundary values", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "boundary-func"},
				Spec: &AlicloudFunctionSpec{
					Region:              "cn-hangzhou",
					FunctionName:        "boundary",
					Handler:             "index.handler",
					Runtime:             "python3.12",
					Cpu:                 proto.Float64(0.05),
					MemorySize:          proto.Int32(64),
					Timeout:             proto.Int32(1),
					DiskSize:            proto.Int32(512),
					InstanceConcurrency: proto.Int32(1),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			input.Spec.Cpu = proto.Float64(16.0)
			input.Spec.MemorySize = proto.Int32(32768)
			input.Spec.Timeout = proto.Int32(86400)
			input.Spec.InstanceConcurrency = proto.Int32(200)
			err = protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when function_name is missing", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:  "cn-hangzhou",
					Handler: "index.handler",
					Runtime: "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when handler is missing", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when runtime is missing", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when runtime is invalid", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "ruby3.2",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudFunction{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cpu is below minimum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					Cpu:          proto.Float64(0.01),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when cpu exceeds maximum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					Cpu:          proto.Float64(32.0),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when memory_size is below minimum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					MemorySize:   proto.Int32(32),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when timeout exceeds maximum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					Timeout:      proto.Int32(100000),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when disk_size is below minimum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					DiskSize:     proto.Int32(256),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_concurrency exceeds maximum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:              "cn-hangzhou",
					FunctionName:        "test",
					Handler:             "index.handler",
					Runtime:             "python3.12",
					InstanceConcurrency: proto.Int32(500),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when function_name exceeds 128 characters", func() {
			longName := ""
			for i := 0; i < 130; i++ {
				longName += "a"
			}
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: longName,
					Handler:      "index.handler",
					Runtime:      "python3.12",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when gpu_type is invalid", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					GpuConfig: &AlicloudFunctionGpuConfig{
						GpuMemorySize: 8192,
						GpuType:       "nvidia.a100",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when log_begin_rule is invalid", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					LogConfig: &AlicloudFunctionLogConfig{
						LogBeginRule: proto.String("CustomRegex"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when custom_container_config image is empty", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "not-used",
					Runtime:      "custom-container",
					CustomContainerConfig: &AlicloudFunctionCustomContainerConfig{
						Image: "",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when health check initial_delay_seconds exceeds maximum", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "not-used",
					Runtime:      "custom-container",
					CustomContainerConfig: &AlicloudFunctionCustomContainerConfig{
						Image: "registry.cn-hangzhou.aliyuncs.com/ns/img:v1",
						HealthCheckConfig: &AlicloudFunctionHealthCheckConfig{
							InitialDelaySeconds: proto.Int32(200),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when NAS mount_dir is empty", func() {
			input := &AlicloudFunction{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudFunction",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudFunctionSpec{
					Region:       "cn-hangzhou",
					FunctionName: "test",
					Handler:      "index.handler",
					Runtime:      "python3.12",
					NasConfig: &AlicloudFunctionNasConfig{
						MountPoints: []*AlicloudFunctionNasMountPoint{
							{
								ServerAddr: "addr",
								MountDir:   "",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
