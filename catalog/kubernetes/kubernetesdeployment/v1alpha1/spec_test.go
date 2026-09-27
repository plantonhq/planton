package kubernetesdeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func validContainer() *kubernetes.WorkloadContainer {
	return &kubernetes.WorkloadContainer{
		Image: &kubernetes.WorkloadContainerImage{
			Repo: "ghcr.io/acme/api",
			Tag:  "v1.0.0",
		},
	}
}

func validSpec() *KubernetesDeploymentSpec {
	return &KubernetesDeploymentSpec{
		Namespace: literal("prod"),
		Container: &KubernetesDeploymentContainer{
			App: validContainer(),
		},
	}
}

func TestKubernetesDeploymentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesDeploymentSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesDeploymentSpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal spec with namespace and app container", func() {
			err := protovalidate.Validate(validSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a full-featured spec with sidecars, init containers, scheduling, and availability", func() {
			app := validContainer()
			app.Name = "app"
			app.Ports = []*kubernetes.WorkloadContainerPort{
				{Name: "http", ContainerPort: 8080, ServicePort: 80},
			}
			spec := &KubernetesDeploymentSpec{
				Namespace:       literal("prod"),
				CreateNamespace: true,
				Version:         proto.String("review-42"),
				Container: &KubernetesDeploymentContainer{
					App: app,
					Sidecars: []*kubernetes.WorkloadContainer{
						{
							Name: "log-shipper",
							Image: &kubernetes.WorkloadContainerImage{
								Repo: "ghcr.io/acme/fluent-bit",
								Tag:  "2.2.0",
							},
						},
					},
				},
				Pod: &kubernetes.WorkloadPod{
					ServiceAccount: literal("checkout-identity"),
					InitContainers: []*kubernetes.WorkloadContainer{
						{
							Name: "init-migrate",
							Image: &kubernetes.WorkloadContainerImage{
								Repo: "ghcr.io/acme/migrator",
								Tag:  "v1.0.0",
							},
						},
					},
					Scheduling: &kubernetes.WorkloadScheduling{
						NodeSelector: map[string]string{"workload-tier": "general"},
						Tolerations: []*kubernetes.WorkloadToleration{
							{Key: "dedicated", Operator: "Equal", Value: "checkout", Effect: "NoSchedule"},
						},
						TopologySpreadConstraints: []*kubernetes.WorkloadTopologySpreadConstraint{
							{MaxSkew: 1, TopologyKey: "topology.kubernetes.io/zone", WhenUnsatisfiable: "ScheduleAnyway"},
						},
					},
					SecurityContext: &kubernetes.WorkloadPodSecurityContext{
						RunAsNonRoot: proto.Bool(true),
						FsGroup:      proto.Int64(2000),
					},
					TerminationGracePeriodSeconds: proto.Int64(60),
				},
				Availability: &KubernetesDeploymentAvailability{
					Replicas: proto.Int32(3),
					HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
						Enabled:                     true,
						MaxReplicas:                 10,
						TargetCpuUtilizationPercent: proto.Int32(60),
					},
					Strategy: &KubernetesDeploymentStrategy{
						Type:           "RollingUpdate",
						MaxUnavailable: "0",
						MaxSurge:       "1",
					},
					PodDisruptionBudget: &KubernetesDeploymentPodDisruptionBudget{
						Enabled:      true,
						MinAvailable: "2",
					},
					MinReadySeconds:         proto.Int32(15),
					RevisionHistoryLimit:    proto.Int32(5),
					ProgressDeadlineSeconds: proto.Int32(600),
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts version \"main\" and a single-character version", func() {
			spec := validSpec()
			spec.Version = proto.String("main")
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())

			spec.Version = proto.String("a")
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts HPA enabled with only a memory target", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
					Enabled:                 true,
					MaxReplicas:             5,
					TargetMemoryUtilization: "1Gi",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the Recreate strategy", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				Strategy: &KubernetesDeploymentStrategy{Type: "Recreate"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a PDB bounded only by max_unavailable", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				PodDisruptionBudget: &KubernetesDeploymentPodDisruptionBudget{
					Enabled:        true,
					MaxUnavailable: "25%",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a disabled PDB even when both bounds are set", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				PodDisruptionBudget: &KubernetesDeploymentPodDisruptionBudget{
					Enabled:        false,
					MinAvailable:   "1",
					MaxUnavailable: "1",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a disabled HPA with no metrics", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{Enabled: false},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects a spec without a namespace", func() {
			spec := validSpec()
			spec.Namespace = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a spec without a container", func() {
			spec := validSpec()
			spec.Container = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a container group without an app container", func() {
			spec := validSpec()
			spec.Container = &KubernetesDeploymentContainer{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an app container without an image", func() {
			spec := validSpec()
			spec.Container.App = &kubernetes.WorkloadContainer{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an image without a repo", func() {
			spec := validSpec()
			spec.Container.App.Image = &kubernetes.WorkloadContainerImage{Tag: "v1.0.0"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an image with neither a tag nor a digest", func() {
			spec := validSpec()
			spec.Container.App.Image = &kubernetes.WorkloadContainerImage{Repo: "ghcr.io/acme/api"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("accepts an image pinned by digest, with or without its tag", func() {
			const digest = "sha256:b5e2a1c0d9f8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2"
			for _, image := range []*kubernetes.WorkloadContainerImage{
				{Repo: "ghcr.io/acme/api", Tag: "a743940", Digest: digest},
				{Repo: "ghcr.io/acme/api", Digest: digest},
			} {
				spec := validSpec()
				spec.Container.App.Image = image
				gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
			}
		})

		ginkgo.It("rejects a digest that is not sha256 and 64 lowercase hex characters", func() {
			for _, digest := range []string{"sha256:abc", "b5e2a1c0d9f8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2",
				"sha256:B5E2A1C0D9F8E7A6B5C4D3E2F1A0B9C8D7E6F5A4B3C2D1E0F9A8B7C6D5E4F3A2"} {
				spec := validSpec()
				spec.Container.App.Image = &kubernetes.WorkloadContainerImage{Repo: "ghcr.io/acme/api", Tag: "a743940", Digest: digest}
				gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.Succeed(), digest)
			}
		})

		ginkgo.It("rejects a version with uppercase characters", func() {
			spec := validSpec()
			spec.Version = proto.String("Main")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a version ending with a hyphen", func() {
			spec := validSpec()
			spec.Version = proto.String("review-")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a version longer than 30 characters", func() {
			spec := validSpec()
			spec.Version = proto.String("a-very-long-version-name-over-30-chars")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unnamed sidecar", func() {
			spec := validSpec()
			spec.Container.Sidecars = []*kubernetes.WorkloadContainer{
				{
					Image: &kubernetes.WorkloadContainerImage{Repo: "ghcr.io/acme/proxy", Tag: "1.0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a container name that is not a lowercase DNS label", func() {
			spec := validSpec()
			spec.Container.Sidecars = []*kubernetes.WorkloadContainer{
				{
					Name:  "Log_Shipper",
					Image: &kubernetes.WorkloadContainerImage{Repo: "ghcr.io/acme/proxy", Tag: "1.0"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an enabled HPA with no target metric", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
					Enabled:     true,
					MaxReplicas: 5,
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an enabled HPA without max_replicas", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
					Enabled:                     true,
					TargetCpuUtilizationPercent: proto.Int32(60),
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a CPU utilization target outside 1-100", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
					Enabled:                     true,
					MaxReplicas:                 5,
					TargetCpuUtilizationPercent: proto.Int32(101),
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a memory utilization target that is not a Kubernetes quantity", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				HorizontalPodAutoscaling: &KubernetesDeploymentHpa{
					Enabled:                 true,
					MaxReplicas:             5,
					TargetMemoryUtilization: "10gigs",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown strategy type", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				Strategy: &KubernetesDeploymentStrategy{Type: "BlueGreen"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a strategy with both max_unavailable and max_surge at zero", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				Strategy: &KubernetesDeploymentStrategy{
					Type:           "RollingUpdate",
					MaxUnavailable: "0",
					MaxSurge:       "0%",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an enabled PDB with both min_available and max_unavailable", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				PodDisruptionBudget: &KubernetesDeploymentPodDisruptionBudget{
					Enabled:        true,
					MinAvailable:   "1",
					MaxUnavailable: "1",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative replica count", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				Replicas: proto.Int32(-1),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a progress deadline below 1 second", func() {
			spec := validSpec()
			spec.Availability = &KubernetesDeploymentAvailability{
				ProgressDeadlineSeconds: proto.Int32(0),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a port name that is not a lowercase DNS label", func() {
			spec := validSpec()
			spec.Container.App.Ports = []*kubernetes.WorkloadContainerPort{
				{Name: "Http", ContainerPort: 8080},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a container port outside 1-65535", func() {
			spec := validSpec()
			spec.Container.App.Ports = []*kubernetes.WorkloadContainerPort{
				{Name: "http", ContainerPort: 70000},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a service port outside 1-65535", func() {
			spec := validSpec()
			spec.Container.App.Ports = []*kubernetes.WorkloadContainerPort{
				{Name: "http", ContainerPort: 8080, ServicePort: 70000},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	// The pull login declared on the workload itself: each entry needs a server,
	// a username, and a password (a $secret/ reference in a real manifest), and a
	// server may appear only once — every entry lands in the same docker-config.
	ginkgo.Context("When image registries are declared on the pod", func() {

		ginkgo.It("accepts one login per private registry", func() {
			spec := validSpec()
			spec.Pod = &kubernetes.WorkloadPod{
				ImageRegistries: []*kubernetes.WorkloadImageRegistry{
					{Server: "ghcr.io", Username: "acme-pull-bot", Password: "$secret/ghcr-pull-token"},
					{Server: "quay.io", Username: "acme+robot", Password: "$secret/quay-robot-token", Email: "ops@acme.io"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("rejects two entries naming the same registry server, saying so", func() {
			spec := validSpec()
			spec.Pod = &kubernetes.WorkloadPod{
				ImageRegistries: []*kubernetes.WorkloadImageRegistry{
					{Server: "ghcr.io", Username: "a", Password: "$secret/a"},
					{Server: "ghcr.io", Username: "b", Password: "$secret/b"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("one login per registry"))
		})

		ginkgo.It("rejects an entry without a server, a username, or a password", func() {
			for _, entry := range []*kubernetes.WorkloadImageRegistry{
				{Username: "a", Password: "$secret/a"},
				{Server: "ghcr.io", Password: "$secret/a"},
				{Server: "ghcr.io", Username: "a"},
			} {
				spec := validSpec()
				spec.Pod = &kubernetes.WorkloadPod{ImageRegistries: []*kubernetes.WorkloadImageRegistry{entry}}
				gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
			}
		})
	})
})
