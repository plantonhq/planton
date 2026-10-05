package kubernetesseaweedfsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesSeaweedFs(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSeaweedFs Suite")
}

func boolPtr(b bool) *bool       { return &b }
func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesSeaweedFs Validation Tests", func() {
	var input *KubernetesSeaweedFs

	ginkgo.BeforeEach(func() {
		input = &KubernetesSeaweedFs{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSeaweedFs",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-seaweedfs",
			},
			Spec: &KubernetesSeaweedFsSpec{
				Namespace: literal("object-store"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "object-store", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an HA topology (3 masters, 3 volumes, replication 001) should be valid", func() {
			input.Spec.Master = &KubernetesSeaweedFsMaster{Replicas: int32Ptr(3)}
			input.Spec.Volume = &KubernetesSeaweedFsVolume{Replicas: int32Ptr(3)}
			input.Spec.Replication = "001"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every volume index mode in the vocabulary should be valid", func() {
			for _, mode := range []string{"memory", "leveldb", "leveldbMedium", "leveldbLarge"} {
				input.Spec.Volume = &KubernetesSeaweedFsVolume{IndexMode: mode}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("buckets with TTL, versioning, object lock and anonymous read should be valid", func() {
			input.Spec.S3 = &KubernetesSeaweedFsS3{
				Buckets: []*KubernetesSeaweedFsS3Bucket{
					{Name: "artifacts", Ttl: "7d", Versioning: true},
					{Name: "public-assets", AnonymousRead: true},
					{Name: "compliance-archive", ObjectLock: true},
					{Name: "a1.b-2", Ttl: "255y"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a dedicated S3 gateway should be valid", func() {
			input.Spec.S3 = &KubernetesSeaweedFsS3{
				Dedicated: &KubernetesSeaweedFsS3Dedicated{Replicas: int32Ptr(2)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (tiers, s3, admin, replication, image) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("4.40.0")
			input.Spec.Master = &KubernetesSeaweedFsMaster{
				Replicas:          int32Ptr(3),
				DataVolume:        &KubernetesSeaweedFsDataVolume{Size: "5Gi", StorageClass: literal("gp3")},
				VolumeSizeLimitMb: int32Ptr(1000),
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "256Mi"},
				},
			}
			input.Spec.Volume = &KubernetesSeaweedFsVolume{
				Replicas:            int32Ptr(3),
				DataVolume:          &KubernetesSeaweedFsDataVolume{Size: "100Gi"},
				MaxVolumes:          0,
				IndexMode:           "leveldb",
				MinFreeSpacePercent: int32Ptr(5),
			}
			input.Spec.Filer = &KubernetesSeaweedFsFiler{
				Replicas:          int32Ptr(1),
				DataVolume:        &KubernetesSeaweedFsDataVolume{Size: "10Gi"},
				EncryptVolumeData: true,
				ExtraEnvironmentVars: map[string]string{
					"WEED_FILER_OPTIONS_RECURSIVE_DELETE": "true",
				},
			}
			input.Spec.S3 = &KubernetesSeaweedFsS3{
				Enabled:    boolPtr(true),
				EnableAuth: boolPtr(true),
				Buckets:    []*KubernetesSeaweedFsS3Bucket{{Name: "artifacts"}},
				DomainName: "s3.example.internal",
				Dedicated:  &KubernetesSeaweedFsS3Dedicated{Replicas: int32Ptr(2)},
			}
			input.Spec.Replication = "001"
			input.Spec.Admin = &KubernetesSeaweedFsAdmin{
				Enabled:    true,
				DataVolume: &KubernetesSeaweedFsDataVolume{Size: "10Gi"},
			}
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.Image = &KubernetesSeaweedFsImage{
				Registry:   "mirror.example.com",
				Repository: "chrislusf/seaweedfs",
				Tag:        "4.40",
			}
			input.Spec.HelmValues = "sftp:\n  enabled: true\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a malformed replication code should fail", func() {
			for _, bad := range []string{"01", "0011", "abc", "0-1"} {
				input.Spec.Replication = bad
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("zero master replicas should fail", func() {
			input.Spec.Master = &KubernetesSeaweedFsMaster{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an unknown volume index mode should fail", func() {
			input.Spec.Volume = &KubernetesSeaweedFsVolume{IndexMode: "redis"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a min_free_space_percent above 100 should fail", func() {
			input.Spec.Volume = &KubernetesSeaweedFsVolume{MinFreeSpacePercent: int32Ptr(101)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an invalid bucket name should fail", func() {
			for _, bad := range []string{"UPPER", "ab", "-starts-with-dash", "ends-with-dash-"} {
				input.Spec.S3 = &KubernetesSeaweedFsS3{
					Buckets: []*KubernetesSeaweedFsS3Bucket{{Name: bad}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("a bucket without a name should fail", func() {
			input.Spec.S3 = &KubernetesSeaweedFsS3{
				Buckets: []*KubernetesSeaweedFsS3Bucket{{Ttl: "7d"}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a malformed bucket TTL should fail", func() {
			for _, bad := range []string{"7", "d", "0d", "256d", "7s", "7 d"} {
				input.Spec.S3 = &KubernetesSeaweedFsS3{
					Buckets: []*KubernetesSeaweedFsS3Bucket{{Name: "artifacts", Ttl: bad}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			}
		})

		ginkgo.It("a malformed data-volume size should fail", func() {
			input.Spec.Volume = &KubernetesSeaweedFsVolume{
				DataVolume: &KubernetesSeaweedFsDataVolume{Size: "100 gigabytes"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero dedicated-gateway replicas should fail", func() {
			input.Spec.S3 = &KubernetesSeaweedFsS3{
				Dedicated: &KubernetesSeaweedFsS3Dedicated{Replicas: int32Ptr(0)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
