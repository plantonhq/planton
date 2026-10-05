package gcpcertificatemapv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCertificateMapSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const certName = "projects/my-project/locations/global/certificates/my-cert"

var _ = ginkgo.Describe("GcpCertificateMapSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCertificateMap {
		return &GcpCertificateMap{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCertificateMap",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-certificate-map",
			},
			Spec: &GcpCertificateMapSpec{},
		}
	}

	hostnameEntry := func() *GcpCertificateMapEntry {
		return &GcpCertificateMapEntry{
			EntryName:    "www",
			Hostname:     "www.example.com",
			Certificates: []*foreignkeyv1.StringValueOrRef{litRef(certName)},
		}
	}

	withEntry := func(e *GcpCertificateMapEntry) *GcpCertificateMap {
		m := minimal()
		m.Spec.Entries = []*GcpCertificateMapEntry{e}
		return m
	}

	ginkgo.It("accepts an entry-less map (attach entries later)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts a hostname entry and a PRIMARY matcher entry", func() {
		gomega.Expect(validator.Validate(withEntry(hostnameEntry()))).To(gomega.Succeed())

		e := hostnameEntry()
		e.Hostname = ""
		e.Matcher = "PRIMARY"
		gomega.Expect(validator.Validate(withEntry(e))).To(gomega.Succeed())
	})

	ginkgo.Context("entries", func() {
		ginkgo.It("require an entry name", func() {
			e := hostnameEntry()
			e.EntryName = ""
			gomega.Expect(validator.Validate(withEntry(e))).ToNot(gomega.Succeed())
		})

		ginkgo.It("require exactly one of hostname or matcher", func() {
			e := hostnameEntry()
			e.Matcher = "PRIMARY"
			gomega.Expect(validator.Validate(withEntry(e))).ToNot(gomega.Succeed(), "both set")

			e = hostnameEntry()
			e.Hostname = ""
			gomega.Expect(validator.Validate(withEntry(e))).ToNot(gomega.Succeed(), "neither set")
		})

		ginkgo.It("require 1 to 15 certificates (the API's per-entry cap)", func() {
			e := hostnameEntry()
			e.Certificates = nil
			gomega.Expect(validator.Validate(withEntry(e))).ToNot(gomega.Succeed(), "no certificates")

			e = hostnameEntry()
			e.Certificates = nil
			for i := 0; i < 15; i++ {
				e.Certificates = append(e.Certificates, litRef(certName))
			}
			gomega.Expect(validator.Validate(withEntry(e))).To(gomega.Succeed(), "15 certificates")

			e.Certificates = append(e.Certificates, litRef(certName))
			gomega.Expect(validator.Validate(withEntry(e))).ToNot(gomega.Succeed(), "16 certificates")
		})
	})

	ginkgo.Context("deletion_policy", func() {
		ginkgo.It("accepts the documented values and rejects others", func() {
			for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
				m := minimal()
				m.Spec.DeletionPolicy = v
				gomega.Expect(validator.Validate(m)).To(gomega.Succeed(), "value %q", v)
			}
			m := minimal()
			m.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(validator.Validate(m)).ToNot(gomega.Succeed())
		})
	})
})
