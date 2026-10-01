package validatecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const secretReferenceChart = `apiVersion: infra-hub.planton.ai/v1alpha1
kind: InfraChart
metadata:
  name: Runner CA
spec:
  selector:
    kind: platform
  description: a binary secret whose data comes from the secret store
`

const secretReferenceDocument = `---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSecret
metadata:
  name: runner-ca
spec:
  name: runner-ca
  namespace:
    value: builds
  opaque:
    binaryData:
      ca.crt: $secret/runner-ca
`

// A host that resolves reference tokens before anything deploys passes its
// grammar through the command, so a schema rule written about the literal --
// here the binary data's base64 pattern -- waits for the token's value
// instead of failing the chart. Without the host's grammar the command judges
// the token as written and refuses the chart.
func TestChartValidateDefersARuleOnAHostToken(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Chart.yaml"), secretReferenceChart)
	writeFile(t, filepath.Join(dir, "values.yaml"), "params: []\n")
	writeFile(t, filepath.Join(dir, "templates", "secret.yaml"), secretReferenceDocument)
	isToken := func(value string) bool { return strings.HasPrefix(value, "$secret/") }

	if err := runChartValidate(Options{}, dir); err == nil {
		t.Fatal("without a host's token grammar the command must refuse the token as a literal")
	}
	if err := runChartValidate(Options{IsDeferredToken: isToken}, dir); err != nil {
		t.Fatalf("a rule on a host token must wait for its value, got: %v", err)
	}
}

func runChartValidate(opts Options, dir string) error {
	cmd := NewChartValidateCommand(opts)
	cmd.SetArgs([]string{dir})
	return cmd.Execute()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
