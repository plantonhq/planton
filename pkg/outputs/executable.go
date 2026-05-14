package outputs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

const executableTimeout = 30 * time.Second

// executableInput is the JSON payload sent to a transform-outputs executable
// on stdin.
type executableInput struct {
	Kind    string                 `json:"kind"`
	Outputs map[string]interface{} `json:"outputs"`
}

// runTransformExecutable invokes the transform-outputs executable in moduleDir,
// passing the kind and raw outputs on stdin as JSON, and parsing the flat
// map[string]string response from stdout.
//
// The process is killed after 30 seconds. A non-zero exit code is treated as
// an error; stderr is included in the error message for diagnostics.
func runTransformExecutable(
	moduleDir string,
	kind cloudresourcekind.CloudResourceKind,
	rawOutputs map[string]interface{},
) (map[string]string, error) {
	absModuleDir, err := filepath.Abs(moduleDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve absolute path for module dir %s", moduleDir)
	}
	execPath := filepath.Join(absModuleDir, executableFileName)

	input := executableInput{
		Kind:    kind.String(),
		Outputs: rawOutputs,
	}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal executable input")
	}

	ctx, cancel := context.WithTimeout(context.Background(), executableTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, execPath)
	cmd.Dir = absModuleDir
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.Errorf(
				"transform-outputs executable timed out after %s (stderr: %s)",
				executableTimeout, stderrStr)
		}
		return nil, errors.Wrapf(err,
			"transform-outputs executable failed (stderr: %s)", stderrStr)
	}

	outBytes := stdout.Bytes()
	if len(outBytes) == 0 {
		return nil, errors.New("transform-outputs executable produced empty output")
	}

	var result map[string]string
	if err := json.Unmarshal(outBytes, &result); err != nil {
		return nil, errors.Wrapf(err,
			"transform-outputs executable produced invalid JSON: %s",
			truncate(string(outBytes), 200))
	}

	return result, nil
}

// truncate returns at most maxLen bytes of s, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return fmt.Sprintf("%s...", s[:maxLen])
}
