package protobufyaml

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/fileutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"os"
	"regexp"
	"sigs.k8s.io/yaml"
	"strings"
)

// Load reads the contents of the input file and load it into an object
// the bytes from the input files are first converted to json
// the converted json is then converted into the protobuf message using protojson package.
func Load(inputFile string, obj proto.Message) error {
	isExists, err := fileutil.IsExists(inputFile)
	if err != nil {
		return errors.Wrapf(err, "failed to check file: %s", inputFile)
	}
	if !isExists {
		return errors.New("file does not exist")
	}
	inputFileBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return errors.Wrap(err, "failed to read input file")
	}
	jsonBytes, err := yaml.YAMLToJSON(inputFileBytes)
	if err != nil {
		return errors.Wrap(err, "failed to load yaml to json")
	}
	if err := protojson.Unmarshal(jsonBytes, obj); err != nil {
		return errors.New(TrimProtoPrefix(err.Error()))
	}
	return nil
}

func LoadYamlBytes(yamlBytes []byte, obj proto.Message) error {
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return errors.Wrap(err, "failed to load yaml to json")
	}
	if err := protojson.Unmarshal(jsonBytes, obj); err != nil {
		return errors.New(TrimProtoPrefix(err.Error()))
	}
	return nil
}

// TrimProtoPrefix removes the "proto: (line x:y): " prefix if present.
func TrimProtoPrefix(msg string) string {
	const marker = "proto:"
	if !strings.HasPrefix(msg, marker) {
		return msg // already clean
	}
	// proto: (line 12:34): <-- tolerate extra spaces
	re := regexp.MustCompile(`^proto:\s*\(line \d+:\d+\):\s*`)
	return re.ReplaceAllString(msg, "")
}
