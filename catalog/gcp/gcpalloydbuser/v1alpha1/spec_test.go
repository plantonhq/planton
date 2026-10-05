package gcpalloydbuserv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpAlloydbUserSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func fromRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func ptr(s string) *string {
	return &s
}

var _ = ginkgo.Describe("GcpAlloydbUserSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimalBuiltIn := func() *GcpAlloydbUser {
		return &GcpAlloydbUser{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpAlloydbUser",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-user",
			},
			Spec: &GcpAlloydbUserSpec{
				Cluster:  litRef("projects/p/locations/us-central1/clusters/orders"),
				UserId:   "orders-app",
				Password: "AppPassword123!",
			},
		}
	}

	expectValid := func(r *GcpAlloydbUser) {
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	}

	expectInvalid := func(r *GcpAlloydbUser, substr string) {
		err := validator.Validate(r)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), substr)).To(
			gomega.BeTrue(), "expected error to contain %q, got: %s", substr, err)
	}

	ginkgo.It("accepts a minimal built-in user", func() {
		expectValid(minimalBuiltIn())
	})

	ginkgo.It("accepts a passwordless built-in user", func() {
		r := minimalBuiltIn()
		r.Spec.Password = ""
		expectValid(r)
	})

	ginkgo.It("rejects a missing cluster", func() {
		r := minimalBuiltIn()
		r.Spec.Cluster = nil
		expectInvalid(r, "cluster")
	})

	ginkgo.It("accepts a cluster reference", func() {
		r := minimalBuiltIn()
		r.Spec.Cluster = fromRef("orders-cluster")
		expectValid(r)
	})

	ginkgo.It("rejects a missing user_id", func() {
		r := minimalBuiltIn()
		r.Spec.UserId = ""
		expectInvalid(r, "user_id")
	})

	ginkgo.It("rejects an unknown user_type", func() {
		r := minimalBuiltIn()
		r.Spec.UserType = ptr("SUPERUSER")
		expectInvalid(r, "user_type")
	})

	ginkgo.It("rejects a password on an IAM user", func() {
		r := minimalBuiltIn()
		r.Spec.UserType = ptr("ALLOYDB_IAM_USER")
		r.Spec.UserId = "dev@example.com"
		expectInvalid(r, "must not set a password")
	})

	ginkgo.It("accepts an IAM user without a password", func() {
		r := minimalBuiltIn()
		r.Spec.UserType = ptr("ALLOYDB_IAM_USER")
		r.Spec.UserId = "dev@example.com"
		r.Spec.Password = ""
		expectValid(r)
	})

	ginkgo.It("accepts database roles", func() {
		r := minimalBuiltIn()
		r.Spec.DatabaseRoles = []string{"alloydbiamuser", "alloydbsuperuser"}
		expectValid(r)
	})

	ginkgo.It("accepts every deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			r := minimalBuiltIn()
			r.Spec.DeletionPolicy = v
			expectValid(r)
		}
	})

	ginkgo.It("rejects an unknown deletion_policy", func() {
		r := minimalBuiltIn()
		r.Spec.DeletionPolicy = "FORCE"
		expectInvalid(r, "deletion_policy")
	})

	ginkgo.It("rejects a wrong kind constant", func() {
		r := minimalBuiltIn()
		r.Kind = "GcpAlloydbCluster"
		expectInvalid(r, "kind")
	})
})
