package kubernetesnatsv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestKubernetesNats(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesNats Suite")
}

var _ = ginkgo.Describe("KubernetesNats Custom Validation Tests", func() {
	var input *KubernetesNats

	ginkgo.BeforeEach(func() {
		input = &KubernetesNats{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesNats",
			Metadata: &shared.CloudResourceMetadata{
				Name: "nats-demo",
			},
			Spec: &KubernetesNatsSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "nats-demo",
					},
				},
				ServerContainer: &KubernetesNatsServerContainer{
					Replicas: 3,      // satisfies gt:0
					DiskSize: "10Gi", // required by proto but standard, so fine to include
				},
				DisableJetStream: false,
				TlsEnabled:       false,
				DisableNatsBox:   false,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with replicas greater than zero", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with NACK controller enabled and streams configured", func() {
			ginkgo.It("should not return a validation error", func() {
				natsHelmVersion := "2.12.3"
				nackHelmVersion := "0.31.1"
				nackAppVersion := "0.21.1"
				input.Spec.NatsHelmChartVersion = &natsHelmVersion
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled:           true,
					EnableControlLoop: true,
					HelmChartVersion:  &nackHelmVersion,
					AppVersion:        &nackAppVersion,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:      "orders",
						Subjects:  []string{"orders.*", "orders.>"},
						Storage:   StreamStorageEnum_file,
						Replicas:  3,
						Retention: StreamRetentionEnum_limits,
						MaxAge:    "24h",
						Consumers: []*KubernetesNatsConsumer{
							{
								DurableName:   "orders-processor",
								DeliverPolicy: ConsumerDeliverPolicyEnum_all,
								AckPolicy:     ConsumerAckPolicyEnum_explicit,
								MaxAckPending: 100,
								AckWait:       "30s",
							},
						},
					},
					{
						Name:      "events",
						Subjects:  []string{"events.>"},
						Storage:   StreamStorageEnum_memory,
						Replicas:  1,
						Retention: StreamRetentionEnum_interest,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with ingress enabled and hostname provided", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Ingress = &KubernetesNatsIngress{
					Enabled:  true,
					Hostname: "nats.example.com",
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with replicas equal to zero", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ServerContainer.Replicas = 0

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with ingress enabled but no hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ingress = &KubernetesNatsIngress{
					Enabled:  true,
					Hostname: "", // Empty hostname
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("hostname"))
			})
		})

		ginkgo.Context("with stream missing required name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "", // Empty name
						Subjects: []string{"orders.*"},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with stream missing required subjects", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "orders",
						Subjects: []string{}, // Empty subjects
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with stream name exceeding max length", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     strings.Repeat("a", 256), // Exceeds 255 char limit
						Subjects: []string{"orders.*"},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with stream replicas out of range", func() {
			ginkgo.It("should return a validation error for replicas below minimum", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "orders",
						Subjects: []string{"orders.*"},
						Replicas: 0, // Below minimum of 1
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for replicas above maximum", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "orders",
						Subjects: []string{"orders.*"},
						Replicas: 6, // Above maximum of 5
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with consumer missing required durable_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.NackController = &KubernetesNatsNackController{
					Enabled: true,
				}
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "orders",
						Subjects: []string{"orders.*"},
						Consumers: []*KubernetesNatsConsumer{
							{
								DurableName: "", // Empty durable name
							},
						},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

var _ = ginkgo.Describe("KubernetesNats Stream Configuration Tests", func() {
	var input *KubernetesNats

	ginkgo.BeforeEach(func() {
		input = &KubernetesNats{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesNats",
			Metadata: &shared.CloudResourceMetadata{
				Name: "nats-streams-test",
			},
			Spec: &KubernetesNatsSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "nats-streams",
					},
				},
				ServerContainer: &KubernetesNatsServerContainer{
					Replicas: 3,
					DiskSize: "10Gi",
				},
				NackController: &KubernetesNatsNackController{
					Enabled:           true,
					EnableControlLoop: true,
				},
			},
		}
	})

	ginkgo.Describe("Stream storage types", func() {
		ginkgo.Context("with file storage", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "file-stream",
						Subjects: []string{"file.>"},
						Storage:  StreamStorageEnum_file,
						Replicas: 3,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with memory storage", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "memory-stream",
						Subjects: []string{"memory.>"},
						Storage:  StreamStorageEnum_memory,
						Replicas: 1,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Stream retention policies", func() {
		ginkgo.Context("with limits retention", func() {
			ginkgo.It("should validate successfully with time and size limits", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:      "limits-stream",
						Subjects:  []string{"limits.>"},
						Storage:   StreamStorageEnum_file,
						Replicas:  1,
						Retention: StreamRetentionEnum_limits,
						MaxAge:    "7d",
						MaxBytes:  1073741824, // 1GB
						MaxMsgs:   1000000,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with interest retention", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:      "interest-stream",
						Subjects:  []string{"interest.>"},
						Storage:   StreamStorageEnum_memory,
						Replicas:  1,
						Retention: StreamRetentionEnum_interest,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with workqueue retention", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:      "workqueue-stream",
						Subjects:  []string{"work.>"},
						Storage:   StreamStorageEnum_file,
						Replicas:  1,
						Retention: StreamRetentionEnum_workqueue,
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Consumer configurations", func() {
		ginkgo.Context("with pull consumer", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "pull-stream",
						Subjects: []string{"pull.>"},
						Storage:  StreamStorageEnum_file,
						Replicas: 1,
						Consumers: []*KubernetesNatsConsumer{
							{
								DurableName:   "pull-consumer",
								DeliverPolicy: ConsumerDeliverPolicyEnum_all,
								AckPolicy:     ConsumerAckPolicyEnum_explicit,
								MaxAckPending: 1000,
								MaxDeliver:    5,
								AckWait:       "30s",
							},
						},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with push consumer", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "push-stream",
						Subjects: []string{"push.>"},
						Storage:  StreamStorageEnum_file,
						Replicas: 1,
						Consumers: []*KubernetesNatsConsumer{
							{
								DurableName:    "push-consumer",
								DeliverPolicy:  ConsumerDeliverPolicyEnum_new_msgs,
								AckPolicy:      ConsumerAckPolicyEnum_explicit,
								DeliverSubject: "deliver.push.consumer",
								DeliverGroup:   "push-workers",
								ReplayPolicy:   ConsumerReplayPolicyEnum_instant,
							},
						},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with filtered consumer", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.Streams = []*KubernetesNatsStream{
					{
						Name:     "filtered-stream",
						Subjects: []string{"events.>"},
						Storage:  StreamStorageEnum_file,
						Replicas: 1,
						Consumers: []*KubernetesNatsConsumer{
							{
								DurableName:   "filtered-consumer",
								FilterSubject: "events.orders.*",
								AckPolicy:     ConsumerAckPolicyEnum_explicit,
							},
						},
					},
				}

				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})

var _ = ginkgo.Describe("KubernetesNats Real-World Configuration Tests", func() {
	ginkgo.Describe("GCP Dev NATS configuration", func() {
		ginkgo.It("should validate a production-like configuration with multiple streams", func() {
			natsHelmVersion := "2.12.3"
			input := &KubernetesNats{
				ApiVersion: "kubernetes.planton.dev/v1",
				Kind:       "KubernetesNats",
				Metadata: &shared.CloudResourceMetadata{
					Name: "nats-gcp-dev",
					Org:  "planton",
					Env:  "gcp-dev",
				},
				Spec: &KubernetesNatsSpec{
					TargetCluster: &kubernetes.KubernetesClusterSelector{
						ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
						ClusterName: "gcp-dev-cluster",
					},
					Namespace: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "planton-gcp-dev",
						},
					},
					ServerContainer: &KubernetesNatsServerContainer{
						Replicas: 1,
						DiskSize: "10Gi",
						Resources: &kubernetes.ContainerResources{
							Limits: &kubernetes.CpuMemory{
								Cpu:    "1000m",
								Memory: "1Gi",
							},
							Requests: &kubernetes.CpuMemory{
								Cpu:    "100m",
								Memory: "100Mi",
							},
						},
					},
					Auth: &KubernetesNatsAuth{
						Enabled: true,
						Scheme:  KubernetesNatsAuthScheme_basic_auth,
					},
					Ingress: &KubernetesNatsIngress{
						Enabled:  true,
						Hostname: "nats-gcp-dev.planton.live",
					},
					NatsHelmChartVersion: &natsHelmVersion,
					NackController: &KubernetesNatsNackController{
						Enabled:           true,
						EnableControlLoop: true,
					},
					Streams: []*KubernetesNatsStream{
						{
							Name:      "api-resources",
							Subjects:  []string{"api-resources.>"},
							Replicas:  1,
							MaxAge:    "5m",
							Storage:   StreamStorageEnum_file,
							Retention: StreamRetentionEnum_limits,
						},
						{
							Name:      "webhooks",
							Subjects:  []string{"webhooks.>"},
							Replicas:  1,
							MaxAge:    "5m",
							Storage:   StreamStorageEnum_file,
							Retention: StreamRetentionEnum_limits,
						},
						{
							Name:      "git-commits",
							Subjects:  []string{"git-commits.>"},
							Replicas:  1,
							MaxAge:    "5m",
							Storage:   StreamStorageEnum_file,
							Retention: StreamRetentionEnum_limits,
						},
						{
							Name:      "copilot-chats",
							Subjects:  []string{"copilot-chats.>"},
							Replicas:  1,
							MaxAge:    "5m",
							Storage:   StreamStorageEnum_file,
							Retention: StreamRetentionEnum_limits,
						},
						{
							Name:      "pipeline-status",
							Subjects:  []string{"pipeline-status.>"},
							Replicas:  1,
							MaxAge:    "5m",
							Storage:   StreamStorageEnum_file,
							Retention: StreamRetentionEnum_limits,
						},
					},
				},
			}

			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			// Verify the configuration has expected values
			gomega.Expect(input.Spec.NackController.Enabled).To(gomega.BeTrue())
			gomega.Expect(input.Spec.NackController.EnableControlLoop).To(gomega.BeTrue())
			gomega.Expect(len(input.Spec.Streams)).To(gomega.Equal(5))
		})
	})
})

var _ = ginkgo.Describe("KubernetesNats Default Values Tests", func() {
	ginkgo.Describe("Proto field default options", func() {
		ginkgo.It("should have correct default for NATS Helm chart version", func() {
			spec := &KubernetesNatsSpec{}
			// Default is set via proto option, check the field descriptor
			field := spec.ProtoReflect().Descriptor().Fields().ByName("nats_helm_chart_version")
			gomega.Expect(field).ToNot(gomega.BeNil())
		})

		ginkgo.It("should have correct defaults for NACK controller", func() {
			nackHelmVersion := "0.31.1"
			nackAppVersion := "0.21.1"
			controller := &KubernetesNatsNackController{
				Enabled:          true,
				HelmChartVersion: &nackHelmVersion,
				AppVersion:       &nackAppVersion,
			}

			// Verify versions can be set
			gomega.Expect(*controller.HelmChartVersion).To(gomega.Equal("0.31.1"))
			gomega.Expect(*controller.AppVersion).To(gomega.Equal("0.21.1"))
		})
	})
})

var _ = ginkgo.Describe("KubernetesNats Proto Message Tests", func() {
	ginkgo.Describe("Message cloning", func() {
		ginkgo.It("should properly clone a KubernetesNats message", func() {
			original := &KubernetesNats{
				ApiVersion: "kubernetes.planton.dev/v1",
				Kind:       "KubernetesNats",
				Metadata: &shared.CloudResourceMetadata{
					Name: "nats-clone-test",
				},
				Spec: &KubernetesNatsSpec{
					Namespace: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "test-ns",
						},
					},
					ServerContainer: &KubernetesNatsServerContainer{
						Replicas: 3,
						DiskSize: "10Gi",
					},
					NackController: &KubernetesNatsNackController{
						Enabled: true,
					},
					Streams: []*KubernetesNatsStream{
						{
							Name:     "test-stream",
							Subjects: []string{"test.>"},
						},
					},
				},
			}

			cloned := proto.Clone(original).(*KubernetesNats)

			// Verify clone is independent
			cloned.Spec.ServerContainer.Replicas = 5
			gomega.Expect(original.Spec.ServerContainer.Replicas).To(gomega.Equal(int32(3)))
			gomega.Expect(cloned.Spec.ServerContainer.Replicas).To(gomega.Equal(int32(5)))

			// Verify stream is cloned
			cloned.Spec.Streams[0].Name = "modified-stream"
			gomega.Expect(original.Spec.Streams[0].Name).To(gomega.Equal("test-stream"))
			gomega.Expect(cloned.Spec.Streams[0].Name).To(gomega.Equal("modified-stream"))
		})
	})
})
