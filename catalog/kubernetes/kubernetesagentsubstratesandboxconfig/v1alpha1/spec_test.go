package kubernetesagentsubstratesandboxconfigv1alpha1

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

// The validation suite pins what the spec refuses before the cluster sees the
// SandboxConfig: the object's name, the closed set of classes, the
// definition's own rules (the default names an Enabled, listed version), the
// per-class asset rules upstream's sandboxconfig-assets admission policy
// enforces at apply,
// unique version names, and well-formed files (a URL and a lower-case hex
// SHA-256).

func TestKubernetesAgentSubstrateSandboxConfig(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSubstrateSandboxConfig Suite")
}

const sum = "fa72f0d45756c69a30beec21222c78241b0d330c36f4f2185648617cc23415c8"

func ptr[T any](v T) *T { return &v }

func files(names ...string) *KubernetesAgentSubstrateSandboxConfigAssetFiles {
	entries := map[string]*KubernetesAgentSubstrateSandboxConfigAssetFile{}
	for _, name := range names {
		entries[name] = &KubernetesAgentSubstrateSandboxConfigAssetFile{Url: "gs://assets/" + name, Sha256: sum}
	}
	return &KubernetesAgentSubstrateSandboxConfigAssetFiles{Entries: entries}
}

func gvisorVersion(name string) *KubernetesAgentSubstrateSandboxConfigVersion {
	return &KubernetesAgentSubstrateSandboxConfigVersion{
		Name: name,
		Assets: map[string]*KubernetesAgentSubstrateSandboxConfigAssetFiles{
			"amd64": files("gvisor"),
			"arm64": files("gvisor"),
		},
	}
}

func microvmVersion(name string) *KubernetesAgentSubstrateSandboxConfigVersion {
	return &KubernetesAgentSubstrateSandboxConfigVersion{
		Name: name,
		Assets: map[string]*KubernetesAgentSubstrateSandboxConfigAssetFiles{
			"amd64": files("cloud-hypervisor", "virtiofsd", "kata-kernel", "kata-image"),
		},
	}
}

func expectViolation(err error, ruleID string) {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var validationErr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "not a validation error: %v", err)
	var ids []string
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	gomega.Expect(ids).To(gomega.ContainElement(ruleID), "violations: %v", err)
}

var _ = ginkgo.Describe("KubernetesAgentSubstrateSandboxConfig Validation Tests", func() {
	var input *KubernetesAgentSubstrateSandboxConfig

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSubstrateSandboxConfig{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSubstrateSandboxConfig",
			Metadata:   &shared.CatalogObjectMetadata{Name: "gvisor-pinned"},
			Spec: &KubernetesAgentSubstrateSandboxConfigSpec{
				Name:           "gvisor-pinned",
				SandboxClass:   "gvisor",
				DefaultVersion: "v2",
				Versions:       []*KubernetesAgentSubstrateSandboxConfigVersion{gvisorVersion("v2"), gvisorVersion("v1")},
			},
		}
	})

	ginkgo.It("accepts a gVisor config a person would write", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a microVM config", func() {
		input.Spec.SandboxClass = "microvm"
		input.Spec.DefaultVersion = "v1"
		input.Spec.Versions = []*KubernetesAgentSubstrateSandboxConfigVersion{microvmVersion("v1")}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("accepts a retired version that is not the default", func() {
		input.Spec.Versions[1].State = ptr("Disabled")
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("refuses a name that is not a DNS subdomain", func() {
		input.Spec.Name = "Gvisor_Pinned"
		expectViolation(protovalidate.Validate(input), "spec.name.dns_subdomain")
	})

	ginkgo.It("requires a class, and only gvisor or microvm", func() {
		input.Spec.SandboxClass = ""
		expectViolation(protovalidate.Validate(input), "required")
		input.Spec.SandboxClass = "kata"
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses a default that names no listed version", func() {
		input.Spec.DefaultVersion = "v3"
		expectViolation(protovalidate.Validate(input), "spec.default_version.listed")
	})

	ginkgo.It("refuses a Disabled default", func() {
		input.Spec.Versions[0].State = ptr("Disabled")
		expectViolation(protovalidate.Validate(input), "spec.default_version.enabled")
	})

	ginkgo.It("refuses a state outside Enabled and Disabled", func() {
		input.Spec.Versions[1].State = ptr("Retired")
		expectViolation(protovalidate.Validate(input), "string.in")
	})

	ginkgo.It("refuses a gVisor version with no assets", func() {
		input.Spec.Versions[1].Assets = nil
		expectViolation(protovalidate.Validate(input), "spec.assets.gvisor")
	})

	ginkgo.It("refuses a gVisor architecture without the gvisor release tarball, a bare runsc binary included", func() {
		input.Spec.Versions[1].Assets["arm64"] = files("kata-kernel")
		expectViolation(protovalidate.Validate(input), "spec.assets.gvisor")
		input.Spec.Versions[1].Assets["arm64"] = files("runsc")
		expectViolation(protovalidate.Validate(input), "spec.assets.gvisor")
	})

	ginkgo.It("refuses a microVM architecture missing one of its four assets", func() {
		input.Spec.SandboxClass = "microvm"
		input.Spec.DefaultVersion = "v1"
		input.Spec.Versions = []*KubernetesAgentSubstrateSandboxConfigVersion{microvmVersion("v1")}
		input.Spec.Versions[0].Assets["arm64"] = files("cloud-hypervisor", "virtiofsd", "kata-kernel")
		expectViolation(protovalidate.Validate(input), "spec.assets.microvm")
	})

	ginkgo.It("refuses two versions with one name", func() {
		input.Spec.Versions[1].Name = "v2"
		expectViolation(protovalidate.Validate(input), "spec.versions.unique_names")
	})

	ginkgo.It("refuses a version name that is not a DNS label", func() {
		input.Spec.Versions[1].Name = "V1.0"
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("requires one to sixteen versions", func() {
		input.Spec.Versions = nil
		expectViolation(protovalidate.Validate(input), "repeated.min_items")
		input.Spec.DefaultVersion = "v0"
		for i := 0; i < 17; i++ {
			input.Spec.Versions = append(input.Spec.Versions, gvisorVersion(fmt.Sprintf("v%d", i)))
		}
		expectViolation(protovalidate.Validate(input), "repeated.max_items")
		input.Spec.Versions = input.Spec.Versions[:16]
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("bounds the name, the default and every version name", func() {
		input.Spec.Name = strings.Repeat("a", 254)
		expectViolation(protovalidate.Validate(input), "string.max_len")
		input.Spec.Name = "gvisor-pinned"
		input.Spec.DefaultVersion = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
		long := strings.Repeat("v", 64)
		input.Spec.DefaultVersion = long
		input.Spec.Versions[0].Name = long
		expectViolation(protovalidate.Validate(input), "string.max_len")
	})

	ginkgo.It("bounds every file's URL", func() {
		input.Spec.Versions[1].Assets["amd64"].Entries["gvisor"].Url = "gs://bucket/" + strings.Repeat("u", 2048)
		expectViolation(protovalidate.Validate(input), "string.max_len")
	})

	ginkgo.It("requires every file's checksum", func() {
		input.Spec.Versions[1].Assets["amd64"].Entries["gvisor"].Sha256 = ""
		expectViolation(protovalidate.Validate(input), "required")
	})

	ginkgo.It("refuses a microVM version with no assets", func() {
		input.Spec.SandboxClass = "microvm"
		input.Spec.DefaultVersion = "v1"
		input.Spec.Versions = []*KubernetesAgentSubstrateSandboxConfigVersion{{Name: "v1"}}
		expectViolation(protovalidate.Validate(input), "spec.assets.microvm")
	})

	ginkgo.It("refuses label keys and values outside Kubernetes' grammar", func() {
		input.Spec.Labels = map[string]string{"team/": "runtime"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
		input.Spec.Labels = map[string]string{"team": "run time"}
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("refuses a checksum that is not lower-case hex SHA-256", func() {
		input.Spec.Versions[1].Assets["amd64"].Entries["gvisor"].Sha256 = "ABC"
		expectViolation(protovalidate.Validate(input), "string.pattern")
	})

	ginkgo.It("requires every file's URL", func() {
		input.Spec.Versions[1].Assets["amd64"].Entries["gvisor"].Url = ""
		expectViolation(protovalidate.Validate(input), "string.min_len")
	})
})
