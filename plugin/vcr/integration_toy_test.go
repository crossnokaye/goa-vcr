package vcr

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// goaVersion is the goa release the repository pins; the generator binary
// must match the module version or it fails to compile.
const goaVersion = "v3.23.4"

// TestToyExample_GenerateAndRun runs the whole pipeline on examples/toy:
// `goa gen` with the plugin linked in, then the toy module's own tests, which
// compile and exercise the generated VCR glue.
//
// `go test -short` skips it.
func TestToyExample_GenerateAndRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the toy pipeline in short mode")
	}

	toy := filepath.Join(mustRepoRoot(t), "examples", "toy")

	// Generate code into the toy module using the standard Goa tool.
	run(t, toy, "go", "run", "goa.design/goa/v3/cmd/goa@"+goaVersion, "gen",
		"github.com/xeger/goa-vcr/examples/toy/design", "-o", ".")

	for _, rel := range []string{
		"gen/toy/service.go",
		"gen/http/toy/server/server.go",
		"gen/http/toy/vcr/vcr.go",
	} {
		if _, err := os.Stat(filepath.Join(toy, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected generated file %s: %v", rel, err)
		}
	}
	if t.Failed() {
		t.FailNow()
	}

	// Compile + run the generated code and the toy's smoke tests.
	run(t, toy, "go", "test", "./...")
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	root := filepath.Clean(filepath.Join(dir, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("expected go.mod at repo root %q: %v", root, err)
	}
	return root
}

func run(t *testing.T, dir string, exe string, args ...string) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOFLAGS=-mod=mod",
		"GOWORK=off",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %s\n%s\n%v", exe, strings.Join(args, " "), string(out), err)
	}
}
