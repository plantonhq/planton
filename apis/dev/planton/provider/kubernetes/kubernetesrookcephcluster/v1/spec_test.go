package kubernetesrookcephclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesRookCephCluster(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesRookCephCluster Suite")
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

var _ = ginkgo.Describe("KubernetesRookCephClusterSpec Validation Tests", func() {
	var spec *KubernetesRookCephClusterSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesRookCephClusterSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "rook-ceph",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with block pool configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "ceph-blockpool",
						StorageClass: &CephStorageClassSpec{
							Name: "ceph-block",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with filesystem configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Filesystems = []*CephFilesystemSpec{
					{
						Name: "ceph-filesystem",
						StorageClass: &CephStorageClassSpec{
							Name: "ceph-filesystem",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with object store configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ObjectStores = []*CephObjectStoreSpec{
					{
						Name: "ceph-objectstore",
						StorageClass: &CephStorageClassSpec{
							Name: "ceph-bucket",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with cluster configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					DataDirHostPath: stringPtr("/var/lib/rook"),
					Mon: &CephMonSpec{
						Count:                int32Ptr(3),
						AllowMultiplePerNode: boolPtr(false),
					},
					Mgr: &CephMgrSpec{
						Count:                int32Ptr(2),
						AllowMultiplePerNode: boolPtr(false),
					},
					Storage: &CephStorageSpec{
						UseAllNodes:   boolPtr(true),
						UseAllDevices: boolPtr(true),
					},
					Network: &CephNetworkSpec{
						EnableEncryption:  boolPtr(false),
						EnableCompression: boolPtr(false),
						RequireMsgr2:      boolPtr(false),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with ceph image configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.CephImage = &CephImageSpec{
					Repository:       stringPtr("quay.io/ceph/ceph"),
					Tag:              stringPtr("v19.2.3"),
					AllowUnsupported: boolPtr(false),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with storage class reclaim policy Delete", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:          "ceph-block",
							ReclaimPolicy: stringPtr("Delete"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with storage class reclaim policy Retain", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:          "ceph-block",
							ReclaimPolicy: stringPtr("Retain"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with volume binding mode Immediate", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:              "ceph-block",
							VolumeBindingMode: stringPtr("Immediate"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with volume binding mode WaitForFirstConsumer", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:              "ceph-block",
							VolumeBindingMode: stringPtr("WaitForFirstConsumer"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with storage node spec", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					Storage: &CephStorageSpec{
						UseAllNodes:   boolPtr(false),
						UseAllDevices: boolPtr(false),
						Nodes: []*CephStorageNodeSpec{
							{
								Name:    "node1",
								Devices: []string{"sda", "sdb"},
							},
							{
								Name:         "node2",
								DeviceFilter: "^sd[a-z]$",
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with missing namespace", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Namespace = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty operator namespace", func() {
			ginkgo.It("should return a validation error", func() {
				spec.OperatorNamespace = stringPtr("")
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty helm chart version", func() {
			ginkgo.It("should return a validation error", func() {
				spec.HelmChartVersion = stringPtr("")
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty block pool name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty filesystem name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Filesystems = []*CephFilesystemSpec{
					{
						Name: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty object store name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ObjectStores = []*CephObjectStoreSpec{
					{
						Name: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid storage class reclaim policy", func() {
			ginkgo.It("should return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:          "ceph-block",
							ReclaimPolicy: stringPtr("Invalid"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid volume binding mode", func() {
			ginkgo.It("should return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name:              "ceph-block",
							VolumeBindingMode: stringPtr("Invalid"),
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with mon count out of range (too high)", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					Mon: &CephMonSpec{
						Count: int32Ptr(15),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with mon count out of range (zero)", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					Mon: &CephMonSpec{
						Count: int32Ptr(0),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with mgr count out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					Mgr: &CephMgrSpec{
						Count: int32Ptr(10),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with replicated size out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name:           "test-pool",
						ReplicatedSize: int32Ptr(10),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with gateway port out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ObjectStores = []*CephObjectStoreSpec{
					{
						Name:        "test-store",
						GatewayPort: int32Ptr(70000),
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty storage node name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					Storage: &CephStorageSpec{
						Nodes: []*CephStorageNodeSpec{
							{
								Name: "",
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid data dir host path (not absolute)", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Cluster = &CephClusterConfig{
					DataDirHostPath: stringPtr("relative/path"),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty ceph image repository", func() {
			ginkgo.It("should return a validation error", func() {
				spec.CephImage = &CephImageSpec{
					Repository: stringPtr(""),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty ceph image tag", func() {
			ginkgo.It("should return a validation error", func() {
				spec.CephImage = &CephImageSpec{
					Tag: stringPtr(""),
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty storage class name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.BlockPools = []*CephBlockPoolSpec{
					{
						Name: "test-pool",
						StorageClass: &CephStorageClassSpec{
							Name: "",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
