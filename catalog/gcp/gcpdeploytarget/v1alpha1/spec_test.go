package gcpdeploytargetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDeployTargetSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpDeployTargetSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDeployTarget {
		return &GcpDeployTarget{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDeployTarget",
			Metadata:   &shared.CatalogObjectMetadata{Name: "staging"},
			Spec: &GcpDeployTargetSpec{
				Location: "us-central1",
				Run:      &GcpDeployTargetRun{Location: "projects/my-app-staging/locations/us-central1"},
			},
		}
	}

	full := func() *GcpDeployTarget {
		msg := minimal()
		msg.Spec.ProjectId = reference(catalogkind.CatalogKind_GcpProject, "delivery")
		msg.Spec.TargetId = "staging"
		msg.Spec.Description = "Staging on Cloud Run"
		msg.Spec.Labels = map[string]string{"env": "staging"}
		msg.Spec.Annotations = map[string]string{"owner": "platform"}
		msg.Spec.RequireApproval = true
		msg.Spec.DeployParameters = map[string]string{"replicas": "2"}
		msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{
			{
				Usages:           []string{"RENDER", "DEPLOY"},
				WorkerPool:       reference(catalogkind.CatalogKind_GcpCloudBuildWorkerPool, "private-builds"),
				ServiceAccount:   reference(catalogkind.CatalogKind_GcpServiceAccount, "deployer"),
				ArtifactStorage:  reference(catalogkind.CatalogKind_GcpGcsBucket, "deploy-artifacts"),
				ExecutionTimeout: "3600s",
				Verbose:          true,
			},
			{
				Usages:      []string{"VERIFY"},
				DefaultPool: &GcpDeployTargetDefaultPool{ServiceAccount: literal("verifier@delivery.iam.gserviceaccount.com"), ArtifactStorage: literal("gs://deploy-artifacts/verify")},
			},
			{
				Usages:      []string{"PREDEPLOY", "POSTDEPLOY"},
				PrivatePool: &GcpDeployTargetPrivatePool{WorkerPool: literal("projects/delivery/locations/us-central1/workerPools/hooks")},
			},
		}
		return msg
	}

	ginkgo.It("should accept a minimal Cloud Run target and a fully declared one", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(full())).To(gomega.Succeed())
	})

	ginkgo.It("should accept each of the other target types", func() {
		gke := minimal()
		gke.Spec.Run = nil
		gke.Spec.Gke = &GcpDeployTargetGke{Cluster: reference(catalogkind.CatalogKind_GcpGkeCluster, "prod"), DnsEndpoint: true, ProxyUrl: "http://10.0.0.5:3128"}
		gke.Spec.AssociatedEntities = []*GcpDeployTargetAssociatedEntity{
			{
				EntityId:       "config-cluster",
				GkeClusters:    []*GcpDeployTargetAssociatedGkeCluster{{Cluster: literal("projects/p/locations/us-central1/clusters/config"), InternalIp: true}},
				AnthosClusters: []*GcpDeployTargetAssociatedAnthosCluster{{Membership: reference(catalogkind.CatalogKind_GcpGkeCluster, "edge")}},
			},
		}
		gomega.Expect(validator.Validate(gke)).To(gomega.Succeed())

		anthos := minimal()
		anthos.Spec.Run = nil
		anthos.Spec.AnthosCluster = &GcpDeployTargetAnthosCluster{Membership: literal("projects/p/locations/global/memberships/edge")}
		gomega.Expect(validator.Validate(anthos)).To(gomega.Succeed())

		multi := minimal()
		multi.Spec.Run = nil
		multi.Spec.MultiTarget = &GcpDeployTargetMultiTarget{TargetIds: []*foreignkeyv1.StringValueOrRef{
			reference(catalogkind.CatalogKind_GcpDeployTarget, "prod-us"),
			literal("prod-eu"),
		}}
		gomega.Expect(validator.Validate(multi)).To(gomega.Succeed())

		custom := minimal()
		custom.Spec.Run = nil
		custom.Spec.CustomTarget = &GcpDeployTargetCustomTarget{CustomTargetType: reference(catalogkind.CatalogKind_GcpDeployCustomTargetType, "terraform")}
		gomega.Expect(validator.Validate(custom)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one target type", func() {
		msg := minimal()
		msg.Spec.Run = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Gke = &GcpDeployTargetGke{Cluster: literal("projects/p/locations/us-central1/clusters/prod")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse dns_endpoint with internal_ip and malformed cluster names", func() {
		msg := minimal()
		msg.Spec.Run = nil
		msg.Spec.Gke = &GcpDeployTargetGke{Cluster: literal("projects/p/locations/us-central1/clusters/prod"), DnsEndpoint: true, InternalIp: true}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Gke = &GcpDeployTargetGke{Cluster: literal("prod")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AssociatedEntities = []*GcpDeployTargetAssociatedEntity{{EntityId: "config", GkeClusters: []*GcpDeployTargetAssociatedGkeCluster{{Cluster: literal("config")}}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require full membership names when literal", func() {
		msg := minimal()
		msg.Spec.Run = nil
		msg.Spec.AnthosCluster = &GcpDeployTargetAnthosCluster{Membership: literal("edge")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AssociatedEntities = []*GcpDeployTargetAssociatedEntity{{EntityId: "edge", AnthosClusters: []*GcpDeployTargetAssociatedAnthosCluster{{Membership: literal("edge")}}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a well-formed Cloud Run location", func() {
		for _, location := range []string{"", "us-central1", "projects/p/regions/us-central1"} {
			msg := minimal()
			msg.Spec.Run.Location = location
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), location)
		}
	})

	ginkgo.It("should require bare child target IDs and at least one", func() {
		msg := minimal()
		msg.Spec.Run = nil
		msg.Spec.MultiTarget = &GcpDeployTargetMultiTarget{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.MultiTarget = &GcpDeployTargetMultiTarget{TargetIds: []*foreignkeyv1.StringValueOrRef{literal("projects/p/locations/us-central1/targets/prod-us")}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a full custom target type name", func() {
		msg := minimal()
		msg.Spec.Run = nil
		msg.Spec.CustomTarget = &GcpDeployTargetCustomTarget{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.CustomTarget = &GcpDeployTargetCustomTarget{CustomTargetType: literal("terraform")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.CustomTarget = &GcpDeployTargetCustomTarget{CustomTargetType: literal("projects/p/locations/us-central1/customTargetTypes/terraform")}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require well-formed, unique entity IDs", func() {
		msg := minimal()
		msg.Spec.AssociatedEntities = []*GcpDeployTargetAssociatedEntity{{EntityId: "config"}, {EntityId: "config"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		for _, id := range []string{"", "Config", "1config", "config-"} {
			msg = minimal()
			msg.Spec.AssociatedEntities = []*GcpDeployTargetAssociatedEntity{{EntityId: id}}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should hold execution configurations to Google's usage rules", func() {
		msg := full()
		msg.Spec.ExecutionConfigs[1].Usages = []string{"DEPLOY"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a usage in two configurations")

		msg = minimal()
		msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{{Usages: []string{"RENDER"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "DEPLOY not covered")

		msg = minimal()
		msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{{Usages: []string{"RENDER", "DEPLOY", "ANALYSIS"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "unknown usage")

		msg = minimal()
		msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{{Usages: []string{"RENDER", "DEPLOY", "DEPLOY"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "usage repeated in one configuration")

		msg = minimal()
		msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{{Usages: []string{"RENDER", "DEPLOY"}}, {}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a configuration without usages")
	})

	ginkgo.It("should refuse both pool blocks and malformed execution fields", func() {
		msg := full()
		msg.Spec.ExecutionConfigs[2].DefaultPool = &GcpDeployTargetDefaultPool{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "both pools")

		msg = full()
		msg.Spec.ExecutionConfigs[2].PrivatePool = &GcpDeployTargetPrivatePool{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "private pool without a worker pool")

		bad := []func(*GcpDeployTargetExecutionConfig){
			func(c *GcpDeployTargetExecutionConfig) { c.WorkerPool = literal("hooks") },
			func(c *GcpDeployTargetExecutionConfig) { c.ArtifactStorage = literal("deploy-artifacts") },
			func(c *GcpDeployTargetExecutionConfig) { c.ExecutionTimeout = "1h" },
			func(c *GcpDeployTargetExecutionConfig) {
				c.PrivatePool = &GcpDeployTargetPrivatePool{WorkerPool: literal("hooks")}
			},
			func(c *GcpDeployTargetExecutionConfig) {
				c.PrivatePool = &GcpDeployTargetPrivatePool{WorkerPool: literal("projects/p/locations/l/workerPools/w"), ArtifactStorage: literal("bucket")}
			},
			func(c *GcpDeployTargetExecutionConfig) {
				c.DefaultPool = &GcpDeployTargetDefaultPool{ArtifactStorage: literal("bucket/path")}
			},
		}
		for i, mutate := range bad {
			msg = minimal()
			config := &GcpDeployTargetExecutionConfig{Usages: []string{"RENDER", "DEPLOY"}}
			mutate(config)
			msg.Spec.ExecutionConfigs = []*GcpDeployTargetExecutionConfig{config}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), i)
		}
	})

	ginkgo.It("should reject a malformed target ID, a long description, and an unknown deletion policy", func() {
		for _, id := range []string{"Staging", "1staging", "staging-", "staging_us"} {
			msg := minimal()
			msg.Spec.TargetId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg := minimal()
		msg.Spec.Description = string(make([]byte, 256))
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

// A fleet-registered cluster composes from either way a cluster joins a
// fleet: an explicit membership's name, or the fleet_membership a
// GcpGkeCluster with fleet_project exports.
func TestAnthosMembershipAcceptsBothRegistrationPaths(t *testing.T) {
	for name, field := range map[string]refannotations.Field{
		"anthos_cluster.membership":           refannotations.Of((&GcpDeployTargetAnthosCluster{}).ProtoReflect().Descriptor().Fields().ByName("membership")),
		"associated_entities.anthos_clusters": refannotations.Of((&GcpDeployTargetAssociatedAnthosCluster{}).ProtoReflect().Descriptor().Fields().ByName("membership")),
	} {
		if path, ok := field.DefaultPath(catalogkind.CatalogKind_GcpGkeCluster); !ok || path != "status.outputs.fleet_membership" {
			t.Fatalf("%s composes from a GcpGkeCluster at %q (ok=%t), want status.outputs.fleet_membership", name, path, ok)
		}
		if path, ok := field.DefaultPath(catalogkind.CatalogKind_GcpGkeFleetMembership); !ok || path != "status.outputs.name" {
			t.Fatalf("%s composes from a GcpGkeFleetMembership at %q (ok=%t), want status.outputs.name", name, path, ok)
		}
		if kind := field.EffectiveKind(catalogkind.CatalogKind_unspecified); kind != catalogkind.CatalogKind_GcpGkeFleetMembership {
			t.Fatalf("%s: a kindless valueFrom reads as %s, want GcpGkeFleetMembership", name, kind)
		}
	}
}
