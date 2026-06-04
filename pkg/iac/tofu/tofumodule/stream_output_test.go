package tofumodule

import (
	"fmt"
	"os/exec"
	"runtime"
	"testing"
)

// TestStreamCommandJSONOutput_NoCloseRace drives the helper against a command
// that emits many stdout lines and then exits, asserting every line is delivered
// and no error is returned. Before the read-before-Wait fix this surfaced the
// sporadic "read |0: file already closed" race; run with -race -count=N to guard
// against regression.
func TestStreamCommandJSONOutput_NoCloseRace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test relies on a POSIX shell")
	}

	const lineCount = 5000

	collected := make([]string, 0, lineCount)
	done := make(chan struct{})
	lines := make(chan string, 100)
	go func() {
		for line := range lines {
			collected = append(collected, line)
		}
		close(done)
	}()

	cmd := exec.Command("sh", "-c", fmt.Sprintf("for i in $(seq 1 %d); do echo line-$i; done", lineCount))
	err := streamCommandJSONOutput("tofu", cmd, lines)
	close(lines)
	<-done

	if err != nil {
		t.Fatalf("streamCommandJSONOutput returned error: %v", err)
	}
	if len(collected) != lineCount {
		t.Fatalf("expected %d lines, got %d", lineCount, len(collected))
	}
	if collected[0] != "line-1" || collected[lineCount-1] != fmt.Sprintf("line-%d", lineCount) {
		t.Fatalf("unexpected line content: first=%q last=%q", collected[0], collected[lineCount-1])
	}
}

// TestStreamCommandJSONOutput_NonZeroExit verifies that a non-zero process exit
// is surfaced as an error even when stdout was consumed cleanly.
func TestStreamCommandJSONOutput_NonZeroExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test relies on a POSIX shell")
	}

	lines := make(chan string, 16)
	go func() {
		for range lines {
		}
	}()

	cmd := exec.Command("sh", "-c", "echo hello; exit 3")
	err := streamCommandJSONOutput("tofu", cmd, lines)
	close(lines)

	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}
