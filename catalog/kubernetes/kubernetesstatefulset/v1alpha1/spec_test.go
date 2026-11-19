package kubernetesstatefulsetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/internal/manifest/protodefaults"
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
			Repo: "docker.io/library/postgres",
			Tag:  "16.2",
		},
	}
}

func validSpec() *KubernetesStatefulSetSpec {
	return &KubernetesStatefulSetSpec{
		Namespace: literal("prod"),
		Container: &KubernetesStatefulSetContainer{
			App: validContainer(),
		},
	}
}

func validVolumeClaimTemplate() *KubernetesStatefulSetVolumeClaimTemplate {
	return &KubernetesStatefulSetVolumeClaimTemplate{
		Name: "data",
		Size: "10Gi",
	}
}

func TestKubernetesStatefulSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesStatefulSetSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesStatefulSetSpec validations", func() {

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal spec with namespace and app container", func() {
			err := protovalidate.Validate(validSpec())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a full-featured spec with sidecars, storage, scheduling, and availability", func() {
			app := validContainer()
			app.Name = "postgres"
			app.Ports = []*kubernetes.WorkloadContainerPort{
				{Name: "pg", ContainerPort: 5432, ServicePort: 5432},
			}
			spec := &KubernetesStatefulSetSpec{
				Namespace:       literal("prod"),
				CreateNamespace: true,
				Container: &KubernetesStatefulSetContainer{
					App: app,
					Sidecars: []*kubernetes.WorkloadContainer{
						{
							Name: "metrics-exporter",
							Image: &kubernetes.WorkloadContainerImage{
								Repo: "quay.io/prometheuscommunity/postgres-exporter",
								Tag:  "v0.15.0",
							},
						},
					},
				},
				Pod: &kubernetes.WorkloadPod{
					ServiceAccount: literal("db-identity"),
					InitContainers: []*kubernetes.WorkloadContainer{
						{
							Name: "init-permissions",
							Image: &kubernetes.WorkloadContainerImage{
								Repo: "busybox",
								Tag:  "1.36",
							},
						},
					},
					Scheduling: &kubernetes.WorkloadScheduling{
						Tolerations: []*kubernetes.WorkloadToleration{
							{Key: "dedicated", Operator: "Equal", Value: "database", Effect: "NoSchedule"},
						},
					},
					SecurityContext: &kubernetes.WorkloadPodSecurityContext{
						RunAsNonRoot: proto.Bool(true),
						FsGroup:      proto.Int64(999),
						SeccompProfile: &kubernetes.WorkloadSeccompProfile{
							Type: "RuntimeDefault",
						},
					},
					TerminationGracePeriodSeconds: proto.Int64(120),
				},
				Availability: &KubernetesStatefulSetAvailability{
					Replicas: proto.Int32(3),
					PodDisruptionBudget: &KubernetesStatefulSetPodDisruptionBudget{
						Enabled:      true,
						MinAvailable: "2",
					},
					MinReadySeconds:      proto.Int32(10),
					RevisionHistoryLimit: proto.Int32(5),
				},
				VolumeClaimTemplates: []*KubernetesStatefulSetVolumeClaimTemplate{
					{
						Name:         "data",
						StorageClass: "ssd-expandable",
						Size:         "100Gi",
						AccessModes:  []string{"ReadWriteOnce"},
						VolumeMode:   "Filesystem",
					},
				},
				UpdateStrategy: &KubernetesStatefulSetUpdateStrategy{
					Type:      "RollingUpdate",
					Partition: proto.Int32(2),
				},
				PodManagementPolicy: "OrderedReady",
				PvcRetentionPolicy: &KubernetesStatefulSetPvcRetentionPolicy{
					WhenDeleted: "Retain",
					WhenScaled:  "Delete",
				},
				Ordinals: &KubernetesStatefulSetOrdinals{
					Start: proto.Int32(1),
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the Parallel pod management policy", func() {
			spec := validSpec()
			spec.PodManagementPolicy = "Parallel"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts the OnDelete update strategy", func() {
			spec := validSpec()
			spec.UpdateStrategy = &KubernetesStatefulSetUpdateStrategy{Type: "OnDelete"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a volume claim template with a Block volume mode", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.VolumeMode = "Block"
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a PDB bounded only by max_unavailable", func() {
			spec := validSpec()
			spec.Availability = &KubernetesStatefulSetAvailability{
				PodDisruptionBudget: &KubernetesStatefulSetPodDisruptionBudget{
					Enabled:        true,
					MaxUnavailable: "1",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a container-level seccomp profile of type Localhost with a profile path", func() {
			spec := validSpec()
			spec.Container.App.SecurityContext = &kubernetes.WorkloadContainerSecurityContext{
				SeccompProfile: &kubernetes.WorkloadSeccompProfile{
					Type:             "Localhost",
					LocalhostProfile: "profiles/audit.json",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a toleration with the Exists operator and no value", func() {
			spec := validSpec()
			spec.Pod = &kubernetes.WorkloadPod{
				Scheduling: &kubernetes.WorkloadScheduling{
					Tolerations: []*kubernetes.WorkloadToleration{
						{Key: "dedicated", Operator: "Exists", Effect: "NoExecute"},
					},
				},
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
			spec.Container = &KubernetesStatefulSetContainer{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an app image without a tag", func() {
			spec := validSpec()
			spec.Container.App.Image = &kubernetes.WorkloadContainerImage{Repo: "postgres"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unnamed sidecar", func() {
			spec := validSpec()
			spec.Container.Sidecars = []*kubernetes.WorkloadContainer{
				{
					Image: &kubernetes.WorkloadContainerImage{Repo: "busybox", Tag: "1.36"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown pod management policy", func() {
			spec := validSpec()
			spec.PodManagementPolicy = "Sequential"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown update strategy type", func() {
			spec := validSpec()
			spec.UpdateStrategy = &KubernetesStatefulSetUpdateStrategy{Type: "Recreate"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative update strategy partition", func() {
			spec := validSpec()
			spec.UpdateStrategy = &KubernetesStatefulSetUpdateStrategy{
				Type:      "RollingUpdate",
				Partition: proto.Int32(-1),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown when_deleted retention policy", func() {
			spec := validSpec()
			spec.PvcRetentionPolicy = &KubernetesStatefulSetPvcRetentionPolicy{
				WhenDeleted: "Remove",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown when_scaled retention policy", func() {
			spec := validSpec()
			spec.PvcRetentionPolicy = &KubernetesStatefulSetPvcRetentionPolicy{
				WhenScaled: "Keep",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a volume claim template without a name", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.Name = ""
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a volume claim template name that is not a lowercase DNS label", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.Name = "Data_Volume"
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a volume claim template without a size", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.Size = ""
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a volume claim template size that is not a Kubernetes quantity", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.Size = "ten-gigs"
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown access mode", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.AccessModes = []string{"ReadWriteEverywhere"}
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown volume mode", func() {
			spec := validSpec()
			vct := validVolumeClaimTemplate()
			vct.VolumeMode = "Raw"
			spec.VolumeClaimTemplates = []*KubernetesStatefulSetVolumeClaimTemplate{vct}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an enabled PDB with both min_available and max_unavailable", func() {
			spec := validSpec()
			spec.Availability = &KubernetesStatefulSetAvailability{
				PodDisruptionBudget: &KubernetesStatefulSetPodDisruptionBudget{
					Enabled:        true,
					MinAvailable:   "2",
					MaxUnavailable: "1",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative replica count", func() {
			spec := validSpec()
			spec.Availability = &KubernetesStatefulSetAvailability{
				Replicas: proto.Int32(-1),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative ordinals start", func() {
			spec := validSpec()
			spec.Ordinals = &KubernetesStatefulSetOrdinals{Start: proto.Int32(-1)}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a Localhost seccomp profile without a profile path", func() {
			spec := validSpec()
			spec.Container.App.SecurityContext = &kubernetes.WorkloadContainerSecurityContext{
				SeccompProfile: &kubernetes.WorkloadSeccompProfile{
					Type: "Localhost",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a RuntimeDefault seccomp profile carrying a localhost profile path", func() {
			spec := validSpec()
			spec.Container.App.SecurityContext = &kubernetes.WorkloadContainerSecurityContext{
				SeccompProfile: &kubernetes.WorkloadSeccompProfile{
					Type:             "RuntimeDefault",
					LocalhostProfile: "profiles/audit.json",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown toleration operator", func() {
			spec := validSpec()
			spec.Pod = &kubernetes.WorkloadPod{
				Scheduling: &kubernetes.WorkloadScheduling{
					Tolerations: []*kubernetes.WorkloadToleration{
						{Key: "dedicated", Operator: "Matches", Value: "database"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown toleration effect", func() {
			spec := validSpec()
			spec.Pod = &kubernetes.WorkloadPod{
				Scheduling: &kubernetes.WorkloadScheduling{
					Tolerations: []*kubernetes.WorkloadToleration{
						{Key: "dedicated", Operator: "Exists", Effect: "Evict"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("When availability leaves replicas unset", func() {
		// Unset replicas means Kubernetes or an external autoscaler owns the size,
		// so no default may fill it in: a filled-in 1 would make every apply reset
		// the size the autoscaler chose.
		ginkgo.It("keeps replicas unset through the manifest defaults", func() {
			spec := validSpec()
			spec.Availability = &KubernetesStatefulSetAvailability{MinReadySeconds: proto.Int32(10)}
			gomega.Expect(protodefaults.ApplyDefaults(spec)).To(gomega.Succeed())
			gomega.Expect(spec.GetAvailability().Replicas).To(gomega.BeNil())
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})
})
