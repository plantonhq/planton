package runner

import (
	"os"
	"path/filepath"
	"testing"

	awsiamrolev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
	awsnlbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsnlb/v1alpha1"
	awssubnetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssubnet/v1alpha1"
	gcpbackendservicev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbackendservice/v1alpha1"
	gcptargethttpsproxyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptargethttpsproxy/v1alpha1"
	gcpurlmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpurlmap/v1alpha1"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/shared/catalogkind"
)

const subnetManifestWithRef = `apiVersion: aws.planton.dev/v1alpha1
kind: AwsSubnet
metadata:
  name: ref-subnet
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
  availabilityZone: us-west-2a
  cidrBlock: 10.0.1.0/24
`

func writeTempManifest(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatalf("failed to write temp manifest: %v", err)
	}
	return path
}

// singleInstance wraps one instance's outputs in the kind+name keyed shape,
// for tests where only one prerequisite of the kind is deployed.
func singleInstance(kind catalogkind.CatalogKind, name string, outputs map[string]interface{}) DependencyOutputs {
	return DependencyOutputs{kind: {name: outputs}}
}

func TestResolveManifestRefs_ResolvesVpcIdFromPrerequisite(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsVpc, "my-vpc", map[string]interface{}{
		"vpc_id":   "vpc-resolved123",
		"vpc_cidr": "10.0.0.0/16",
	})

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath == manifestPath {
		t.Fatal("expected a new resolved manifest path, got the original")
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	subnet, ok := obj.(*awssubnetv1alpha1.AwsSubnet)
	if !ok {
		t.Fatalf("resolved manifest is not an AwsSubnet: %T", obj)
	}
	if got := subnet.GetSpec().GetVpcId().GetValue(); got != "vpc-resolved123" {
		t.Errorf("vpc_id value = %q, want %q", got, "vpc-resolved123")
	}
	if subnet.GetSpec().GetVpcId().GetValueFrom() != nil {
		t.Error("vpc_id should be a literal after resolution, but value_from is still set")
	}
}

func TestResolveManifestRefs_KeepsScenarioBasename(t *testing.T) {
	// Verifier dispatch keys behavioral variants off the scenario name in
	// the manifest path — the resolved copy must preserve the basename
	// (a randomly named copy would silently demote reference-carrying
	// behavioral scenarios to their plain verifiers).
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "behavioral-node-launch.yaml")
	if err := os.WriteFile(manifestPath, []byte(subnetManifestWithRef), 0600); err != nil {
		t.Fatalf("failed to write temp manifest: %v", err)
	}

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsVpc, "my-vpc", map[string]interface{}{
		"vpc_id": "vpc-resolved123",
	})

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if filepath.Base(resolvedPath) != "behavioral-node-launch.yaml" {
		t.Errorf("resolved copy basename = %q, want %q", filepath.Base(resolvedPath), "behavioral-node-launch.yaml")
	}
}

// The sole-instance fallback: a reference whose name matches no deployed
// instance still resolves when exactly one instance of the kind exists, so
// scenario manifests are not coupled to the install profile's fixed names.
func TestResolveManifestRefs_SoleInstanceResolvesDespiteNameMismatch(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsVpc, "some-other-name", map[string]interface{}{
		"vpc_id": "vpc-sole456",
	})

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	subnet := obj.(*awssubnetv1alpha1.AwsSubnet)
	if got := subnet.GetSpec().GetVpcId().GetValue(); got != "vpc-sole456" {
		t.Errorf("vpc_id value = %q, want the sole instance's %q", got, "vpc-sole456")
	}
}

// Multi-instance selection: with two prerequisites of the same kind deployed,
// each reference resolves against the instance its name addresses -- the
// mechanism that lets a load balancer reference two different-AZ subnets.
const nlbManifestWithTwoSubnetRefs = `apiVersion: aws.planton.dev/v1alpha1
kind: AwsNlb
metadata:
  name: ref-nlb
spec:
  region: us-west-2
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: subnet-az-a
          fieldPath: status.outputs.subnet_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: subnet-az-b
          fieldPath: status.outputs.subnet_id
`

func TestResolveManifestRefs_MultiInstanceResolvesByName(t *testing.T) {
	manifestPath := writeTempManifest(t, nlbManifestWithTwoSubnetRefs)

	depOutputs := DependencyOutputs{
		catalogkind.CatalogKind_AwsSubnet: {
			"subnet-az-a": {"subnet_id": "subnet-aaa"},
			"subnet-az-b": {"subnet_id": "subnet-bbb"},
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	nlb, ok := obj.(*awsnlbv1alpha1.AwsNlb)
	if !ok {
		t.Fatalf("resolved manifest is not an AwsNlb: %T", obj)
	}
	mappings := nlb.GetSpec().GetSubnetMappings()
	if len(mappings) != 2 {
		t.Fatalf("subnet_mappings length = %d, want 2", len(mappings))
	}
	if got := mappings[0].GetSubnetId().GetValue(); got != "subnet-aaa" {
		t.Errorf("first mapping = %q, want subnet-aaa (the instance named subnet-az-a)", got)
	}
	if got := mappings[1].GetSubnetId().GetValue(); got != "subnet-bbb" {
		t.Errorf("second mapping = %q, want subnet-bbb (the instance named subnet-az-b)", got)
	}
}

// The nested-refs test above also proves recursion: subnet_id refs live inside
// the repeated AwsNlbSubnetMapping message, not directly on the spec. This case
// guards the ambiguity error instead: several instances deployed and a name
// matching none of them must fail loudly, never pick one arbitrarily.
func TestResolveManifestRefs_AmbiguousNameErrors(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef) // references AwsVpc name "my-vpc"

	depOutputs := DependencyOutputs{
		catalogkind.CatalogKind_AwsVpc: {
			"vpc-one": {"vpc_id": "vpc-111"},
			"vpc-two": {"vpc_id": "vpc-222"},
		},
	}

	if _, err := ResolveManifestRefs(manifestPath, depOutputs); err == nil {
		t.Fatal("expected an ambiguity error when the name matches none of several instances, got nil")
	}
}

const roleManifestWithRepeatedRefs = `apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamRole
metadata:
  name: ref-role
spec:
  region: us-west-2
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: ec2.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
    - valueFrom:
        kind: AwsIamPolicy
        name: my-policy
        fieldPath: status.outputs.policy_arn
`

// The repeated case: a list mixing a literal (an AWS-managed policy ARN) with a
// reference to a deployed prerequisite. Each element resolves independently and
// the literal passes through untouched.
func TestResolveManifestRefs_ResolvesRepeatedRefsFromPrerequisite(t *testing.T) {
	manifestPath := writeTempManifest(t, roleManifestWithRepeatedRefs)

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsIamPolicy, "my-policy", map[string]interface{}{
		"policy_arn":  "arn:aws:iam::123456789012:policy/my-policy",
		"policy_name": "my-policy",
	})

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath == manifestPath {
		t.Fatal("expected a new resolved manifest path, got the original")
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	role, ok := obj.(*awsiamrolev1alpha1.AwsIamRole)
	if !ok {
		t.Fatalf("resolved manifest is not an AwsIamRole: %T", obj)
	}
	arns := role.GetSpec().GetManagedPolicyArns()
	if len(arns) != 2 {
		t.Fatalf("managed_policy_arns length = %d, want 2", len(arns))
	}
	if got := arns[0].GetValue(); got != "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore" {
		t.Errorf("literal element = %q, want the untouched AWS-managed ARN", got)
	}
	if got := arns[1].GetValue(); got != "arn:aws:iam::123456789012:policy/my-policy" {
		t.Errorf("resolved element = %q, want the prerequisite's policy_arn", got)
	}
	if arns[1].GetValueFrom() != nil {
		t.Error("resolved element should be a literal, but value_from is still set")
	}
}

// A repeated ref whose kind has no deployed prerequisite is left untouched
// rather than erroring -- matching the singular behavior.
func TestResolveManifestRefs_RepeatedRefWithoutPrerequisiteLeftUntouched(t *testing.T) {
	manifestPath := writeTempManifest(t, roleManifestWithRepeatedRefs)

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsVpc, "unrelated-vpc", map[string]interface{}{
		"vpc_id": "vpc-unrelated",
	})

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath != manifestPath {
		t.Errorf("expected original path when no matching prerequisite is deployed, got %q", resolvedPath)
	}
}

func TestResolveManifestRefs_NoDependenciesReturnsOriginal(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	resolvedPath, err := ResolveManifestRefs(manifestPath, nil)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}
	if resolvedPath != manifestPath {
		t.Errorf("expected original path when there are no dependencies, got %q", resolvedPath)
	}
}

func TestResolveManifestRefs_MissingOutputErrors(t *testing.T) {
	manifestPath := writeTempManifest(t, subnetManifestWithRef)

	depOutputs := singleInstance(catalogkind.CatalogKind_AwsVpc, "my-vpc", map[string]interface{}{
		"vpc_cidr": "10.0.0.0/16", // vpc_id intentionally absent
	})

	if _, err := ResolveManifestRefs(manifestPath, depOutputs); err == nil {
		t.Fatal("expected an error when the prerequisite output is missing, got nil")
	}
}

const backendServiceNegManifest = `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBackendService
metadata:
  name: neg-backend
spec:
  protocol: HTTP
  loadBalancingScheme: EXTERNAL_MANAGED
  backends:
    - group:
        valueFrom:
          kind: GcpRegionNetworkEndpointGroup
          name: my-neg
          fieldPath: status.outputs.self_link
      balancingMode: RATE
      maxRate: 100
`

const urlMapExplicitKindManifest = `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpUrlMap
metadata:
  name: my-url-map
spec:
  defaultService:
    valueFrom:
      kind: GcpBackendService
      name: my-backend
      fieldPath: status.outputs.self_link
`

func TestResolveManifestRefs_ResolvesNestedBackendGroupByExplicitKind(t *testing.T) {
	manifestPath := writeTempManifest(t, backendServiceNegManifest)

	depOutputs := DependencyOutputs{
		catalogkind.CatalogKind_GcpRegionNetworkEndpointGroup: {
			"my-neg": {
				"self_link": "https://www.googleapis.com/compute/v1/projects/p/regions/r/networkEndpointGroups/n",
			},
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	bs, ok := obj.(*gcpbackendservicev1alpha1.GcpBackendService)
	if !ok {
		t.Fatalf("resolved manifest is not a GcpBackendService: %T", obj)
	}
	if len(bs.GetSpec().GetBackends()) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(bs.GetSpec().GetBackends()))
	}
	got := bs.GetSpec().GetBackends()[0].GetGroup().GetValue()
	want := "https://www.googleapis.com/compute/v1/projects/p/regions/r/networkEndpointGroups/n"
	if got != want {
		t.Errorf("backends[0].group = %q, want %q", got, want)
	}
}

const httpsProxyRepeatedRefManifest = `apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpsProxy
metadata:
  name: my-https-proxy
spec:
  urlMap:
    valueFrom:
      kind: GcpUrlMap
      name: my-url-map
      fieldPath: status.outputs.self_link
  sslCertificates:
    - valueFrom:
        kind: GcpManagedSslCertificate
        name: my-cert
        fieldPath: status.outputs.self_link
`

// A REPEATED StringValueOrRef field (e.g. a target HTTPS proxy's
// ssl_certificates) must resolve each element in place — the field descriptor
// is a list, so it must never be read as a singular message.
func TestResolveManifestRefs_ResolvesRepeatedRefElements(t *testing.T) {
	manifestPath := writeTempManifest(t, httpsProxyRepeatedRefManifest)

	depOutputs := DependencyOutputs{
		catalogkind.CatalogKind_GcpUrlMap: {
			"my-url-map": {
				"self_link": "https://www.googleapis.com/compute/v1/projects/p/global/urlMaps/um",
			},
		},
		catalogkind.CatalogKind_GcpManagedSslCertificate: {
			"my-cert": {
				"self_link": "https://www.googleapis.com/compute/v1/projects/p/global/sslCertificates/cert",
			},
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	proxy, ok := obj.(*gcptargethttpsproxyv1alpha1.GcpTargetHttpsProxy)
	if !ok {
		t.Fatalf("resolved manifest is not a GcpTargetHttpsProxy: %T", obj)
	}
	if got, want := proxy.GetSpec().GetUrlMap().GetValue(), "https://www.googleapis.com/compute/v1/projects/p/global/urlMaps/um"; got != want {
		t.Errorf("url_map = %q, want %q", got, want)
	}
	if len(proxy.GetSpec().GetSslCertificates()) != 1 {
		t.Fatalf("expected 1 ssl certificate, got %d", len(proxy.GetSpec().GetSslCertificates()))
	}
	got := proxy.GetSpec().GetSslCertificates()[0].GetValue()
	want := "https://www.googleapis.com/compute/v1/projects/p/global/sslCertificates/cert"
	if got != want {
		t.Errorf("ssl_certificates[0] = %q, want %q", got, want)
	}
	if proxy.GetSpec().GetSslCertificates()[0].GetValueFrom() != nil {
		t.Error("ssl_certificates[0] should be a literal after resolution, but value_from is still set")
	}
}

func TestResolveManifestRefs_ResolvesExplicitValueFromKindWithoutDefaultKind(t *testing.T) {
	manifestPath := writeTempManifest(t, urlMapExplicitKindManifest)

	depOutputs := DependencyOutputs{
		catalogkind.CatalogKind_GcpBackendService: {
			"my-backend": {
				"self_link": "https://www.googleapis.com/compute/v1/projects/p/global/backendServices/bs",
			},
		},
	}

	resolvedPath, err := ResolveManifestRefs(manifestPath, depOutputs)
	if err != nil {
		t.Fatalf("ResolveManifestRefs failed: %v", err)
	}

	obj, err := manifest.LoadManifest(resolvedPath)
	if err != nil {
		t.Fatalf("failed to load resolved manifest: %v", err)
	}
	um, ok := obj.(*gcpurlmapv1alpha1.GcpUrlMap)
	if !ok {
		t.Fatalf("resolved manifest is not a GcpUrlMap: %T", obj)
	}
	got := um.GetSpec().GetDefaultService().GetValue()
	want := "https://www.googleapis.com/compute/v1/projects/p/global/backendServices/bs"
	if got != want {
		t.Errorf("default_service = %q, want %q", got, want)
	}
}
