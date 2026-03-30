package atlasmongodbv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAtlasMongodb(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AtlasMongodb Suite")
}

var _ = ginkgo.Describe("AtlasMongodb Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("atlas_mongodb with minimal valid fields", func() {
			var input *AtlasMongodb

			ginkgo.BeforeEach(func() {
				input = &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-atlas-resource",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "some-project-id",
							ClusterType:              "REPLICASET",
							ElectableNodes:           3,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "7.0",
							ProviderName:             "AWS",
							ProviderInstanceSizeName: "M10",
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("atlas_mongodb with GCP provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mongodb-gcp",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id-gcp",
							ClusterType:              "REPLICASET",
							ElectableNodes:           5,
							Priority:                 7,
							ReadOnlyNodes:            2,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: false,
							MongoDbMajorVersion:      "6.0",
							ProviderName:             "GCP",
							ProviderInstanceSizeName: "M30",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("atlas_mongodb with Azure provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mongodb-azure",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id-azure",
							ClusterType:              "SHARDED",
							ElectableNodes:           7,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "5.0",
							ProviderName:             "AZURE",
							ProviderInstanceSizeName: "M50",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("atlas_mongodb with GEOSHARDED cluster type", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mongodb-geosharded",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id-geo",
							ClusterType:              "GEOSHARDED",
							ElectableNodes:           5,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "7.0",
							ProviderName:             "AWS",
							ProviderInstanceSizeName: "M40",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata:   nil,
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id",
							ClusterType:              "REPLICASET",
							ElectableNodes:           3,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "7.0",
							ProviderName:             "AWS",
							ProviderInstanceSizeName: "M10",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-atlas-mongodb",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "AtlasMongodb",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-atlas-mongodb",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id",
							ClusterType:              "REPLICASET",
							ElectableNodes:           3,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "7.0",
							ProviderName:             "AWS",
							ProviderInstanceSizeName: "M10",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &AtlasMongodb{
					ApiVersion: "atlas.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-atlas-mongodb",
					},
					Spec: &AtlasMongodbSpec{
						ClusterConfig: &AtlasMongodbClusterConfig{
							ProjectId:                "test-project-id",
							ClusterType:              "REPLICASET",
							ElectableNodes:           3,
							Priority:                 7,
							ReadOnlyNodes:            0,
							CloudBackup:              true,
							AutoScalingDiskGbEnabled: true,
							MongoDbMajorVersion:      "7.0",
							ProviderName:             "AWS",
							ProviderInstanceSizeName: "M10",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
