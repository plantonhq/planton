package gcpbinaryauthorizationattestorv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBinaryAuthorizationAttestorSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const pgpKey = "-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nmQENBFtP0doBCADF\n-----END PGP PUBLIC KEY BLOCK-----\n"
const pemKey = "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\n-----END PUBLIC KEY-----\n"

var _ = ginkgo.Describe("GcpBinaryAuthorizationAttestorSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBinaryAuthorizationAttestor {
		return &GcpBinaryAuthorizationAttestor{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBinaryAuthorizationAttestor",
			Metadata:   &shared.CloudResourceMetadata{Name: "built-by-ci"},
			Spec: &GcpBinaryAuthorizationAttestorSpec{
				Note: &GcpBinaryAuthorizationAttestorNote{HumanReadableName: "CI build pipeline"},
			},
		}
	}

	withKey := func(key *GcpBinaryAuthorizationAttestorPublicKey) *GcpBinaryAuthorizationAttestor {
		msg := minimal()
		msg.Spec.AttestationAuthorityNote = &GcpBinaryAuthorizationAttestorAuthorityNote{
			PublicKeys: []*GcpBinaryAuthorizationAttestorPublicKey{key},
		}
		return msg
	}

	ginkgo.It("should accept an attestor that creates its note, without keys yet", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every key form", func() {
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{AsciiArmoredPgpPublicKey: pgpKey}))).To(gomega.Succeed())
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{
			Id:            "ni:///sha-256;abc",
			PkixPublicKey: &GcpBinaryAuthorizationAttestorPkixPublicKey{PublicKeyPem: pemKey, SignatureAlgorithm: "ECDSA_P256_SHA256"},
		}))).To(gomega.Succeed())
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{
			Comment:       "held by Cloud KMS",
			PkixPublicKey: &GcpBinaryAuthorizationAttestorPkixPublicKey{KmsKeyVersion: litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k/cryptoKeyVersions/1")},
		}))).To(gomega.Succeed())
	})

	ginkgo.It("should accept an existing note reference with every field set", func() {
		msg := withKey(&GcpBinaryAuthorizationAttestorPublicKey{AsciiArmoredPgpPublicKey: pgpKey})
		msg.Spec.Note = nil
		msg.Spec.AttestationAuthorityNote.NoteReference = "projects/sec/notes/built-by-ci"
		msg.Spec.ProjectId = litRef("sec")
		msg.Spec.AttestorName = "built-by-ci"
		msg.Spec.Description = "Images built by the CI pipeline"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a fully described created note", func() {
		msg := minimal()
		msg.Spec.Note = &GcpBinaryAuthorizationAttestorNote{
			NoteName:          "ci-note",
			HumanReadableName: "CI build pipeline",
			ShortDescription:  "Signed by CI",
			LongDescription:   "Every image CI builds from main is signed by this attestor.",
			ExpirationTime:    "2027-01-01T00:00:00Z",
			RelatedNoteNames:  []string{"projects/sec/notes/qa"},
			RelatedUrl:        []*GcpBinaryAuthorizationAttestorNoteRelatedUrl{{Url: "https://runbooks.example.com/ci", Label: "Runbook"}},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of a created note or a note reference", func() {
		msg := minimal()
		msg.Spec.Note = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AttestationAuthorityNote = &GcpBinaryAuthorizationAttestorAuthorityNote{NoteReference: "projects/sec/notes/n"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require the created note's readable name", func() {
		msg := minimal()
		msg.Spec.Note.HumanReadableName = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of a PGP or a PKIX key, and no id on a PGP key", func() {
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{
			AsciiArmoredPgpPublicKey: pgpKey,
			PkixPublicKey:            &GcpBinaryAuthorizationAttestorPkixPublicKey{PublicKeyPem: pemKey, SignatureAlgorithm: "ECDSA_P256_SHA256"},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{AsciiArmoredPgpPublicKey: pgpKey, Id: "x"}))).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one PKIX source and the algorithm with a PEM only", func() {
		kms := litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k/cryptoKeyVersions/1")
		for _, pkix := range []*GcpBinaryAuthorizationAttestorPkixPublicKey{
			{},
			{PublicKeyPem: pemKey},
			{SignatureAlgorithm: "ECDSA_P256_SHA256"},
			{PublicKeyPem: pemKey, SignatureAlgorithm: "ECDSA_P256_SHA256", KmsKeyVersion: kms},
			{SignatureAlgorithm: "ECDSA_P256_SHA256", KmsKeyVersion: kms},
			{PublicKeyPem: pemKey, SignatureAlgorithm: "DSA_SHA1"},
		} {
			gomega.Expect(validator.Validate(withKey(&GcpBinaryAuthorizationAttestorPublicKey{PkixPublicKey: pkix}))).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should reject a KMS key literal that is not a key version", func() {
		msg := withKey(&GcpBinaryAuthorizationAttestorPublicKey{
			PkixPublicKey: &GcpBinaryAuthorizationAttestorPkixPublicKey{KmsKeyVersion: litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k")},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
