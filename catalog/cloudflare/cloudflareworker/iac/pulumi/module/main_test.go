package module

import (
	"encoding/base64"
	"strings"
	"sync"
	"testing"

	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareworkerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworker/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The program's R2 bundle path, run against mocks. It pins what the tofu module's
// offline guard pins for the other engine: the storage provider signs with the
// connection's R2 pair, the object is read as raw bytes whatever its Content-Type
// (an application/javascript bundle would otherwise read back empty), those bytes
// become the script content, and an empty bundle stops the program by name.

const (
	testAccountID    = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
	testR2AccessKey  = "r2-access-key-id-0123456789"
	testR2SecretKey  = "r2-secret-access-key-0123456789abcdef"
	testBundleBucket = "worker-builds"
	testBundleKey    = "api/index.js"
	testBundleSource = "export default { async fetch() { return new Response(\"from-bundle\"); } };"
)

// bundleMocks serves the bundle read and records what the program registered.
type bundleMocks struct {
	bundle []byte

	mu                sync.Mutex
	getObjectArgs     resource.PropertyMap
	r2ProviderInputs  resource.PropertyMap
	workerScriptInput resource.PropertyMap
}

func (m *bundleMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch args.TypeToken {
	case "pulumi:providers:aws":
		m.r2ProviderInputs = args.Inputs.Copy()
	case "cloudflare:index/workersScript:WorkersScript":
		m.workerScriptInput = args.Inputs.Copy()
	}
	return args.Name + "-id", args.Inputs.Copy(), nil
}

func (m *bundleMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if args.Token != "aws:s3/getObject:getObject" {
		return args.Args, nil
	}
	m.mu.Lock()
	m.getObjectArgs = args.Args.Copy()
	m.mu.Unlock()
	result := args.Args.Copy()
	result["bodyBase64"] = resource.NewStringProperty(base64.StdEncoding.EncodeToString(m.bundle))
	// The plain body stays empty, as the provider leaves it for a non-"readable" type.
	result["body"] = resource.NewStringProperty("")
	result["contentType"] = resource.NewStringProperty("application/javascript")
	return result, nil
}

func bundleStackInput() *cloudflareworkerv1alpha1.CloudflareWorkerStackInput {
	return &cloudflareworkerv1alpha1.CloudflareWorkerStackInput{
		Target: &cloudflareworkerv1alpha1.CloudflareWorker{
			Metadata: &shared.CloudResourceMetadata{Name: "bundled-api"},
			Spec: &cloudflareworkerv1alpha1.CloudflareWorkerSpec{
				AccountId:  testAccountID,
				WorkerName: "bundled-api",
				Source: &cloudflareworkerv1alpha1.CloudflareWorkerSpec_R2Bundle{
					R2Bundle: &cloudflareworkerv1alpha1.CloudflareWorkerScriptBundle{
						Bucket: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: testBundleBucket},
						},
						Path: testBundleKey,
					},
				},
			},
		},
		ProviderConfig: &cloudflareprovider.CloudflareProviderConfig{
			AuthScheme: cloudflareprovider.CloudflareAuthScheme_api_token,
			ApiToken:   "cf-test-token-0123456789abcdef",
			R2: &cloudflareprovider.CloudflareCredentialsR2Spec{
				AccessKeyId:     testR2AccessKey,
				SecretAccessKey: testR2SecretKey,
			},
		},
	}
}

// stringInput reads a string property, unwrapping a secret (the aws provider marks
// its keys secret).
func stringInput(t *testing.T, props resource.PropertyMap, key resource.PropertyKey) string {
	t.Helper()
	v, ok := props[key]
	if !ok {
		t.Fatalf("input %q was not registered", key)
	}
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsString() {
		t.Fatalf("input %q is not a string: %v", key, v)
	}
	return v.StringValue()
}

func TestResources_R2Bundle_DeploysTheBundleBytesWithTheConnectionsR2Pair(t *testing.T) {
	mocks := &bundleMocks{bundle: []byte(testBundleSource)}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return Resources(ctx, bundleStackInput())
	}, pulumi.WithMocks("cloudflare-worker", "test", mocks))
	if err != nil {
		t.Fatalf("the program must run to completion: %v", err)
	}

	mocks.mu.Lock()
	defer mocks.mu.Unlock()

	if mocks.getObjectArgs == nil {
		t.Fatal("the program never read the bundle")
	}
	if got := stringInput(t, mocks.getObjectArgs, "downloadBody"); got != "true" {
		t.Errorf("the bundle read must request the raw bytes (downloadBody=true), got %q", got)
	}
	if got := stringInput(t, mocks.getObjectArgs, "bucket"); got != testBundleBucket {
		t.Errorf("bundle bucket = %q, want %q", got, testBundleBucket)
	}
	if got := stringInput(t, mocks.getObjectArgs, "key"); got != testBundleKey {
		t.Errorf("bundle key = %q, want %q", got, testBundleKey)
	}

	if mocks.r2ProviderInputs == nil {
		t.Fatal("the program never built the R2 storage provider")
	}
	if got := stringInput(t, mocks.r2ProviderInputs, "accessKey"); got != testR2AccessKey {
		t.Errorf("R2 provider access key = %q, want the connection's %q", got, testR2AccessKey)
	}
	if got := stringInput(t, mocks.r2ProviderInputs, "secretKey"); got != testR2SecretKey {
		t.Error("R2 provider secret key is not the connection's")
	}

	if mocks.workerScriptInput == nil {
		t.Fatal("the program never registered the Worker script")
	}
	if got := stringInput(t, mocks.workerScriptInput, "content"); got != testBundleSource {
		t.Errorf("script content = %q, want the bundle's bytes %q", got, testBundleSource)
	}
}

func TestResources_R2Bundle_EmptyBundleStopsTheProgramByName(t *testing.T) {
	mocks := &bundleMocks{bundle: nil}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return Resources(ctx, bundleStackInput())
	}, pulumi.WithMocks("cloudflare-worker", "test", mocks))
	if err == nil {
		t.Fatal("an empty bundle must stop the program, never deploy a Worker with no code")
	}
	want := "r2://" + testBundleBucket + "/" + testBundleKey + " is empty"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error must name the empty object (%q), got: %v", want, err)
	}
}
