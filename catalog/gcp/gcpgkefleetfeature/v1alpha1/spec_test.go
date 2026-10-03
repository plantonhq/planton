package gcpgkefleetfeaturev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeFleetFeatureSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

func feature(name string, spec *GcpGkeFleetFeatureSpec) *GcpGkeFleetFeature {
	spec.Feature = name
	return &GcpGkeFleetFeature{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpGkeFleetFeature",
		Metadata:   &shared.CatalogObjectMetadata{Name: name},
		Spec:       spec,
	}
}

func gitSync() *GcpGkeFleetFeatureConfigSync {
	return &GcpGkeFleetFeatureConfigSync{
		Enabled: proto.Bool(true),
		Git: &GcpGkeFleetFeatureConfigSyncGit{
			SecretType: "none",
			SyncRepo:   "https://github.com/example/platform-config",
			SyncBranch: "main",
			PolicyDir:  "clusters",
		},
		SourceFormat: "unstructured",
	}
}

func hubConfig() *GcpGkeFleetFeaturePolicyControllerHubConfig {
	return &GcpGkeFleetFeaturePolicyControllerHubConfig{
		InstallSpec:          "INSTALL_SPEC_ENABLED",
		AuditIntervalSeconds: proto.Int64(60),
		Monitoring:           &GcpGkeFleetFeaturePolicyControllerMonitoring{Backends: []string{"CLOUD_MONITORING"}},
		DeploymentConfigs: []*GcpGkeFleetFeaturePolicyControllerDeploymentConfig{
			{Component: "admission", ReplicaCount: proto.Int64(2), PodAffinity: "ANTI_AFFINITY"},
		},
		PolicyContent: &GcpGkeFleetFeaturePolicyControllerPolicyContent{
			Bundles:         []*GcpGkeFleetFeaturePolicyControllerBundle{{Bundle: "pss-baseline-v2022"}},
			TemplateLibrary: &GcpGkeFleetFeaturePolicyControllerTemplateLibrary{Installation: "ALL"},
		},
	}
}

const membershipName = "projects/platform-host/locations/us-central1/memberships/orders-uc1"

var _ = ginkgo.Describe("GcpGkeFleetFeatureSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a feature with no settings", func() {
		gomega.Expect(validator.Validate(feature("multiclusterservicediscovery", &GcpGkeFleetFeatureSpec{}))).To(gomega.Succeed())
	})

	ginkgo.It("should accept each fleet-wide settings block on its own feature", func() {
		for _, msg := range []*GcpGkeFleetFeature{
			feature("multiclusteringress", &GcpGkeFleetFeatureSpec{Multiclusteringress: &GcpGkeFleetFeatureMultiClusterIngress{ConfigMembership: reference(catalogkind.CatalogKind_GcpGkeCluster, "config-cluster")}}),
			feature("fleetobservability", &GcpGkeFleetFeatureSpec{Fleetobservability: &GcpGkeFleetFeatureFleetObservability{LoggingConfig: &GcpGkeFleetFeatureFleetLoggingConfig{DefaultConfig: &GcpGkeFleetFeatureLogRoutingConfig{Mode: "COPY"}, FleetScopeLogsConfig: &GcpGkeFleetFeatureLogRoutingConfig{Mode: "MOVE"}}}}),
			feature("clusterupgrade", &GcpGkeFleetFeatureSpec{Clusterupgrade: &GcpGkeFleetFeatureClusterUpgrade{
				UpstreamFleets: []*foreignkeyv1.StringValueOrRef{reference(catalogkind.CatalogKind_GcpGkeFleet, "dev-fleet")},
				PostConditions: &GcpGkeFleetFeatureUpgradePostConditions{Soaking: "604800s"},
				GkeUpgradeOverrides: []*GcpGkeFleetFeatureGkeUpgradeOverride{{
					Upgrade:        &GcpGkeFleetFeatureGkeUpgrade{Name: "k8s_control_plane", Version: "1.31.1-gke.1146000"},
					PostConditions: &GcpGkeFleetFeatureUpgradePostConditions{Soaking: "86400s"},
				}},
			}}),
			feature("rbacrolebindingactuation", &GcpGkeFleetFeatureSpec{Rbacrolebindingactuation: &GcpGkeFleetFeatureRbacRoleBindingActuation{AllowedCustomRoles: []string{"deployer"}}}),
			feature("workloadidentity", &GcpGkeFleetFeatureSpec{Workloadidentity: &GcpGkeFleetFeatureWorkloadIdentity{ScopeTenancyPool: literal("projects/123/locations/global/workloadIdentityPools/fleet")}}),
		} {
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), msg.Spec.Feature)
		}
	})

	ginkgo.It("should accept member defaults and per-cluster overrides on the member features", func() {
		gomega.Expect(validator.Validate(feature("configmanagement", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Configmanagement: &GcpGkeFleetFeatureConfigManagement{Management: "MANAGEMENT_AUTOMATIC", ConfigSync: gitSync()}},
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{{
				Membership: literal(membershipName),
				Configmanagement: &GcpGkeFleetFeatureMembershipConfigManagement{ConfigSync: &GcpGkeFleetFeatureMembershipConfigSync{
					Oci:         &GcpGkeFleetFeatureConfigSyncOci{SecretType: "gcpserviceaccount", SyncRepo: "us-docker.pkg.dev/p/configs/orders", GcpServiceAccountEmail: reference(catalogkind.CatalogKind_GcpServiceAccount, "config-sync")},
					StopSyncing: proto.Bool(true),
					DeploymentOverrides: []*GcpGkeFleetFeatureConfigSyncDeploymentOverride{{
						DeploymentName: "root-reconciler", DeploymentNamespace: "config-management-system",
						Containers: []*GcpGkeFleetFeatureConfigSyncContainerOverride{{ContainerName: "reconciler", MemoryLimit: "1Gi"}},
					}},
				}},
			}},
		}))).To(gomega.Succeed())

		gomega.Expect(validator.Validate(feature("servicemesh", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_AUTOMATIC"}},
			MembershipConfigs:        []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal(membershipName), Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_MANUAL"}}},
		}))).To(gomega.Succeed())

		gomega.Expect(validator.Validate(feature("policycontroller", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Policycontroller: &GcpGkeFleetFeaturePolicyController{PolicyControllerHubConfig: hubConfig()}},
			MembershipConfigs:        []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal(membershipName), Policycontroller: &GcpGkeFleetFeaturePolicyController{Version: "1.20.0", PolicyControllerHubConfig: hubConfig()}}},
		}))).To(gomega.Succeed())
	})

	ginkgo.It("should refuse a settings block on another feature", func() {
		for _, msg := range []*GcpGkeFleetFeature{
			feature("configmanagement", &GcpGkeFleetFeatureSpec{Multiclusteringress: &GcpGkeFleetFeatureMultiClusterIngress{ConfigMembership: literal(membershipName)}}),
			feature("multiclusteringress", &GcpGkeFleetFeatureSpec{Fleetobservability: &GcpGkeFleetFeatureFleetObservability{}}),
			feature("fleetobservability", &GcpGkeFleetFeatureSpec{Clusterupgrade: &GcpGkeFleetFeatureClusterUpgrade{UpstreamFleets: []*foreignkeyv1.StringValueOrRef{literal("dev-host")}}}),
			feature("workloadidentity", &GcpGkeFleetFeatureSpec{Rbacrolebindingactuation: &GcpGkeFleetFeatureRbacRoleBindingActuation{}}),
			feature("clusterupgrade", &GcpGkeFleetFeatureSpec{Workloadidentity: &GcpGkeFleetFeatureWorkloadIdentity{}}),
			feature("servicemesh", &GcpGkeFleetFeatureSpec{FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Configmanagement: &GcpGkeFleetFeatureConfigManagement{}}}),
			feature("configmanagement", &GcpGkeFleetFeatureSpec{FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_AUTOMATIC"}}}),
			feature("servicemesh", &GcpGkeFleetFeatureSpec{FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Policycontroller: &GcpGkeFleetFeaturePolicyController{PolicyControllerHubConfig: hubConfig()}}}),
		} {
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), msg.Spec.Feature)
		}
	})

	ginkgo.It("should accept membership configs only on member features, one matching block each", func() {
		msg := feature("multiclusteringress", &GcpGkeFleetFeatureSpec{
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal(membershipName), Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_AUTOMATIC"}}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = feature("servicemesh", &GcpGkeFleetFeatureSpec{
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal(membershipName), Policycontroller: &GcpGkeFleetFeaturePolicyController{PolicyControllerHubConfig: hubConfig()}}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = feature("servicemesh", &GcpGkeFleetFeatureSpec{
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal(membershipName)}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = feature("servicemesh", &GcpGkeFleetFeatureSpec{
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{
				{Membership: literal(membershipName), Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_AUTOMATIC"}},
				{Membership: literal(membershipName), Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_MANUAL"}},
			},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = feature("servicemesh", &GcpGkeFleetFeatureSpec{
			MembershipConfigs: []*GcpGkeFleetFeatureMembershipConfig{{Membership: literal("orders-uc1"), Mesh: &GcpGkeFleetFeatureMesh{Management: "MANAGEMENT_AUTOMATIC"}}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold Config Sync to one source, a known secret type, and a repository", func() {
		bad := []*GcpGkeFleetFeatureConfigSync{
			{Git: gitSync().Git, Oci: &GcpGkeFleetFeatureConfigSyncOci{SecretType: "none", SyncRepo: "us-docker.pkg.dev/p/r/i"}},
			{Git: &GcpGkeFleetFeatureConfigSyncGit{SecretType: "SSH", SyncRepo: "git@example.com:c.git"}},
			{Git: &GcpGkeFleetFeatureConfigSyncGit{SyncRepo: "https://example.com/c.git"}},
			{Git: &GcpGkeFleetFeatureConfigSyncGit{SecretType: "none"}},
			{Oci: &GcpGkeFleetFeatureConfigSyncOci{SecretType: "token", SyncRepo: "us-docker.pkg.dev/p/r/i"}},
			{Git: &GcpGkeFleetFeatureConfigSyncGit{SecretType: "none", SyncRepo: "https://example.com/c.git", SyncWaitSecs: -1}},
			{SourceFormat: "flat"},
		}
		for i, sync := range bad {
			msg := feature("configmanagement", &GcpGkeFleetFeatureSpec{
				FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Configmanagement: &GcpGkeFleetFeatureConfigManagement{ConfigSync: sync}},
			})
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), i)
		}
		msg := feature("configmanagement", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Configmanagement: &GcpGkeFleetFeatureConfigManagement{Management: "AUTOMATIC"}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold Policy Controller to its install states, components, and bundles", func() {
		mutate := []func(*GcpGkeFleetFeaturePolicyControllerHubConfig){
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.InstallSpec = "" },
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.InstallSpec = "ENABLED" },
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) {
				h.DeploymentConfigs = append(h.DeploymentConfigs, &GcpGkeFleetFeaturePolicyControllerDeploymentConfig{Component: "admission"})
			},
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.DeploymentConfigs[0].Component = "webhook" },
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.DeploymentConfigs[0].PodAffinity = "SPREAD" },
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) {
				h.PolicyContent.Bundles = append(h.PolicyContent.Bundles, &GcpGkeFleetFeaturePolicyControllerBundle{Bundle: "pss-baseline-v2022"})
			},
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) {
				h.PolicyContent.TemplateLibrary.Installation = "SOME"
			},
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.Monitoring.Backends = []string{"DATADOG"} },
			func(h *GcpGkeFleetFeaturePolicyControllerHubConfig) { h.AuditIntervalSeconds = proto.Int64(-5) },
		}
		for i, change := range mutate {
			hub := hubConfig()
			change(hub)
			msg := feature("policycontroller", &GcpGkeFleetFeatureSpec{
				FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Policycontroller: &GcpGkeFleetFeaturePolicyController{PolicyControllerHubConfig: hub}},
			})
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), i)
		}
		msg := feature("policycontroller", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Policycontroller: &GcpGkeFleetFeaturePolicyController{}},
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a mesh management mode, an upstream fleet, and well-formed upgrade conditions", func() {
		gomega.Expect(validator.Validate(feature("servicemesh", &GcpGkeFleetFeatureSpec{
			FleetDefaultMemberConfig: &GcpGkeFleetFeatureFleetDefaultMemberConfig{Mesh: &GcpGkeFleetFeatureMesh{}},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(feature("clusterupgrade", &GcpGkeFleetFeatureSpec{
			Clusterupgrade: &GcpGkeFleetFeatureClusterUpgrade{},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(feature("clusterupgrade", &GcpGkeFleetFeatureSpec{
			Clusterupgrade: &GcpGkeFleetFeatureClusterUpgrade{UpstreamFleets: []*foreignkeyv1.StringValueOrRef{literal("dev-host")}, PostConditions: &GcpGkeFleetFeatureUpgradePostConditions{Soaking: "7d"}},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(feature("clusterupgrade", &GcpGkeFleetFeatureSpec{
			Clusterupgrade: &GcpGkeFleetFeatureClusterUpgrade{
				UpstreamFleets:      []*foreignkeyv1.StringValueOrRef{literal("dev-host")},
				GkeUpgradeOverrides: []*GcpGkeFleetFeatureGkeUpgradeOverride{{Upgrade: &GcpGkeFleetFeatureGkeUpgrade{Name: "k8s_node"}}},
			},
		}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(feature("fleetobservability", &GcpGkeFleetFeatureSpec{
			Fleetobservability: &GcpGkeFleetFeatureFleetObservability{LoggingConfig: &GcpGkeFleetFeatureFleetLoggingConfig{DefaultConfig: &GcpGkeFleetFeatureLogRoutingConfig{Mode: "MIRROR"}}},
		}))).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a well-formed feature name and deletion policy", func() {
		for _, name := range []string{"", "ConfigManagement", "config-management", "1feature"} {
			gomega.Expect(validator.Validate(feature(name, &GcpGkeFleetFeatureSpec{}))).ToNot(gomega.Succeed(), name)
		}
		gomega.Expect(validator.Validate(feature("configmanagement", &GcpGkeFleetFeatureSpec{DeletionPolicy: "KEEP"}))).ToNot(gomega.Succeed())
		gomega.Expect(validator.Validate(feature("configmanagement", &GcpGkeFleetFeatureSpec{Location: "Global"}))).ToNot(gomega.Succeed())
	})
})
