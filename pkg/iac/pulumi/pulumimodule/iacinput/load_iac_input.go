package iacinput

import (
	"fmt"
	"os"

	"buf.build/go/protovalidate"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput/fieldsextractor"
	"github.com/plantonhq/planton/pkg/protobufyaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	PulumiConfigKey   = "planton:iac-input"
	FilePathEnvVar    = "IAC_INPUT_YAML_FILE"
	YamlContentEnvVar = "IAC_INPUT_YAML"

	// OperationEnvVar tells the module program which engine operation spawned
	// it. `pulumi destroy --run-program` re-runs the program only so delete
	// hooks can fire; every resource in state is deleted whatever the program
	// registers. A module step that can FAIL for reasons unrelated to what is
	// being deleted (deriving CRDs from a chart version that was never
	// published, a repository that is unreachable) must therefore not fail
	// the program during a destroy, or the user could never tear down a stack
	// whose manifest is currently wrong. Set by every Pulumi spawner in this
	// repository; absent means an update or preview.
	OperationEnvVar  = "PLANTON_IAC_OPERATION"
	OperationDestroy = "destroy"
)

// IsDestroy reports whether the program runs for a `pulumi destroy`.
func IsDestroy() bool {
	return os.Getenv(OperationEnvVar) == OperationDestroy
}

func LoadIacInput(ctx *pulumi.Context, iacInput proto.Message) error {
	iacInputString, ok := ctx.GetConfig(PulumiConfigKey)
	var jsonBytes, iacInputYamlBytes []byte
	var err error

	if !ok {
		yamlContent := os.Getenv(YamlContentEnvVar)
		if yamlContent != "" {
			iacInputYamlBytes = []byte(yamlContent)
		} else {
			iacInputFilePath := os.Getenv(FilePathEnvVar)
			if iacInputFilePath == "" {
				return errors.Errorf("iac-input not found in pulumi config %s or in %s environment variable",
					PulumiConfigKey, FilePathEnvVar)
			}
			iacInputYamlBytes, err = os.ReadFile(iacInputFilePath)
			if err != nil {
				return errors.Wrap(err, "failed to read input file")
			}
		}

		jsonBytes, err = protobufyaml.YAMLToJSON(iacInputYamlBytes)
		if err != nil {
			return errors.Wrap(err, "failed to load yaml to json")
		}
	} else {
		jsonBytes, err = protobufyaml.YAMLToJSON([]byte(iacInputString))
		if err != nil {
			return errors.Wrap(err, "failed to load yaml to json")
		}
	}

	if err := protojson.Unmarshal(jsonBytes, iacInput); err != nil {
		return errors.Wrap(err, "failed to load json into proto message")
	}

	targetSpec, err := fieldsextractor.ExtractApiResourceSpecField(iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to extract api resource spec field")
	}

	v, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages((*targetSpec).Interface()),
	)
	if err != nil {
		fmt.Println("failed to initialize validator:", err)
	}

	if err = v.Validate((*targetSpec).Interface()); err != nil {
		return errors.Errorf("%s", err)
	}
	return nil
}
