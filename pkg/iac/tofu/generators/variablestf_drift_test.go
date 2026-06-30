package generators

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"google.golang.org/protobuf/proto"
)

// migratedKinds is the allowlist of cloud-resource kinds whose committed
// variables.tf is owned by the generator (ProtoToVariablesTF) and guarded
// against drift. A kind is added here only after its module has been regenerated
// and validated (tofu validate against a null-pruned tfvars), so the guard can
// never be red for an unmigrated module. Remaining providers/kinds are migrated
// in tracked batches, each appended here.
//
// Scope note: only providers whose module conventions match the generator (it
// flattens wrapper types like StringValueOrRef to primitives and emits the
// canonical metadata block) belong here. AWS modules follow these conventions.
// Providers that intentionally diverge (e.g. OCI modules expose the wrapper
// object) are out of scope until their modules are migrated to the generator.
var migratedKinds = []string{
	// aws-ecs-environment chart kinds (the set that surfaced the schema bug).
	"AwsRoute53Zone",
	"AwsEcsCluster",
	"AwsIamRole",
	"AwsEcrRepo",
	"AwsSecurityGroup",
	"AwsAlb",
	"AwsCertManagerCert",
	"AwsEcsService",
	// AWS networking primitives already on the modern schema, brought under the
	// guard so they cannot regress.
	"AwsVpc",
	"AwsSubnet",
	"AwsInternetGateway",
	"AwsNatGateway",
	"AwsElasticIp",
}

// TestVariablesTFDrift asserts that every migrated module's committed
// variables.tf is byte-identical to the generator output. This makes the
// generator the single source of truth: a hand-edit or a legacy schema can never
// silently ship. Run with PLANTON_REGEN_VARIABLES=1 to (re)write the files from
// the generator instead of comparing.
func TestVariablesTFDrift(t *testing.T) {
	root := repoRoot(t)
	regenerate := os.Getenv("PLANTON_REGEN_VARIABLES") == "1"

	for _, kindName := range migratedKinds {
		kindName := kindName
		t.Run(kindName, func(t *testing.T) {
			kind := crkreflect.KindFromString(kindName)
			msg, err := crkreflect.NewInstance(kind)
			if err != nil {
				t.Fatalf("NewInstance(%s): %v", kindName, err)
			}

			want, err := ProtoToVariablesTF(msg)
			if err != nil {
				t.Fatalf("ProtoToVariablesTF(%s): %v", kindName, err)
			}
			want = strings.TrimRight(want, "\n") + "\n"

			path := moduleVariablesPath(root, msg)

			if regenerate {
				if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
					t.Fatalf("write %s: %v", path, err)
				}
				t.Logf("regenerated %s", path)
				return
			}

			gotBytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s (did you run PLANTON_REGEN_VARIABLES=1?): %v", path, err)
			}
			if strings.TrimRight(string(gotBytes), "\n")+"\n" != want {
				t.Errorf("variables.tf for %s is out of sync with the generator.\n"+
					"Run: PLANTON_REGEN_VARIABLES=1 go test ./pkg/iac/tofu/generators/ -run TestVariablesTFDrift\n"+
					"path: %s", kindName, path)
			}
		})
	}
}

// moduleVariablesPath derives a kind's module variables.tf path from its proto
// descriptor's source file: dev/planton/provider/<p>/<kind>/v1/api.proto ->
// <repo>/apis/dev/planton/provider/<p>/<kind>/v1/iac/tf/variables.tf.
func moduleVariablesPath(root string, msg proto.Message) string {
	protoPath := msg.ProtoReflect().Descriptor().ParentFile().Path()
	dir := filepath.Dir(protoPath)
	return filepath.Join(root, "apis", dir, "iac", "tf", "variables.tf")
}

// repoRoot walks up from this test file to the directory containing go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repo root (go.mod)")
		}
		dir = parent
	}
}
