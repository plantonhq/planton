package awsefsaccesspointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsEfsAccessPointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEfsAccessPointSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsEfsAccessPointSpec validations", func() {
	var spec *AwsEfsAccessPointSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: region + file_system_id.
		spec = &AwsEfsAccessPointSpec{
			Region:       "us-east-1",
			FileSystemId: strRef("fs-0123456789abcdef0"),
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec (file system only)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a POSIX user with secondary GIDs", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{
			Uid:           proto.Int64(1000),
			Gid:           proto.Int64(1000),
			SecondaryGids: []int64{1001, 1002},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts uid/gid 0 (root)", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(0), Gid: proto.Int64(0)}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("rejects a POSIX user that names only one of uid and gid", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(1000)}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).ToNot(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("posix_user.gid"))
	})

	ginkgo.It("accepts a root directory with creation info", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(1000), Gid: proto.Int64(1000)}
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{
			Path: "/app/data",
			CreationInfo: &AwsEfsAccessPointCreationInfo{
				OwnerUid:    1000,
				OwnerGid:    1000,
				Permissions: "0755",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts the file system root as the root directory", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{Path: "/"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts 3-digit octal permissions", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{
			Path: "/logs",
			CreationInfo: &AwsEfsAccessPointCreationInfo{
				OwnerUid:    1001,
				OwnerGid:    1001,
				Permissions: "750",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure cases
	// -------------------------------------------------------------------------

	ginkgo.It("fails when region is missing", func() {
		spec.Region = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when file_system_id is missing", func() {
		spec.FileSystemId = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when uid is out of range", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(4294967296), Gid: proto.Int64(1000)}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when uid is negative", func() {
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(-1), Gid: proto.Int64(1000)}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when more than 16 secondary GIDs are set", func() {
		gids := make([]int64, 17)
		for i := range gids {
			gids[i] = int64(2000 + i)
		}
		spec.PosixUser = &AwsEfsAccessPointPosixUser{Uid: proto.Int64(1000), Gid: proto.Int64(1000), SecondaryGids: gids}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when root directory path is relative", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{Path: "app/data"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when root directory path is empty", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{Path: ""}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when permissions are not octal", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{
			Path: "/app",
			CreationInfo: &AwsEfsAccessPointCreationInfo{
				OwnerUid:    1000,
				OwnerGid:    1000,
				Permissions: "0788",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when permissions are missing from creation info", func() {
		spec.RootDirectory = &AwsEfsAccessPointRootDirectory{
			Path: "/app",
			CreationInfo: &AwsEfsAccessPointCreationInfo{
				OwnerUid: 1000,
				OwnerGid: 1000,
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
